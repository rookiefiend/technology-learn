package main

import (
	"technology/message-oriented-middleware/client"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	TestDestName = "test1"
)

func main() {
	c := client.NewClient("127.0.0.1:8080")

	err := c.RegistryDestName(client.RegistryDestNameReq{DestName: TestDestName})
	if err != nil {
		logrus.WithField("destName", TestDestName).
			Errorf("failed to registry dset name, error = %v", err)
		return
	}
	logrus.Infof("success registry dest name")

	for {

		req := client.ConsumeReq{
			DestName: TestDestName,
		}
		consumeResp, err := c.Consume(req)
		if err != nil {
			logrus.WithField("req", req).Errorf("failed to consume message, error = %v", err)
		} else {
			logrus.Infof("success consume message %s", consumeResp.Msg)
		}
		time.Sleep(1 * time.Second)
	}
}
