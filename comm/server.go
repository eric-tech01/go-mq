package comm

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"time"

	json "github.com/eric-tech01/simple-json"
)

// StartServer 启动comm服务
func (comm *Comm) StartServer(addr string, key, cert, ca []byte) {
	fmt.Printf("comm[%p].StartServer(%v)", comm, addr)
	serverID := ""
	server, err := listen(comm, serverID, addr, key, cert, ca)
	if err != nil {
		fmt.Printf("Listen error: %v", err)
		comm.onStartServer(false, fmt.Sprintf("Listen error: %v", err))
		return
	}
	comm.server = server
	fmt.Printf("comm[%p].StartServer() success", comm)
	comm.onStartServer(true, "")
}

// onStartServer
func (comm *Comm) onStartServer(success bool, msg string) {
	fmt.Printf("comm[%p].onStartServer(%v,%v)", comm, success, msg)
	m := json.New()
	m.Set("cmd", "onStartServer")
	m.Set("success", success)
	m.Set("msg", msg)
	comm.input(m)
}

// StopServer 停止comm服务
func (comm *Comm) StopServer() {
	fmt.Printf("comm[%p].StopServer()", comm)
	if comm.server != nil {
		comm.server.Close()
	}
}

func (comm *Comm) IsServerAlive() bool {
	return comm.server != nil
}

// dispatchTopic comm根据收到的topic分发给订阅的client
func (comm *Comm) dispatchTopic(topic string, m *json.Json) {
	fmt.Printf("comm[%p].dispatchTopic(%v) %p", comm, topic, comm.server)
	for _, peer := range comm.server.peers {
		if peer.topics.match(topic) {
			fmt.Printf("dispatchTopic(%v, %v) to peer: %p", topic, m, peer)
			peer.send(m)
		}
	}
}

// commServer comm服务端
type commServer struct {
	id       string
	listener net.Listener
	peers    []*commClient
	// cloud    *commCloud
	comm *Comm
}

// listen comm服务端启动监听
func listen(comm *Comm, id, addr string, key, cert, ca []byte) (*commServer, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		fmt.Printf("ResolveTCPAddr(%v) error: %v", addr, err)
		return nil, err
	}
	// pool
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(ca)
	// serverCert
	serverCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		fmt.Printf("X509KeyPair error: %v", err)
		return nil, err
	}
	// tlsConfig
	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		ClientCAs:    pool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{serverCert},
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Printf("ListenTCP() error: %v", err)
		return nil, err
	}
	fmt.Printf("listen success")
	// commServer
	server := &commServer{
		listener: listener,
		id:       id,
		comm:     comm,
	}
	// accept
	go func() {
		for {
			tcpConn, err := listener.Accept()
			if err != nil {
				fmt.Printf("accept error: %v", err)
				return
			}
			// keepalive
			setKeepAlive(tcpConn, 10*time.Second)
			// setKeepAlive(tcpConn.(*net.TCPConn), 10, 3, 10)
			conn := tls.Server(tcpConn, tlsConfig)
			// handshake
			err = conn.Handshake()
			if err != nil {
				fmt.Printf("server handshake error: %v", err)
				continue
			}
			peer := &commClient{
				conn:   conn,
				server: server,
				topics: &commTopics{
					topics: make(map[string]interface{}, 16),
				},
			}
			server.peers = append(server.peers, peer)
			fmt.Printf("comm[%p] accepted peer[%p]", comm, peer)
			peer.start(comm)
		}
	}()
	return server, nil
}

// Close 关闭comm服务端
func (s *commServer) Close() {
	s.listener.Close()
	for _, c := range s.peers {
		c.close()
	}
}
