package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"technology/message-oriented-middleware/comm"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	defaultRetryInternal = 1 * time.Second
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
	defer request.Body.Close()
	v := new(ConsumeReq)
	err := json.NewDecoder(request.Body).Decode(v)
	if err != nil {
		ServeJSON(writer, http.StatusBadRequest, comm.ResponseData{
			Err: err.Error(),
		})
		return
	}

	queue, ok := ConsumeQueueMap.Get(v.DestName)
	if !ok {
		ServeJSON(writer, http.StatusBadRequest, comm.ResponseData{
			Err: fmt.Sprintf("consume a unknown dest name %s", v.DestName),
		})
		return
	}

	ServeJSON(writer, http.StatusOK, comm.ResponseData{
		Msg:  "consume msg success",
		Data: ConsumeResp{Msg: queue.Get()},
	})
}

// 生产者生产函数
var Product http.HandlerFunc = func(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	v := new(ProductReq)
	err := json.NewDecoder(request.Body).Decode(v)
	if err != nil {
		ServeJSON(writer, http.StatusBadRequest, comm.ResponseData{
			Err: err.Error(),
		})
		return
	}

	queue, ok := ConsumeQueueMap.Get(v.DestName)
	if !ok {
		ServeJSON(writer, http.StatusBadRequest, comm.ResponseData{
			Err: fmt.Sprintf("product msg to a unknown dest name %s", v.DestName),
		})
		return
	}
	err = queue.Add(v.Msg)
	if err != nil {
		ServeJSON(writer, http.StatusInternalServerError, comm.ResponseData{
			Err: err.Error(),
		})
		return
	}

	ServeJSON(writer, http.StatusOK, comm.ResponseData{
		Msg: "product msg success",
	})
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
	for {
		_, err = write.Write(dataJSON)
		if err != nil {
			logrus.Errorf("failed to write json resp [%s], error = %v", dataJSON, err)
			time.Sleep(defaultRetryInternal)
		}
	}
}
