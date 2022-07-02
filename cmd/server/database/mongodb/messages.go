package mongodb

import (
	"context"
	"log"
	"strconv"

	"github.com/plankton4/chat-app-server/pb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	getAllMessagesLimit = 100
)

func AddChatMessage(clientID uint32, userID *uint32, chatID *uint32, pbMess *pb.ChatMessageData) (insertedID interface{}) {
	if userID == nil && chatID == nil {
		log.Fatal("Both userID and chatID are nil!")
	}

	collectionName := ""
	if chatID != nil {
		collectionName = CollChatMessages + strconv.FormatUint(uint64(*chatID), 10)
	} else {
		collectionName = getPMCollectionName(clientID, *userID)
	}

	messagesCollection := MongoClient.Database(dbname).Collection(collectionName)

	result, err := messagesCollection.InsertOne(context.TODO(), convertToBChatMessageData(pbMess))

	if err != nil {
		panic(err)
	}

	if result != nil {
		insertedID = result.InsertedID
	}

	return
}

func EditChatMessage(pbMess *pb.ChatMessageData, fieldsToUpdate []string) error {
	clientID := pbMess.FromUserID
	toUserID := pbMess.ToUserID
	toChatID := pbMess.ToChatID

	if toUserID == nil && toChatID == nil {
		log.Fatal("Both userID and chatID are nil!")
	}

	id, err := primitive.ObjectIDFromHex(pbMess.MessageID)
	if err != nil {
		log.Fatal("ID NOT FOUND ", id)
	}

	log.Println("id to update ", id)

	filter := bson.D{{Key: "_id", Value: id}}
	newFields := make(bson.D, len(fieldsToUpdate))

	for i, field := range fieldsToUpdate {
		log.Println("Field ", field)
		var eValue interface{}

		switch field {
		case KeyIsEdited:
			eValue = true
		case KeyText:
			eValue = pbMess.Text
		}

		newFields[i] = primitive.E{Key: field, Value: eValue}
	}

	log.Println("try to update, newFields ", newFields)
	update := bson.D{{Key: "$set", Value: newFields}}

	collectionName := ""
	if toChatID != nil {
		collectionName = CollChatMessages + strconv.FormatUint(uint64(*toChatID), 10)
	} else {
		collectionName = getPMCollectionName(clientID, *toUserID)
	}

	messagesCollection := MongoClient.Database(dbname).Collection(collectionName)

	_, err = messagesCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Println("Error in EditChatMessage ", err)
		panic(err)
	}

	return err
}

func DeleteChatMessage(messageID string, fromUserID uint32, toUserID *uint32, toChatID *uint32) error {
	if toUserID == nil && toChatID == nil {
		log.Fatal("Both userID and chatID are nil!")
	}

	id, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		log.Fatal("ID CREATION FAILED ", id)
	}

	log.Println("id to delete ", id)

	filter := bson.D{{Key: "_id", Value: id}}

	collectionName := ""
	if toChatID != nil {
		collectionName = CollChatMessages + strconv.FormatUint(uint64(*toChatID), 10)
	} else {
		collectionName = getPMCollectionName(fromUserID, *toUserID)
	}

	messagesCollection := MongoClient.Database(dbname).Collection(collectionName)

	_, err = messagesCollection.DeleteOne(context.TODO(), filter)
	if err != nil {
		log.Println("Error in DeleteChatMessage ", err)
		panic(err)
	}

	return err
}

