package network

import (
	"log"

	"github.com/plankton4/chat-app-server/pb"
	"google.golang.org/protobuf/proto"
)

func ProcessProtoMessage(client *Client, message []byte) {
	//log.Println("Process client message")

	pbMessage := &pb.PBMessage{}
	if err := proto.Unmarshal(message, pbMessage); err != nil {
		log.Println("Failed to unmarshal pbMessage:", err)
	}

	log.Println("ProcessProtoMessage. Client: ", client.clientID, " pbMessage ", pbMessage)

	if !client.isAuthenticated {
		_, ok := pbMessage.InternalMessage.(*pb.PBMessage_MessAuthReq)
		if !ok {
			HandleNotAuthenticatedMessage(client, pbMessage)
			return
		}
	}

	switch internalMess := pbMessage.InternalMessage.(type) {
	// SendChatMessageReq
	case *pb.PBMessage_MessSendChatMessageReq:
		log.Println("ProcessProtoMessage PBMessage_MessSendChatMessageReq case ", internalMess)
		HandleSendChatMessage(client, internalMess.MessSendChatMessageReq)

	// SendEditChatMessageReq
	case *pb.PBMessage_MessEditChatMessageReq:
		log.Println("ProcessProtoMessage PBMessage_MessSendEditChatMessageReq case ", internalMess)
		HandleEditChatMessage(client, internalMess.MessEditChatMessageReq)

	// DeleteChatMessageReq
	case *pb.PBMessage_MessDeleteChatMessageReq:
		log.Println("ProcessProtoMessage DeleteChatMessageReq case ", internalMess)
		HandleDeleteChatMessage(client, internalMess.MessDeleteChatMessageReq)

	// AuthReq
	case *pb.PBMessage_MessAuthReq:
		log.Println("ProcessProtoMessage PBMessage_MessAuthReq case ", internalMess)
		HandleAuthentication(client, pbMessage.RowID, internalMess.MessAuthReq)

	// GetUserDataReq
	case *pb.PBMessage_MessGetUserDataReq:
		log.Println("ProcessProtoMessage PBMessage_MessGetUserDataReq case ", internalMess)
		HandleGetUserData(client, pbMessage.RowID, internalMess.MessGetUserDataReq)

	// GetAllChatMessagesReq
	case *pb.PBMessage_MessGetAllChatMessagesReq:
		log.Println("ProcessProtoMessage PBMessage_MessGetAllChatMessagesReq case ", internalMess)
		HandleGetAllChatMessages(client, pbMessage.RowID, internalMess.MessGetAllChatMessagesReq)

	// GetChatListReq
	case *pb.PBMessage_MessGetChatListReq:
		log.Println("ProcessProtoMessage PBMessage_MessGetChatListReq case ", internalMess)
		HandleGetChatList(client, pbMessage.RowID, internalMess.MessGetChatListReq)

	// GetUnreadInfoReq
	case *pb.PBMessage_MessGetUnreadInfoReq:
		log.Println("ProcessProtoMessage PBMessage_MessGetUnreadInfoReq case ", internalMess)
		HandleGetUnreadInfo(client, pbMessage.RowID, internalMess.MessGetUnreadInfoReq)

	// SubscribeToPushReq
	case *pb.PBMessage_MessSubscribeToPushReq:
		log.Println("ProcessProtoMessage PBMessage_MessSubscribeToPushReq case ", internalMess)
		HandleSubscribeToPush(client, pbMessage.RowID, internalMess.MessSubscribeToPushReq)

	// default
	default:
		log.Println("ProcessProtoMessage default case")
		break
	}
}

func HandleNotAuthenticatedMessage(client *Client, mess *pb.PBMessage) {
	pbm := &pb.PBMessage{
		InternalMessage: &pb.PBMessage_MessReturnedMessageEvent{
			MessReturnedMessageEvent: &pb.ReturnedMessageEvent{
				ReasonOfReturn:  pb.ReturnReason_AuthenticationNeeded,
				ReturnedMessage: mess,
			},
		},
	}

	log.Println("PROTO HandleNotAuthenticatedMessage ", pbm)

	result, err := proto.Marshal(pbm)
	if err != nil {
		log.Println("Error when proto.Marshal!!!! ", err)
	}

	client.send <- result
}
