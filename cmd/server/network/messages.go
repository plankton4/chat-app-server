package network

import (
	"log"
	"strconv"
	"time"

	"firebase.google.com/go/v4/messaging"
	"github.com/golang/protobuf/proto"
	"github.com/plankton4/chat-app-server/cmd/server/database/mongodb"
	"github.com/plankton4/chat-app-server/cmd/server/fcm"
	"github.com/plankton4/chat-app-server/cmd/server/misc"
	"github.com/plankton4/chat-app-server/pb"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func HandleGetAllChatMessages(client *Client, rowID uint32, req *pb.GetAllChatMessagesReq) {
	//log.Println("Handle get all chat messages")
	var errString string

	chatMessages, err := mongodb.GetAllMessages(client.clientID, req.UserID, req.ChatID)

	log.Println("HandleGetAllChatMessages, found messages: ", len(chatMessages))

	if err != nil {
		errString = err.Error()
	}

	pbMess := &pb.PBMessage{
		RowID:  rowID,
		ErrStr: &errString,
		InternalMessage: &pb.PBMessage_MessGetAllChatMessagesAnswer{
			MessGetAllChatMessagesAnswer: &pb.GetAllChatMessagesAnswer{
				ChatID:   req.ChatID,
				UserID:   req.UserID,
				Messages: chatMessages,
			},
		},
	}

	result, err := proto.Marshal(pbMess)
	if err != nil {
		log.Println("Error in HandleGetAllChatMessages! ", err)
	}

	//log.Println("Try broadcast pbMess in get all messages ", pbMess)

	client.send <- result
}

func HandleSendChatMessage(client *Client, req *pb.SendChatMessageReq) {
	//log.Println("Handle send chat mess")

	fromUserID := req.FromUserID
	toUserID := req.ToUserID
	toChatID := req.ToChatID
	time := time.Now().Unix()

	chatMessageData := &pb.ChatMessageData{
		Type: req.Type,
		Time: func() *uint32 {
			convertedToU32 := uint32(time)
			return &convertedToU32
		}(),
		FromUserID:  fromUserID,
		ToChatID:    toChatID,
		ToUserID:    toUserID,
		Text:        req.Text,
		ImageURL:    req.ImageURL,
		AspectRatio: req.AspectRatio,
	}

	// нужно найти оригинал сообщения
	if req.RepliedMessage != nil {
		replyMessage, err := mongodb.GetMessageByID(
			req.RepliedMessage.MessageID,
			req.RepliedMessage.FromUserID,
			req.RepliedMessage.ToUserID,
			req.RepliedMessage.ToChatID,
		)

		if err != nil {
			log.Println("NOT FOUND REPLIED MESS")
			replyMessage.MessageID = "0"
			replyMessage.Type = pb.ChatMessageType_Text
			replyMessage.Text = func() *string {
				var s = "Not found"
				return &s
			}()
			replyMessage.FromUserID = 0
		}

		log.Println("REPLY MESS ", replyMessage)

		chatMessageData.RepliedMessage = replyMessage
	}

	insertedID := mongodb.AddChatMessage(client.clientID, toUserID, toChatID, chatMessageData)
	if insertedID != nil {
		chatMessageData.MessageID = insertedID.(primitive.ObjectID).Hex()
	}

	chatMess := &pb.PBMessage{
		InternalMessage: &pb.PBMessage_MessNewChatMessageEvent{
			MessNewChatMessageEvent: &pb.NewChatMessageEvent{
				ChatMessage: chatMessageData,
			},
		},
	}

	result, err := proto.Marshal(chatMess)
	if err != nil {
		log.Println("Error in HandleSendChatMessageReq! ", err)
	}

	log.Println("Try broadcast chatmess ", chatMess)

	if toChatID != nil {
		if hub, ok := client.hubs[HubID(*toChatID)]; ok {
			log.Println("Found hub ", toChatID)
			hub.broadcast <- result
		}
	}

	var getNotificationText = func() string {
		switch req.Type {
		case pb.ChatMessageType_Text:
			return *chatMessageData.Text
		case pb.ChatMessageType_Image, pb.ChatMessageType_GIF:
			return "Image"
		}

		return ""
	}

	//
	// Sending push notifications
	//
	if toChatID != nil && *toChatID != uint32(HubSpamChat) {
		subs, err := GetSubscribers(*toChatID)
		if err != nil {
			log.Println("Error while finding subs ", err.Error())
		} else {
			log.Println("Found subs ", subs)
		}

		subs = misc.RemoveFromSlice(subs, client.clientID)

		tokens, err := getTokens(subs)
		if err != nil {
			log.Println("Error when trying to get fcm tokens ", err.Error())
		}

		if tokens != nil {
			//log.Println("Found Tokens ", tokens)
			badge := 0
			sound := ""

			if *toChatID != uint32(HubSpamChat) {
				badge = 1
				sound = "default"
			}

			fcm.SendMulticast(
				fcm.FirebaseApp,
				tokens,
				messaging.Notification{
					Title: ChatsDataByID[HubID(*toChatID)].Title,
					Body:  getNotificationText(),
				},
				messaging.Aps{
					Badge: &badge,
					Sound: sound,
				},
				map[string]string{
					"sectionID": "1",
					"toChatID":  strconv.FormatUint(uint64(*toChatID), 10),
				},
			)
		}
	}
}

func HandleEditChatMessage(client *Client, req *pb.EditChatMessageReq) {
	log.Println("Handle edit chat mess")

	chatMessageData := req.OriginMessage
	chatMessageData.IsEdited = func() *bool {
		isEdited := true
		return &isEdited
	}()

	if req.NewText != nil {
		chatMessageData.Text = req.NewText
	}

	newFields := []string{
		mongodb.KeyText,
		mongodb.KeyIsEdited,
	}
	err := mongodb.EditChatMessage(chatMessageData, newFields)

	if err != nil {
		return
	}

	chatMess := &pb.PBMessage{
		InternalMessage: &pb.PBMessage_MessChatMessageChangedEvent{
			MessChatMessageChangedEvent: &pb.ChatMessageChangedEvent{
				MessageID: chatMessageData.MessageID,
				Type:      chatMessageData.Type,
				ToChatID:  chatMessageData.ToChatID,
				ToUserID:  chatMessageData.ToUserID,
				NewText:   req.NewText,
				IsEdited: func() *bool {
					isEdited := true
					return &isEdited
				}(),
			},
		},
	}

	result, err := proto.Marshal(chatMess)
	if err != nil {
		log.Println("Error in HandleEditChatMessage! ", err)
	}

	log.Println("Try broadcast edit chatmess ", chatMess)

	toChatID := req.OriginMessage.ToChatID
	if toChatID != nil {
		if hub, ok := client.hubs[HubID(*toChatID)]; ok {
			log.Println("Found hub ", toChatID)
			hub.broadcast <- result
		}
	}
}

func HandleDeleteChatMessage(client *Client, req *pb.DeleteChatMessageReq) {
	err := mongodb.DeleteChatMessage(req.MessageID, req.FromUserID, req.ToUserID, req.ToChatID)

	if err != nil {
		return
	}

	chatMess := &pb.PBMessage{
		InternalMessage: &pb.PBMessage_MessChatMessageDeletedEvent{
			MessChatMessageDeletedEvent: &pb.ChatMessageDeletedEvent{
				MessageID: req.MessageID,
				ToChatID:  req.ToChatID,
				ToUserID:  req.ToUserID,
			},
		},
	}

	result, err := proto.Marshal(chatMess)
	if err != nil {
		log.Println("Error in HandleDeleteChatMessage! ", err)
	}

	log.Println("Try broadcast delete chatmess ", chatMess)

	if req.ToChatID != nil {
		if hub, ok := client.hubs[HubID(*req.ToChatID)]; ok {
			log.Println("Found hub ", req.ToChatID)
			hub.broadcast <- result
		}
	}
}
