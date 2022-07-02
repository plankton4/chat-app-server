package network

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/plankton4/chat-app-server/cmd/server/database"
	"github.com/plankton4/chat-app-server/cmd/server/global"
	"github.com/plankton4/chat-app-server/cmd/server/user"
	"github.com/plankton4/chat-app-server/pb"
)

func WorkHttpAppleSignInAuth(queryParams QueryParamsData) (
	result interface{}, errID uint32, errStr string) {

	token := queryParams.Get("token")

	AppleAccountInfo, err := VerifyAppleToken(token)
	if err != nil {
		fmt.Printf("WorkHttpAppleSignInAuth VerifyAppleToken queryParams:%v, err:%v", queryParams, err)
		errID = global.ErrorAppleToken
		return
	}

	if !AppleAccountInfo.Verified {
		fmt.Printf("WorkHttpAppleSignInAuth VerifyAppleToken queryParams:%v; unverified account", queryParams)
		errStr = "Unverified account"
		errID = global.ErrorAppleToken
		return
	}

	userID, profileExists, err := checkProfileByAppleID(AppleAccountInfo.UID)

	fmt.Printf("checkProfileByAppleID; AppleAccountID: %s, userID: %v, profileExists: %v, error: %v", AppleAccountInfo.UID, userID, profileExists, err)

	if profileExists {
		sessionKey, err := user.UserSessionKey(userID)
		if err != nil {
			errID = global.ErrorSessionKey
			return
		}

		userDataFields := []pb.UserDataField{
			pb.UserDataField_FieldName,
		}

		userData, err := user.GetUserDataOne(userID, userDataFields)
		if err != nil {
			log.Println("Error in GetUserData ", err)
			errStr = "Error in GetUserData" + err.Error()
			return
		}

		log.Printf("User data received %+v \n", userData)

		result = &AuthResult{
			UserID: userID,
			// if firstName is empty then registration is incomplete
			IsRegistration: func(name string) uint32 {
				if name == "" {
					return 1
				}
				return 0
			}(*userData.Name),
			SessionKey: sessionKey,
		}
		return
	}

	userID, sessionKey, err := newAppleUserID(AppleAccountInfo.UID)
	if err != nil {
		errStr = "cannot create new userID"
		errID = global.ErrorUserRegistration
		return
	}

	result = &AuthResult{
		UserID:         userID,
		IsRegistration: 1,
		SessionKey:     sessionKey,
	}
	return
}

func checkProfileByAppleID(appleID string) (uint32, bool, error) {
	userID, err := GetUserByAppleID(appleID)
	return userID, userID != 0, err
}

// GetUserByAppleID проверка наличия аккаунта по appleID
func GetUserByAppleID(appleID string) (userID uint32, err error) {
	row := database.MainDB.QueryRow(`
		SELECT user_id FROM Users WHERE apple_id = ? 
	`, appleID)

	err = row.Scan(&userID)
	if err != nil {
		log.Printf("Row error: %v \n", err.Error())
		if err == sql.ErrNoRows {
			err = nil
		}
	}

	return
}

func newAppleUserID(appleID string) (uint32, string, error) {
	return user.RegistrationUser(appleID)
}
