package main

import (
	"log"
	"time"

	"firebase.google.com/go/v4/messaging"
	"github.com/plankton4/chat-app-server/cmd/server/fcm"
	"github.com/plankton4/chat-app-server/cmd/server/user"
	"github.com/plankton4/chat-app-server/pb"
)

func mainTest() {
	go func() {
		timer := time.NewTimer(1 * time.Second)
		<-timer.C

		// // GetUnreadCount
		// var chatID uint32 = 1001
		// var chatID2 uint32 = 1002
		// coll := mongodb.GetMessagesCollection(1, nil, &chatID)
		// unreadCount := mongodb.GetUnreadCount(1, coll, "628c6dbe6e0c551623fdc085")
		// log.Println("UNREAD COUNT ", unreadCount)

		// // GetLastSeenMessageID
		// mongodb.UpdateLastSeenMessage(1, nil, &chatID, "0")
		// mongodb.UpdateLastSeenMessage(1, nil, &chatID2, "0")
		// mongodb.UpdateLastSeenMessage(2, nil, &chatID, "0")

		// // GetLastSeenMessageID
		// mongodb.GetLastSeenMessageID(1, nil, &chatID2)

		//testGettingUserData()

		//testSendPush()
	}()
}

func testSendPush() {
	badge := 1
	sound := "default"

	iphonetoken := []string{"fVEaWXndgUjJt2mJKcdaPd:APA91bGwbhvIlSbpRQkdm9CSDm02rSodbRIdSKnEwq_SqIyshYPkjzbthX9rz7Zig15LbwQt2bvM6GOl9CZ0mjc3VaCsDK77AnjeJr8e__k8GAZWD9CwYWfX7pawczytc_3jywusHnCG"}

	fcm.SendMulticast(
		fcm.FirebaseApp,
		iphonetoken,
		messaging.Notification{
			Title: "Duck",
			Body:  "Hi!",
		},
		messaging.Aps{
			Badge: &badge,
			Sound: sound,
		},
		map[string]string{},
	)
}

func testGettingUserData() {
	userIDs := []uint32{1, 2}
	fields := []pb.UserDataField{
		pb.UserDataField_FieldUserID,
		pb.UserDataField_FieldGender,
		pb.UserDataField_FieldName,
	}
	usersData, err := user.GetUserData(userIDs, fields)

	if err != nil {
		log.Println("Error getting data ", err)
	}

	for index, value := range usersData {
		log.Println("Index: ", index, " User data: ", value)
	}
}
