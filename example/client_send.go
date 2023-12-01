package main

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/eric-tech01/go-mq/comm"
)

func startCommClient(commAddr, certsPath string) (*comm.Comm, error) {
	// 读取证书
	caCert, err := ioutil.ReadFile(certsPath + "ca.cer") // TODO comm封装一个key, cert, ca, err := ReadCerts(certPath)
	if err != nil {
		return nil, err
	}
	clientKey, err := ioutil.ReadFile(certsPath + "client.key")
	if err != nil {
		return nil, err
	}
	clientCert, err := ioutil.ReadFile(certsPath + "client.cer")
	if err != nil {
		return nil, err
	}
	// 初始化commClient
	commClient := comm.NewComm()
	// 因为device是系统服务，随机启动时可能comm还没启动好，需要重连
	fmt.Printf("====startCommClient")
	err = commClient.StartClient(commAddr, clientKey, clientCert, caCert)
	if err != nil {
		fmt.Printf("comm.StartClient error: %v", err)
		return nil, errors.New("StartClient failed")
	}
	fmt.Printf("comm.StartClient success")
	// 接收onClientStart
	m := commClient.Recv()
	if m == nil || m.Get("cmd").MustString() != "onStartClient" || !m.Get("success").MustBool() {
		fmt.Printf("recv mq.onStartClient error: %v", m)
		return commClient, errors.New(fmt.Sprintf("recv mq.onStartClient error: %v", m))
	}
	fmt.Printf("recv mq.onClientStart success")
	return commClient, nil
}

// 从MQ接收消息，并处理
func recvFromMQ(commClient *comm.Comm) error {
	fmt.Printf("start recv from MQ")
	for {
		m := commClient.Recv()
		if m == nil {
			fmt.Printf("ERROR device error recv nil")
			break
		}
		fmt.Printf("device recv mq msg: %v", m)
	}
	return nil
}
