package network

import (
	"encoding/json"
	"log"
)

type BaseProtoResult struct {
	ErrorID  uint32 `json:"ErrorID,omitempty"`
	ErrorStr string `json:"Error,omitempty"`
	Data     []byte
	reqData  *HttpRequestData
}

func (v *BaseProtoResult) GetRequest() *HttpRequestData {
	if v == nil || v.reqData == nil {
		return &HttpRequestData{}
	}

	return v.reqData
}

func (v *BaseProtoResult) Write() {
	b, err := v.Marshal()

	_, err = v.GetRequest().GetWriter().Write(b)
	if err != nil {
		log.Println("Error when write BaseProtoResult")
	}
}

func (v *BaseProtoResult) Marshal() ([]byte, error) {
	return json.Marshal(v)
}