func GetAllMessages(clientID uint32, userID *uint32, chatID *uint32) (results []*pb.ChatMessageData, err error) {
	messagesCollection := GetMessagesCollection(clientID, userID, chatID)
	filter := bson.D{}

	options := options.Find()
	options.SetSort(bson.D{{Key: "_id", Value: -1}})
	options.SetLimit(getAllMessagesLimit)

	cursor, err := messagesCollection.Find(context.TODO(), filter, options)
	defer cursor.Close(context.TODO())

	if err != nil {
		log.Fatal(err)
	}

	results = make([]*pb.ChatMessageData, 0)

	for cursor.Next(context.TODO()) {
		var result BChatMessageData
		if err := cursor.Decode(&result); err != nil {
			log.Fatal(err)
		}
		//log.Printf("Result: %+v", result)
		results = append(results, convertFromBChatMessage(&result))
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	return
}

func GetMessageByID(messageID string, clientID uint32, userID *uint32, chatID *uint32) (result *pb.ChatMessageData, err error) {
	id, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		log.Fatal("ID CREATION FAILED ", id)
	}

	messagesCollection := GetMessagesCollection(clientID, userID, chatID)

	var bMessageData BChatMessageData

	filter := bson.D{{Key: "_id", Value: id}}
	err = messagesCollection.FindOne(context.TODO(), filter).Decode(&bMessageData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// This error means your query did not match any documents.
			return
		}
		panic(err)
	}

	return convertFromBChatMessage(&bMessageData), nil
}

func GetMessagesCollection(clientID uint32, userID *uint32, chatID *uint32) *mongo.Collection {
	collectionName := ""
	if chatID != nil {
		collectionName = CollChatMessages + strconv.FormatUint(uint64(*chatID), 10)
	} else {
		collectionName = getPMCollectionName(clientID, *userID)
	}

	return MongoClient.Database(dbname).Collection(collectionName)
}

// for private messages
func getPMCollectionName(clientID uint32, userID uint32) string {
	return CollPMessages +
		"f" + strconv.FormatUint(uint64(clientID), 10) +
		"t" + strconv.FormatUint(uint64(userID), 10)
}

/**
 * Converting to and from bson
 */

type BChatMessageData struct {
	MessageID      string            `bson:"_id,omitempty"`
	Type           int32             `bson:"tp,omitempty"`
	FromUserID     uint32            `bson:"fuid"`
	Time           *uint32           `bson:"tm"`
	ToChatID       *uint32           `bson:"tcid,omitempty"`
	ToUserID       *uint32           `bson:"tuid,omitempty"`
	Text           *string           `bson:"t,omitempty"`
	IsEdited       *bool             `bson:"e,omitempty"`
	ImageURL       *string           `bson:"iu,omitempty"`
	AspectRatio    *float32          `bson:"ar,omitempty"`
	RepliedMessage *BChatMessageData `bson:"rp,omitempty"`
}

func convertToBChatMessageData(pbMess *pb.ChatMessageData) *BChatMessageData {
	if pbMess == nil {
		return nil
	}

	return &BChatMessageData{
		Type:           int32(pbMess.Type),
		FromUserID:     pbMess.FromUserID,
		Time:           pbMess.Time,
		ToChatID:       pbMess.ToChatID,
		ToUserID:       pbMess.ToUserID,
		Text:           pbMess.Text,
		IsEdited:       pbMess.IsEdited,
		ImageURL:       pbMess.ImageURL,
		AspectRatio:    pbMess.AspectRatio,
		RepliedMessage: convertToBChatMessageData(pbMess.RepliedMessage),
	}
}

func convertFromBChatMessage(bMess *BChatMessageData) *pb.ChatMessageData {
	if bMess == nil {
		return nil
	}

	return &pb.ChatMessageData{
		MessageID:      bMess.MessageID,
		Type:           pb.ChatMessageType(bMess.Type),
		FromUserID:     bMess.FromUserID,
		Time:           bMess.Time,
		ToChatID:       bMess.ToChatID,
		ToUserID:       bMess.ToUserID,
		Text:           bMess.Text,
		IsEdited:       bMess.IsEdited,
		ImageURL:       bMess.ImageURL,
		AspectRatio:    bMess.AspectRatio,
		RepliedMessage: convertFromBChatMessage(bMess.RepliedMessage),
	}
}
