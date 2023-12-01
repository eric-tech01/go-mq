package main

import (
	"fmt"
	"time"

	json "github.com/eric-tech01/simple-json"
)

func main() {

	startServer()

	time.Sleep(1 * time.Second)

	go startSubClient()

	//循环发送
	startSendClient()
}

func startServer() {
	c := &Comm{
		Addr:           "127.0.0.1:8080",
		CertPath:       "../certs/",
		ServerKeyFile:  "../certs/server.key",
		ServerCertFile: "../certs/server.cer",
		CaFile:         "../certs/ca.cer",
	}
	c.Start()
}

func startSubClient() {
	addr := "127.0.0.1:8080"
	certs := "../certs/"
	commClient, err := startCommClient(addr, certs)
	if err != nil {
		fmt.Println(err)
		return
	}

	commClient.Subscribe("deviceFramework")

	err = recvFromMQ(commClient)
	if err != nil {
		fmt.Println(err)
	}
}

func startSendClient() {
	addr := "127.0.0.1:8080"
	certs := "../certs/"
	commClient, err := startCommClient(addr, certs)
	if err != nil {
		fmt.Println(err)
		return
	}

	msg := json.New()
	msg.Set("cmd", "hi")
	i := 0
	for {

		msg.Set("count", i)
		commClient.Publish("deviceFramework", msg)
		i++

		time.Sleep(2 * time.Second)
	}
}
