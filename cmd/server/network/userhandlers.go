package network

import (
	"log"

	"github.com/golang/protobuf/proto"
	"github.com/plankton4/chat-app-server/cmd/server/user"
	"github.com/plankton4/chat-app-server/pb"
)

func HandleSubscribeToPush(client *Client, rowID uint32, req *pb.SubscribeToPushReq) {
	var errString string

	err := user.SubscribeToPush(client.clientID, req.Token)

	if err != nil {
		log.Println("Error in SubscribeToPush ", err.Error())
		errString = err.Error()
	}
	log.Println("NAME ", req.ProtoReflect().Descriptor().Name())

	pbMess := &pb.PBMessage{
		RowID:           rowID,
		ErrStr:          &errString,
		InternalMessage: &pb.PBMessage_MessStandartAnswer{},
	}

	result, err := proto.Marshal(pbMess)
	if err != nil {
		log.Println("Error in HandleSubscribeToPush! ", err)
	}

	client.send <- result
}

func HandleGetUserData(client *Client, rowID uint32, req *pb.GetUserDataReq) {
	var errID uint32
	var errStr string
	var data []*pb.UserData

	defer func() {
		pbm := &pb.PBMessage{
			RowID:  rowID,
			ErrID:  &errID,
			ErrStr: &errStr,
			InternalMessage: &pb.PBMessage_MessGetUserDataAnswer{
				MessGetUserDataAnswer: &pb.GetUserDataAnswer{
					Data: data,
				},
			},
		}

		//log.Println("PROTO HandleGetUserData ", pbm)

		result, err := proto.Marshal(pbm)
		if err != nil {
			log.Println("Error when proto.Marshal!!!! ", err)
		}

		client.send <- result
	}()

	userDataFields := req.Fields
	data, err := user.GetUserData(req.Users, userDataFields)
	if err != nil {
		log.Println("Error in HandleGetUserData, GetUserData ", err)
		errStr = "Error in HandleGetUserData" + err.Error()
		return
	}
}
