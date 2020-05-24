package main

import "fmt"

type F1 func(interface{})

func main() {
	var f F1 = func(i string) {
		fmt.Println("xxx")
	}
}
