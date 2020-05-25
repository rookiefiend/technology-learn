package main

import (
	"net/http"
	"technology/message-oriented-middleware/comm"
	"technology/message-oriented-middleware/core/controllers"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Infof("listening :8080")
	l, err := comm.NewReliableListener("tcp", ":8080")
	if err != nil {
		logrus.Fatal("failed to listening 8080 port, error = %v", err)
		return
	}
	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/registry", controllers.Registry)
	serverMux.HandleFunc("/consume", controllers.Consume)
	serverMux.HandleFunc("/product", controllers.Product)
	err = http.Serve(l, serverMux)
	if err != nil {
		logrus.Fatal("failed to serve http server, error = %v", err)
	}
}
