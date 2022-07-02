package network

import (
	"log"

	"github.com/golang/protobuf/proto"
	"github.com/plankton4/chat-app-server/cmd/server/database/mongodb"
	"github.com/plankton4/chat-app-server/pb"
)

var ChatsDataByID = map[HubID]pb.ChatData{
	HubGeneralChat: {
		ChatID: uint32(HubGeneralChat),
		Title:  "Русский чат",
	},
	HubSpamChat: {
		ChatID: uint32(HubSpamChat),
		Title:  "СпамЧат",
	},
	HubCabinetSS: {
		ChatID: uint32(HubCabinetSS),
		Title:  "Кабинет 🤙",
	},
	Hub1001Task: {
		ChatID: uint32(Hub1001Task),
		Title:  "1001 задача 🍆",
	},
}

func HandleGetChatList(client *Client, rowID uint32, req *pb.GetChatListReq) {
	var errString string

	pbMess := &pb.PBMessage{
		RowID:  rowID,
		ErrStr: &errString,
		InternalMessage: &pb.PBMessage_MessGetChatListResp{
			MessGetChatListResp: &pb.GetChatListResp{
				Chats: []*pb.ChatData{
					{
						ChatID: uint32(HubGeneralChat),
						Title:  ChatsDataByID[HubGeneralChat].Title,
					},
					{
						ChatID: uint32(HubSpamChat),
						Title:  ChatsDataByID[HubSpamChat].Title,
					},
					{
						ChatID: uint32(HubCabinetSS),
						Title:  ChatsDataByID[HubCabinetSS].Title,
					},
					{
						ChatID: uint32(Hub1001Task),
						Title:  ChatsDataByID[Hub1001Task].Title,
					},
				},
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
