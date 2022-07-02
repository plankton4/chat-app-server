package user

import (
	"database/sql"
	"errors"
	"log"

	"github.com/plankton4/chat-app-server/cmd/server/database"
	"github.com/plankton4/chat-app-server/cmd/server/misc"
)

func RegistrationUser(socnetID string) (userID uint32, sessionKey string, err error) {

	userID, err = GetNewUserId()
	if err != nil {
		log.Printf("Talk.RegistrationUser GlobalGetNewUserId err:%v \n", err.Error())
		return 0, "", err
	}

	log.Println("RegistrationUser, userId ", userID)

	sessionKey, err = NewSessionKeyForUser(userID)
	if err != nil {
		log.Printf("Talk.RegistrationUser INewSessionKeyForUser err:%v \n", err.Error())
		return 0, "", err
	}

	_, err = database.MainDB.Exec(`
		INSERT INTO Users 
			(user_id, apple_id)
		VALUES
			(?, ?)
	`, userID, socnetID)

	if err != nil {
		log.Printf("RegistrationUser Insert users err:%v \n", err.Error())
		return 0, "", err
	}

	return
}

// NewSessionKeyForUser генерация и сохранение ключа сессии для пользователя
func NewSessionKeyForUser(userID uint32) (sessionKey string, err error) {
	sessionKey, err = misc.GetNewUniqueKey()
	if err != nil {
		return "", err
	}

	// сохраняем ключ сессии для юзера
	_, err = database.MainDB.Exec(`
		INSERT INTO SessionKeys 
			(user_id, session_key)
		VALUES
			(?, ?)
		ON DUPLICATE KEY UPDATE session_key = ?;
	`, userID, sessionKey, sessionKey)

	if err != nil {
		return "", err
	}

	return
}

func UserSessionKey(userID uint32) (string, error) {
	var sessionKey string

	row := database.MainDB.QueryRow(
		"SELECT session_key FROM SessionKeys WHERE user_id = ?",
		userID)

	err := row.Scan(&sessionKey)
	if err != nil {
		log.Printf("Row error: %v \n", err.Error())
		if err == sql.ErrNoRows {
			return "", nil
		} else {
			return "", err
		}
	}

	if len(sessionKey) == 0 {
		return NewSessionKeyForUser(userID)
	}

	return sessionKey, nil
}

func GetNewUserId() (uint32, error) {
	r, err := database.MainDB.Exec(`
		INSERT INTO UserIDs 
			() 
		VALUES
			();
	`)
	if err != nil {
		return 0, err
	}

	newUserID, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint32(newUserID), nil
}

func SubscribeToPush(userID uint32, token string) error {
	if userID == 0 {
		return errors.New("Error in SubscribeToPush. UserID is 0")
	}

	_, err := database.MainDB.Exec(`
		INSERT INTO FCMTokens 
			(user_id, token)
		VALUES
			(?, ?)
		ON DUPLICATE KEY UPDATE token = ?;
	`, userID, token, token)

	if err != nil {
		log.Println("Error when insert into FCMTokens ", err.Error())
	}

	return err
}
