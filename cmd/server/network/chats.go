package network

import (
	"log"

	"github.com/golang/protobuf/proto"
	"github.com/plankton4/chat-app-server/cmd/server/database/mongodb"
	"github.com/plankton4/chat-app-server/pb"
)

var ChatsDataByID = map[HubID]*pb.ChatData{
	HubGeneralChat: {
		ChatID: uint32(HubGeneralChat),
		Title:  "General",
		IconURL: func() *string {
			url := "https://raw.githubusercontent.com/plankton4/files/main/kermit.jpeg"
			return &url
		}(),
	},
	HubMemesChat: {
		ChatID: uint32(HubMemesChat),
		Title:  "Memes",
		IconURL: func() *string {
			url := "https://raw.githubusercontent.com/plankton4/files/main/doge_laugh.jpeg"
			return &url
		}(),
	},
	HubMoviesChat: {
		ChatID: uint32(HubMoviesChat),
		Title:  "Movies",
		IconURL: func() *string {
			url := "https://raw.githubusercontent.com/plankton4/files/main/waltz.jpeg"
			return &url
		}(),
	},
	HubVideoGamesChat: {
		ChatID: uint32(HubVideoGamesChat),
		Title:  "Video Games",
		IconURL: func() *string {
			url := "https://raw.githubusercontent.com/plankton4/files/main/ellie_tlou2.jpg"
			return &url
		}(),
	},
}

func HandleGetChatList(client *Client, rowID uint32, req *pb.GetChatListReq) {
	var errString string

	var chats []*pb.ChatData

	for _, hubID := range testHubs {
		chats = append(chats, ChatsDataByID[hubID])
	}

	pbMess := &pb.PBMessage{
		RowID:  rowID,
		ErrStr: &errString,
		InternalMessage: &pb.PBMessage_MessGetChatListResp{
			MessGetChatListResp: &pb.GetChatListResp{
				Chats: chats,
			},
		},
	}

	result, err := proto.Marshal(pbMess)
	if err != nil {
		log.Println("Error in HandleGetChatList! ", err)
	}

	//log.Println("GetChatList RESP ", pbMess)

	client.send <- result
}

func HandleGetUnreadInfo(client *Client, rowID uint32, req *pb.GetUnreadInfoReq) {
	var errString string
	var unreadInfo []*pb.UnreadInfo

	defer func() {
		pbMess := &pb.PBMessage{
			RowID:  rowID,
			ErrStr: &errString,
			InternalMessage: &pb.PBMessage_MessGetUnreadInfoResp{
				MessGetUnreadInfoResp: &pb.GetUnreadInfoResp{
					Info: unreadInfo,
				},
			},
		}

		result, err := proto.Marshal(pbMess)
		if err != nil {
			log.Println("Error in HandleGetUnreadInfo! ", err)
		}

		client.send <- result
	}()

	var processUnreadInfo = func(roomIDs []uint32, isPrivate bool) {
		if roomIDs == nil {
			return
		}

		for _, ID := range roomIDs {
			var info *pb.UnreadInfo
			var err error
			var roomID = ID

			if isPrivate {
				info, err = mongodb.GetUnreadInfo(client.clientID, &roomID, nil)
			} else {
				info, err = mongodb.GetUnreadInfo(client.clientID, nil, &roomID)
			}

			if err != nil {
				log.Println("Error in HandleGetUnreadInfo!!!: ", err)
				if err == mongodb.NotFoundError {
					continue
				}

				errString = err.Error()
				return
			}

			if info != nil {
				unreadInfo = append(unreadInfo, info)
			}
		}
	}

	processUnreadInfo(req.UserIDs, true)
	processUnreadInfo(req.ChatIDs, false)
}
