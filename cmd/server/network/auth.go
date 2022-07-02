package network

import (
	"log"

	"github.com/golang/protobuf/proto"
	"github.com/plankton4/chat-app-server/cmd/server/global"
	"github.com/plankton4/chat-app-server/cmd/server/user"
	"github.com/plankton4/chat-app-server/pb"
)

type AuthResult struct {
	UserID         uint32 `json:"UserID,omitempty"`
	SessionKey     string `json:"SessionKey,omitempty"`
	IsRegistration uint32 `json:"IsRegistration"`
}

func HandleAuthentication(client *Client, rowID uint32, req *pb.AuthenticationReq) {
	log.Println("Handle auth, isFirstTime? ", req.IsFirstAuthentication)

	var errID uint32
	var errStr string
	var userID uint32
	var isRegistration uint32 = 1
	var sessionKey string

	defer func() {
		pbm := &pb.PBMessage{
			RowID:  rowID,
			ErrID:  &errID,
			ErrStr: &errStr,
			InternalMessage: &pb.PBMessage_MessAuthAnswer{
				MessAuthAnswer: &pb.AuthenticationAnswer{
					UserID:         userID,
					IsRegistration: isRegistration,
					SessionKey:     sessionKey,
				},
			},
		}

		log.Println("PROTO HandleAuthentication ", pbm)

		result, err := proto.Marshal(pbm)
		if err != nil {
			log.Println("Error!!!! ", err)
		}

		client.send <- result
	}()

	userID = req.UserID
	if userID == 0 {
		log.Println("Error! userID is 0")
		return
	}

	isUserExists := user.IsUserExists(userID)
	log.Println("isUserExists ", isUserExists)
	if !isUserExists {
		errStr = "User not exists, registration is required"
		errID = global.ErrorUserRegistration
		return
	}

	sessionKey = req.SessionKey
	savedSessionKey, _ := user.UserSessionKey(userID)

	if savedSessionKey != "" && savedSessionKey == sessionKey {
		isRegistration = 0
		client.isAuthenticated = true
		client.clientID = userID
		ConnectedClients[userID] = client

		/// добавляем хабы клиенту
		for hubID, activeHub := range ActiveHubs {
			client.hubs[hubID] = activeHub
		}

		// регаем клиента в каждом добавленном хабе
		for _, hub := range client.hubs {
			hub.register <- client

			if req.IsFirstAuthentication {
				go subscribeToChat(client.clientID, uint32(hub.hubID))
			}
		}
	} else {
		errStr = "Not equal session keys"
		errID = global.ErrorSessionKey
		return
	}
}
