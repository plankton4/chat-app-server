package network

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/plankton4/chat-app-server/cmd/server/misc"
)

type QueryParamItem struct {
	Name  string
	Value string
}

type QueryParamsData map[string]*QueryParamItem

func (v QueryParamsData) Set(s, value string) {
	v[strings.ToLower(s)] = &QueryParamItem{
		Name:  s,
		Value: value,
	}
}

func (v QueryParamsData) _get(s string) *QueryParamItem {
	if d, ok := v[strings.ToLower(s)]; ok {
		return d
	}
	return &QueryParamItem{Name: s}
}

func (v QueryParamsData) Get(s string) string {
	return v._get(s).Value
}

func (v QueryParamsData) GetUInt32(s string) uint32 {
	return misc.StrToUInt32V(v.Get(s))
}

// HttpRequestData - данные входящего HTTP запроса
type HttpRequestData struct {
	writer  http.ResponseWriter
	request *http.Request
	path    string

	result      *BaseHttpJsonResult
	protoResult *BaseProtoResult

	// lowerPathA кэш массива для функции GetLowerPathA
	lowerPathA []string
	// pathA кэш массива для функции GetPathA
	pathA []string
}

// NewRequest конструктор HttpRequestData
func NewRequestData(writer http.ResponseWriter, request *http.Request, path string) *HttpRequestData {
	return &HttpRequestData{
		writer:  writer,
		request: request,
		path:    path,
	}
}

// GetHttpResult получить стандартный обработчик результата выполнения запроса
func (v *HttpRequestData) GetHttpResult() *BaseHttpJsonResult {
	if v.result == nil {
		v.result = &BaseHttpJsonResult{
			reqData: v,
		}
	}
	return v.result
}

func (v *HttpRequestData) GetProtoResult() *BaseProtoResult {
	if v.protoResult == nil {
		v.protoResult = &BaseProtoResult{
			reqData: v,
		}
	}
	return v.protoResult
}

// GetWriter получение http.ResponseWriter
func (v *HttpRequestData) GetWriter() http.ResponseWriter {
	if v == nil {
		return nil
	}
	return v.writer
}

// GetPostParams получить данные GET в виде ключ-значение
func (v *HttpRequestData) GetQueryParams() (QueryParamsData, error) {
	urlValues, err := url.ParseQuery(v.request.URL.RawQuery)
	if err != nil {
		return nil, err
	}

	var params = QueryParamsData{}
	for k, v := range urlValues {
		if len(v) > 1 {
			continue
		}
		params.Set(k, v[0])
	}
	return params, nil
}

// GetBody получить тело запроса
func (v *HttpRequestData) GetBody() ([]byte, error) {
	if v == nil || v.request == nil || v.request.Body == nil {
		return []byte{}, nil
	}

	return ioutil.ReadAll(v.request.Body)
}

// BaseHttpJsonResult обработчик результата возращаемого на HTTP запрос
type BaseHttpJsonResult struct {
	// ErrorStr текст ошибки обработки запроса
	ErrorStr string `json:"Error,omitempty"`
	// ErrorDevelop текст ошибки обработки запроса с данными для разработчика
	ErrorDevelop string `json:"ErrorDevelop,omitempty"`
	// ErrorID код ошибки обработки запроса
	ErrorID uint32 `json:"ErrorID,omitempty"`
	// Data результат обработки запроса
	Data interface{} `json:"Data,omitempty"`
	// ServerTime серверное время в формате UnixTime
	ServerTime uint32 `json:"ServerTime,omitempty"`

	// IsWritedResult признак отправки результата обработчиком в поток клиента
	// если установлен в true то функция Write() не будет посылать данные в формате JSON
	IsWritedResult bool             `json:"-"`
	reqData        *HttpRequestData `json:"-"`
}

// Write зотправить результат клиенту
func (v *BaseHttpJsonResult) Write() {
	if v.IsWritedResult {
		return
	}
	//v.ServerTime = misc.GetServerTimeUint32()
	b, err := v.Marshal()
	if err == nil {
		_, err = v.GetRequest().GetWriter().Write(b)
	}
	if err != nil {
		//v.GetRequest().logger.Error("network.%v.Write err:%v", v.GetRequest().GetLowerPathA(), err.Error())
	}
}

// GetRequest получить данные зпроса
func (v *BaseHttpJsonResult) GetRequest() *HttpRequestData {
	if v == nil || v.reqData == nil {
		return &HttpRequestData{}
	}

	return v.reqData
}

// Marshal получить результат в формате JSON
func (v *BaseHttpJsonResult) Marshal() ([]byte, error) {
	return json.Marshal(v)
}

// IsError проверка ошибки и добавление её в ответ клиенту
func (v *BaseHttpJsonResult) IsError(err error) bool {
	if v == nil {
		return false
	}
	// if err != nil && err != sql.ErrNoRows && err != redis.ErrNil {
	// 	// добавляем хидер Warning с текстом ошибки
	// 	v.GetRequest().GetWriter().Header().Add("Warning", fmt.Sprintf("%v (%T)", err.Error(), err))
	// 	v.ErrorStr = err.Error()
	// 	// v.ErrorDevelop = strings.Join(misc.GetSourceCodeCalls(4), "->")
	// 	// v.GetRequest().logger.Error("network.IsError err:%v, ErrorDevelop:%v",
	// 	// 	err.Error(), v.ErrorDevelop)
	// 	return true
	// }
	return false
}
