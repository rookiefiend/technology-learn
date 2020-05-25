package controllers

import (
	"encoding/json"
	"net/http"
	"sync"
	"technology/message-oriented-middleware/comm"

	"github.com/sirupsen/logrus"
)

var (
	ConsumeQueueMap = QueueMap{
		keyMQueue: make(map[string]Queue),
		mutex:     sync.RWMutex{},
	}
)

// 消费者注册函数
var Registry http.HandlerFunc = func(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	v := new(RegistryReq)
	err := json.NewDecoder(request.Body).Decode(v)
	if err != nil {
		ServeJSON(writer, http.StatusBadRequest, comm.ResponseData{
			Err: err.Error(),
		})
		return
	}
	ConsumeQueueMap.Add(v.DestName, NewQueue())
	ServeJSON(writer, http.StatusOK, comm.ResponseData{
		Msg: "registry success",
	})
}

// 消费者消费函数
var Consume http.HandlerFunc = func(writer http.ResponseWriter, request *http.Request) {

}

// 生产者生产函数
var Product http.HandlerFunc = func(writer http.ResponseWriter, request *http.Request) {

}

// ServeJSON ...
func ServeJSON(write http.ResponseWriter, code int, data comm.ResponseData) {
	write.Header().Set("Content-Type", "application/json; charset=utf-8")
	write.WriteHeader(code)

	dataJSON, err := json.Marshal(data)
	if err != nil {
		logrus.WithField("data", data).Errorf("failed to marshal response data")
		return
	}
	write.Write(dataJSON)
}
