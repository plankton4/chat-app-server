package mongodb

import (
	"context"
	"errors"
	"log"
	"strconv"

	"github.com/plankton4/chat-app-server/pb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var chatSubsCollection *mongo.Collection
var usersChatDataCollection *mongo.Collection

type BUsersChatData struct {
	ClientID         uint32            `bson:"cid"`
	LastSeenMessages map[string]string `bson:"lsm"`
}

type UnreadInfo struct {
	UnreadCount       uint32
	LastSeenMessageID string
}

/*===========================================================================
INIT FUNCTIONS
=============================================================================*/
func initChatSubsCollection() {
	chatSubsCollection = MongoClient.Database(dbname).Collection(CollChatSubscriptions)

	/// создаем индекс. Повторно индексы не создаются, то есть проверок не нужно на "isExists"
	indexView := chatSubsCollection.Indexes()

	model := mongo.IndexModel{
		Keys:    bson.D{{Key: KeyChatID, Value: 1}},
		Options: options.Index().SetName("KeyChatID"),
	}
	indexOpts := options.CreateIndexes()
	_, err := indexView.CreateOne(context.TODO(), model, indexOpts)

	if err != nil {
		log.Println("Error while creating KeyChatID index in mongo ", err.Error())
	}
	// закончили создание индекса
}

func initUsersChatDataCollection() {
	usersChatDataCollection = MongoClient.Database(dbname).Collection(CollUsersChatData)

	/// создаем индекс. Повторно индексы не создаются, то есть проверок не нужно на "isExists"
	indexView := usersChatDataCollection.Indexes()

	model := mongo.IndexModel{
		Keys:    bson.D{{Key: KeyClientID, Value: 1}},
		Options: options.Index().SetName("KeyClientID"),
	}
	indexOpts := options.CreateIndexes()
	_, err := indexView.CreateOne(context.TODO(), model, indexOpts)

	if err != nil {
		log.Println("Error while creating KeyChatID index in mongo ", err.Error())
	}
	// закончили создание индекса
}

/*===========================================================================
SUBSCRIBE
=============================================================================*/
func SubscribeToChat(userID uint32, chatID uint32) {
	filter := bson.D{{Key: KeyChatID, Value: chatID}}
	update := bson.D{
		{
			Key: "$addToSet",
			Value: bson.D{
				{
					Key:   KeySubscribedUsers,
					Value: userID,
				},
			},
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := chatSubsCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		log.Fatal(err)
	}
}

func GetSubscribers(chatID uint32) ([]uint32, error) {
	type BChatSubscribers struct {
		ChatID          uint32   `bson:"ci"`
		SubscribedUsers []uint32 `bson:"su"`
	}

	if MongoClient == nil {
		return nil, errors.New("MongoClient is nil")
	}

	opts := options.FindOne()

	var result BChatSubscribers

	err := chatSubsCollection.FindOne(
		context.TODO(),
		bson.D{{Key: KeyChatID, Value: chatID}},
		opts,
	).Decode(&result)

	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in
		// the collection.
		if err == mongo.ErrNoDocuments {
			return nil, err
		}
		log.Fatal(err)
	}

	//log.Printf("found document %v", result)

	return result.SubscribedUsers, err
}

/*===========================================================================
UNREAD COUNTERS / LAST SEEN MESSAGE
=============================================================================*/
func UpdateLastSeenMessage(clientID uint32, userID *uint32, chatID *uint32, messageID string) {
	filter := bson.D{{Key: KeyClientID, Value: clientID}}

	roomName, err := getLastSeenMessageRoomName(userID, chatID)
	if err != nil {
		return
	}

	update := bson.D{
		{
			Key: "$set",
			Value: bson.D{
				{
					Key:   KeyLastSeenMessages + "." + roomName,
					Value: messageID,
				},
			},
		},
	}

	opts := options.Update().SetUpsert(true)

	_, err = usersChatDataCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		log.Fatal(err)
	}
}

func GetUnreadInfo(clientID uint32, userID *uint32, chatID *uint32) (*pb.UnreadInfo, error) {
	collection := GetMessagesCollection(clientID, userID, chatID)

	lastSeenMessageID, err := GetLastSeenMessageID(clientID, userID, chatID)
	if err != nil {
		return nil, err
	}

	unreadCount := GetUnreadCount(clientID, collection, lastSeenMessageID)
	unreadInfo := pb.UnreadInfo{
		UserID:            userID,
		ChatID:            chatID,
		UnreadCount:       &unreadCount,
		LastSeenMessageID: &lastSeenMessageID,
	}

	return &unreadInfo, nil
}

func GetLastSeenMessageID(clientID uint32, userID *uint32, chatID *uint32) (string, error) {
	var chatData BUsersChatData

	projection := bson.D{
		{Key: KeyLastSeenMessages, Value: 1},
		{Key: KeyClientID, Value: 1},
	}
	opts := options.FindOne()
	opts.SetProjection(projection)

	filter := bson.D{{
		Key:   KeyClientID,
		Value: clientID,
	}}

	err := usersChatDataCollection.FindOne(context.TODO(), filter, opts).Decode(&chatData)

	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in the collection.
		if err == mongo.ErrNoDocuments {
			return "", nil
		}
		log.Println("Error in GetLastSeenMessageID ", err)
		return "", err
	}

	roomName, err := getLastSeenMessageRoomName(userID, chatID)
	log.Println("Room name ", roomName)
	log.Println("Chat data ", chatData)
	if err != nil {
		log.Println("Error in GetLastSeenMessageID ", err)
		return "", err
	}

	if result, ok := chatData.LastSeenMessages[roomName]; ok {
		return result, nil
	}

	return "", NotFoundError
}

func GetUnreadCount(clientID uint32, collection *mongo.Collection, lastSeenMessageID string) uint32 {
	var res uint32 = 0

	messagesInfo := GetMessagesInfoForUnreadCount(collection, 100)
	for _, info := range messagesInfo {
		if info.MessageID <= lastSeenMessageID {
			break
		}

		if clientID != info.FromUserID {
			res++
		}
	}

	return res
}

type MessagesInfoForUnreadCount struct {
	MessageID  string `bson:"_id"`
	FromUserID uint32 `bson:"fuid"`
}

func GetMessagesInfoForUnreadCount(collection *mongo.Collection, limit int64) []MessagesInfoForUnreadCount {
	projection := bson.D{
		{Key: KeyMessageID, Value: 1},
		{Key: KeyFromUserID, Value: 1},
	}
	opts := options.Find()
	opts.SetProjection(projection)
	opts.SetSort(bson.D{{Key: KeyMessageID, Value: -1}})
	opts.SetLimit(limit)

	cursor, err := collection.Find(context.TODO(), bson.D{}, opts)
	if err != nil {
		log.Panic(err)
	}

	var results []MessagesInfoForUnreadCount
	for cursor.Next(context.TODO()) {
		var result MessagesInfoForUnreadCount
		if err := cursor.Decode(&result); err != nil {
			log.Fatal(err)
		}

		results = append(results, result)
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	return results
}

func getLastSeenMessageRoomName(userID *uint32, chatID *uint32) (string, error) {
	var roomName string
	if userID != nil {
		roomName = "pm_" + strconv.FormatUint(uint64(*userID), 10)
	} else if chatID != nil {
		roomName = "chat_" + strconv.FormatUint(uint64(*chatID), 10)
	} else {
		log.Println("Error in UpdateLastSeenMessage! both userID and chatID are nil!")
		return "", errors.New("Both userID and chatID are nil")
	}

	return roomName, nil
}
