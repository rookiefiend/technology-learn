package main

import (
	"net"
	"technology/message-oriented-middleware/client"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	c := client.NewClient("127.0.0.1:8080")
	err := c.RegistryDestName(client.RegistryDestNameReq{DestName: "xxx"})
	if err != nil {
		logrus.Errorf("failed to registry dest name, error = %v", err)
		return
	}
	logrus.Infof("success registry dest name")
}

func receive(conn net.Conn) {
	testBytes := make([]byte, 10)
	for {
		n, err := conn.Read(testBytes)
		logrus.Errorf("read conn, n = %d, err = %v, testBytes = %s", n, err, testBytes)
		time.Sleep(1 * time.Second)
	}
}
