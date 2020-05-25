package main

import (
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		logrus.Errorf("failed to listen 8080, error = %v", err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			logrus.Errorf("failed to accept conn, error = %v", err)
			return
		}
		//conn = comm.NewReliableConn(conn)
		go handlerConn(conn)
	}
}

func handlerConn(conn net.Conn) {
	//testInfo := []byte("this is a test info")
	for {
		//n, err := conn.Write(testInfo)
		//logrus.Infof("write message, n = %d, error = %v", n, err)
		time.Sleep(1 * time.Second)
	}
}
