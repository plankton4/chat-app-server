package network

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var lastVideo string = "XG9o1GJrsPk"

const lastVideoType = "lastVideo"

func GetLastVideo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(lastVideo)
	}
}

func takeVideoUrl(dataMap map[string]interface{}) {
	if videoId, ok := dataMap["videoId"]; ok {
		lastVideo = videoId.(string)
		fmt.Println("LastVideo = ", lastVideo)
	}
}
