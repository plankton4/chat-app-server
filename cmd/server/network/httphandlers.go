package network

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/plankton4/chat-app-server/cmd/server/config"
	"github.com/plankton4/chat-app-server/cmd/server/database"
	"github.com/plankton4/chat-app-server/cmd/server/database/mongodb"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func HttpHandlerWrapper(
	reqData *HttpRequestData,
	handlerFunc func(queryParams QueryParamsData) (
		result interface{}, errID uint32, errStr string),
) {
	result := reqData.GetHttpResult()

	defer func() {
		result.Write()
	}()

	queryParams, err := getQueryParams(reqData)
	if err != nil {
		log.Println("Error in getQueryParams ", err.Error())
		return
	}

	resultData, errID, errStr := handlerFunc(*queryParams)

	fmt.Printf("resultData %+v \nerrID: %v \nerrStr: %v \n", resultData, errID, errStr)

	if errID != 0 {
		result.ErrorID = errID
		result.ErrorStr = errStr
	}

	result.Data = resultData
}

func getQueryParams(reqData *HttpRequestData) (*QueryParamsData, error) {
	queryParams, err := reqData.GetQueryParams()

	postParams := map[string]string{}
	if b, _ := reqData.GetBody(); len(b) > 0 {
		err = json.Unmarshal(b, &postParams)
		if err != nil {
			return nil, errors.New("Unmarshal error")
		}
	}
	for p, v := range postParams {
		queryParams.Set(p, v)
	}

	log.Println("Query params:")
	for k, v := range queryParams {
		log.Println("Key: ", k, " Value: ", v)
	}

	return &queryParams, nil
}

func ServeHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode("home")
	}
}

func Ping() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode("pong")
	}
}

func PingMysqlDB() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		res := "no mysql error"

		err := database.MainDB.Ping()
		if err != nil {
			log.Println("Pinging database error ", err.Error())
			res = err.Error()
		}
		json.NewEncoder(w).Encode(res)
	}
}

func PingMongo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		res := "no mongodb error"

		if mongodb.MongoClient != nil {
			err := mongodb.MongoClient.Ping(context.TODO(), readpref.Primary())
			if err != nil {
				log.Println("Pinging mongo error ", err.Error())
				res = err.Error()
			}
		} else {
			res = "Error! Mongo client is nil."
		}

		json.NewEncoder(w).Encode(res)
	}
}

func PingServerAddr() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		res := []string{}
		isContains := false

		addrs, _ := net.InterfaceAddrs()

		for _, addr := range addrs {
			if strings.Contains(addr.String(), config.RemoteServerAddress) {
				isContains = true
			}
			res = append(res, addr.String())
		}

		res = append(res, fmt.Sprintf("%t", isContains))
		json.NewEncoder(w).Encode(res)
	}
}

func AppleSignInAuthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqData := NewRequestData(w, r, r.URL.Path)
		HttpHandlerWrapper(reqData, WorkHttpAppleSignInAuth)
	}
}

func RegistrationEndHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqData := NewRequestData(w, r, r.URL.Path)
		HttpHandlerWrapper(reqData, WorkHttpEndRegistration)
	}
}

func UploadImageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		WorkUploadImage(w, r)
	}
}
