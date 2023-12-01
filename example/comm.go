package main

// 通讯库模块

import (
	"fmt"
	"io/ioutil"

	"github.com/eric-tech01/go-mq/comm"
	json "github.com/eric-tech01/simple-json"
)

// 通讯库模块
type Comm struct {
	Addr           string
	CertPath       string
	ServerKeyFile  string
	ServerCertFile string
	CaFile         string
	*comm.Comm
}

// Start 启动通讯库实例
func (c *Comm) Start() error {

	fmt.Printf("comm.Start() %v", c)
	// comm server
	serverKey, serverCert, caCert, err := readCerts(c.ServerKeyFile, c.ServerCertFile, c.CaFile)
	if err != nil {
		return err
	}
	// 新建comm
	c.Comm = comm.NewComm()
	// 启动comm
	c.StartServer(c.Addr, serverKey, serverCert, caCert)
	// comm client
	clientKey, clientCert, _, err := readCerts(c.CertPath+"/client.key", c.CertPath+"/client.cer", c.CertPath+"/ca.cer")
	if err != nil {
		return err
	}
	c.StartClient(c.Addr, clientKey, clientCert, caCert)
	// comm client subscribe
	c.Subscribe("base.sn")
	c.Subscribe("base.serialPort.list")
	// go c.clientRecv()
	return nil
}

const commConfigPath = "Comm"

func (c *Comm) SendMqMsg(topic string, m *json.Json) {
	if c.Comm != nil {
		c.Comm.Publish(topic, m)
	}
}

func readCerts(keyFile, certFile, caFile string) ([]byte, []byte, []byte, error) {
	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, nil, nil, err
	}
	cert, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, nil, nil, err
	}
	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, nil, nil, err
	}
	return key, cert, ca, nil
}
