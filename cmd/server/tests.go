package main

import (
	"log"
	"time"

	"github.com/plankton4/chat-app-server/cmd/server/database/mongodb"
)

func mainTest() {
	go func() {
		timer := time.NewTimer(1 * time.Second)
		<-timer.C

		// GetUnreadCount
		var chatID uint32 = 1001
		var chatID2 uint32 = 1002
		coll := mongodb.GetMessagesCollection(1, nil, &chatID)
		unreadCount := mongodb.GetUnreadCount(1, coll, "628c6dbe6e0c551623fdc085")
		log.Println("UNREAD COUNT ", unreadCount)

		// GetLastSeenMessageID
		mongodb.UpdateLastSeenMessage(1, nil, &chatID, "0")
		mongodb.UpdateLastSeenMessage(1, nil, &chatID2, "0")
		mongodb.UpdateLastSeenMessage(2, nil, &chatID, "0")

		// GetLastSeenMessageID
		mongodb.GetLastSeenMessageID(1, nil, &chatID2)

		// getting user data
		// userIDs := []uint32{1, 2}
		// fields := []pb.UserDataField{
		// 	pb.UserDataField_FieldUserID,
		// 	pb.UserDataField_FieldGender,
		// 	pb.UserDataField_FieldName,
		// }
		// usersData, err := userdata.GetUserData(userIDs, fields)

		// if err != nil {
		// 	log.Println("Error getting data ", err)
		// }

		// for index, value := range usersData {
		// 	log.Println("Index: ", index, " User data: ", value)
		// }
	}()
}
