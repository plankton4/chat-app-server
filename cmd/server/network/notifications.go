package network

import (
	"github.com/plankton4/chat-app-server/cmd/server/database"
	"github.com/plankton4/chat-app-server/cmd/server/database/mongodb"
)

func subscribeToChat(userID uint32, chatID uint32) {
	mongodb.SubscribeToChat(userID, chatID)
}

func GetSubscribers(chatID uint32) ([]uint32, error) {
	return mongodb.GetSubscribers(chatID)
}

func getTokens(userIDs []uint32) ([]string, error) {
	return database.GetFCMTokens(userIDs)
}
