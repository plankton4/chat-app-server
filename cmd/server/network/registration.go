package network

import (
	"log"

	"github.com/plankton4/chat-app-server/cmd/server/global"
	"github.com/plankton4/chat-app-server/cmd/server/user"
	"github.com/plankton4/chat-app-server/pb"
)

func WorkHttpEndRegistration(queryParams QueryParamsData) (
	result interface{}, errID uint32, errStr string) {

	type Result struct {
		Success bool `json:"Success"`
	}

	userID := queryParams.GetUInt32("userid")
	if userID == 0 {
		log.Println("Error during registration! userID is 0")
		return
	}

	name := queryParams.Get("name")
	if name == "" {
		log.Println("Error during registration! name is empty")
		return
	}

	age := queryParams.GetUInt32("age")
	if age == 0 {
		log.Println("Error during registration! age is empty")
		return
	}

	gender := queryParams.GetUInt32("gender")
	cityName := queryParams.Get("cityname")

	isUserExists := user.IsUserExists(userID)
	log.Println("isUserExists ", isUserExists)
	if !isUserExists {
		errStr = "User not exists, registration is required"
		errID = global.ErrorUserRegistration
		return
	}

	fields := []pb.UserDataField{
		pb.UserDataField_FieldName,
		pb.UserDataField_FieldAge,
		pb.UserDataField_FieldGender,
		pb.UserDataField_FieldCityName,
	}

	rowsUpdated, updateErr := user.UpdateUserData(
		userID,
		fields,
		user.UserData{
			Name:     name,
			Age:      age,
			Gender:   gender,
			CityName: cityName,
		})
	if updateErr != nil {
		log.Println("Error during registration data update ", updateErr)
	}

	isSuccess := rowsUpdated > 0

	result = &Result{
		Success: isSuccess,
	}

	return
}
