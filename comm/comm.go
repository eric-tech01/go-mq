package comm

import (
	"fmt"
	"os/exec"
	"time"

	json "github.com/eric-tech01/simple-json"
)

// Comm 通讯库
type Comm struct {
	server   *commServer
	client   *commClient    // 包含commCloud
	outChan  chan json.Json // 业务通过recv接口获取message的channel
	callback MessageCallback
}

// MessageCallback 消息回调接口
type MessageCallback interface {
	MessageCallback(*json.Json)
}

func init() {
	_, err := execShell("sysctl -w net.ipv4.tcp_retries2=5")
	if err != nil {
		fmt.Print(err)
		return
	}
	fmt.Printf("set tcp_retries success")
}
func execShell(cmd string, args ...string) (string, error) {
	c := exec.Command("sh", "-c", cmd)
	output, err := c.CombinedOutput()
	fmt.Printf("exec command %v: output: %v, err: %v", cmd, string(output), err)
	return string(output), err
}

// NewComm 创建新的通讯库实例
func NewComm() *Comm {
	return &Comm{
		outChan: make(chan json.Json, 128),
	}
}

// input 内部接收到消息，转发给调用方
func (comm *Comm) input(msg *json.Json) {
	fmt.Printf("comm[%p].input() %v", comm, msg)
	select {
	case comm.outChan <- *msg:
	default:
		fmt.Printf("comm.input chan full")
	}
}

// Recv 调用方接收消息。用于单元测试调试。
func (comm *Comm) Recv() *json.Json {
	return comm.RecvWithTimeout(-1)
}

func (comm *Comm) TryRecv() *json.Json {
	return comm.RecvWithTimeout(0)
}

func (comm *Comm) RecvWithTimeout(timeout time.Duration) *json.Json {
	if timeout == -1 {
		select {
		case msg := <-comm.outChan:
			return &msg
		}
	} else {
		select {
		case msg := <-comm.outChan:
			return &msg
		case <-time.After(timeout):
			fmt.Printf("comm.Recv() timeout")
			return nil
		}
	}
}

// RegisterMessageCallback 注册comm消息回调
func (comm *Comm) RegisterMessageCallback(cb MessageCallback) {
	comm.callback = cb
	go func() {
		for m := range comm.outChan {
			cb.MessageCallback(&m)
		}
	}()
}
