package mongodb

import (
	"context"
	"errors"
	"log"

	"github.com/plankton4/chat-app-server/cmd/server/config"
	"github.com/plankton4/chat-app-server/cmd/server/misc"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	dbname = "chatapp"
)

const (
	// коллекции
	CollChatSubscriptions = "chatsubs"
	CollUsersChatData     = "users_chat_data"
	CollChatMessages      = "chatmess_"
	CollPMessages         = "pmess_"
)

const (
	// поля в коллекциях
	KeyMessageID   = "_id"
	KeyType        = "tp"
	KeyFromUserID  = "fuid"
	KeyToChatID    = "tcid"
	KeyToUserID    = "tuid"
	KeyText        = "t"
	KeyImageURL    = "iu"
	KeyAspectRatio = "ar"
	KeyIsEdited    = "e"

	KeyClientID         = "cid"
	KeyChatID           = "ci"
	KeySubscribedUsers  = "su"
	KeyLastSeenMessages = "lsm"
)

var MongoClient *mongo.Client
var hostAddr string
var NotFoundError = errors.New("Not found")

const (
	localAddr  = "mongodb://localhost:27017"
	remoteAddr = config.RemoteMongoAddress
)

func init() {
	if misc.IsConnectedToRemoteServer() {
		hostAddr = remoteAddr
	} else {
		hostAddr = localAddr
	}
	log.Println("Hostname for mongo ", hostAddr)
}

func RunDB() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(hostAddr))

	err = client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		log.Println("Error in RunMongoDB ", err.Error())
	}

	MongoClient = client

	initChatSubsCollection()
	initUsersChatDataCollection()
}

func Disconnect() {
	if err := MongoClient.Disconnect(context.TODO()); err != nil {
		panic(err)
	}
}
