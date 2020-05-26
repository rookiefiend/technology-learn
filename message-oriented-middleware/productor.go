package main

import (
	"technology/message-oriented-middleware/client"
	"time"

	"github.com/google/uuid"

	"github.com/sirupsen/logrus"
)

func main() {
	c := client.NewClient("127.0.0.1:8080")

	for {
		msg := uuid.New().String()
		req := client.ProductReq{
			DestName: "test1",
			Msg:      msg,
		}
		err := c.Product(req)
		if err != nil {
			logrus.WithField("req", req).Errorf("failed to product message, error = %v", err)
		} else {
			logrus.Infof("success product message %s", msg)
		}
		time.Sleep(1 * time.Second)
	}
}
