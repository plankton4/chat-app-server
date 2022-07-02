package user

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/plankton4/chat-app-server/cmd/server/database"
	"github.com/plankton4/chat-app-server/pb"
)

var fieldNames = map[pb.UserDataField]string{
	pb.UserDataField_FieldUserID:   "user_id",
	pb.UserDataField_FieldName:     "name",
	pb.UserDataField_FieldAge:      "age",
	pb.UserDataField_FieldGender:   "gender",
	pb.UserDataField_FieldCityName: "city_name",
}

type UserData struct {
	UserId   uint32
	Name     string
	Age      uint32
	Gender   uint32
	CityName string
}

func IsUserExists(userId uint32) bool {
	log.Println("IsUserExists func, userID: ", userId)

	row := database.MainDB.QueryRow(`
			SELECT 1 FROM Users WHERE user_id = ?
		`, userId)

	dummyNum := 0
	err := row.Scan(&dummyNum)

	if err == sql.ErrNoRows {
		return false
	} else {
		return true
	}
}

func GetUserDataOne(userID uint32, fields []pb.UserDataField) (*pb.UserData, error) {
	userDataSlice, err := GetUserData([]uint32{userID}, fields)

	if err != nil {
		return nil, err
	}

	if len(userDataSlice) != 0 {
		return userDataSlice[0], nil
	}

	return nil, nil
}

func GetUserData(userIDs []uint32, fields []pb.UserDataField) ([]*pb.UserData, error) {
	userDataSlice := make([]*pb.UserData, 0)

	rowString := "SELECT "

	fieldsLen := len(fields)
	for index, field := range fields {
		rowString += fieldNames[field]
		if index != fieldsLen-1 {
			rowString += ","
		}
	}

	args := make([]interface{}, len(userIDs))
	for i, id := range userIDs {
		args[i] = id
	}

	rowString += " FROM Users WHERE user_id IN (?" + strings.Repeat(",?", len(args)-1) + ")"
	rows, err := database.MainDB.Query(rowString, args...)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		userData := pb.UserData{}
		rowsToScan := []interface{}{}

		for _, field := range fields {
			// fill rowsToScan
			switch field {
			case pb.UserDataField_FieldUserID:
				rowsToScan = append(rowsToScan, &userData.UserID)
			case pb.UserDataField_FieldName:
				rowsToScan = append(rowsToScan, &userData.Name)
			case pb.UserDataField_FieldGender:
				rowsToScan = append(rowsToScan, &userData.Gender)
			case pb.UserDataField_FieldCityName:
				rowsToScan = append(rowsToScan, &userData.CityName)
			}
		}

		if err = rows.Scan(rowsToScan...); err != nil {
			if err == sql.ErrNoRows {
				continue
				//return userDataSlice, fmt.Errorf("getUserData %d: no such users", userIDs)
			}
			return userDataSlice, fmt.Errorf("ERROR! getUserData %d: %v", userIDs, err)
		}

		userDataSlice = append(userDataSlice, &userData)
	}

	if err = rows.Err(); err != nil {
		return userDataSlice, err
	}

	return userDataSlice, nil
}

func UpdateUserData(userId uint32, fields []pb.UserDataField, data UserData) (int64, error) {
	fieldsLen := len(fields)
	updateValues := []interface{}{}

	rowString := "UPDATE Users SET "

	for index, field := range fields {
		rowString += fieldNames[field] + " = ?"
		if index != fieldsLen-1 {
			rowString += ","
		}

		// fill rowsToScan
		switch field {
		case pb.UserDataField_FieldName:
			updateValues = append(updateValues, data.Name)
		case pb.UserDataField_FieldAge:
			updateValues = append(updateValues, data.Age)
		case pb.UserDataField_FieldGender:
			updateValues = append(updateValues, data.Gender)
		case pb.UserDataField_FieldCityName:
			updateValues = append(updateValues, data.CityName)
		}
	}

	rowString += " WHERE user_id = ?"
	updateValues = append(updateValues, userId)

	res, err := database.MainDB.Exec(rowString, updateValues...)
	if err != nil {
		log.Println("Error! During UpdateUserData ", err.Error())
	}

	rowsAffected, _ := res.RowsAffected()
	log.Println("Updated rows ", rowsAffected)

	return rowsAffected, err
}
