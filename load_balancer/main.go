package main

import (
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"log/slog"
)

var activeRoutines atomic.Int32

type connection struct {
	conn    net.Conn
	engaged atomic.Bool
}

type server struct {
	addr        string
	connections []*connection
	mu          sync.Mutex
}

func (s *server) getConnection() *connection {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, conn := range s.connections {
		if conn.engaged.CompareAndSwap(false, true) {
			return conn
		}
	}
	newConn := &connection{}
	newConn.engaged.Store(true)
	s.connections = append(s.connections, newConn)
	return newConn
}

var (
	backendServers = []server{
		{addr: "127.0.0.1:8081"},
		{addr: "127.0.0.1:8082"},
		{addr: "127.0.0.1:8083"},
		{addr: "127.0.0.1:8084"},
	}
	currentServer uint32
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		slog.Error("Error starting TCP server", "err", err)
	}
	defer listener.Close()

	slog.Info("Load balancer started on :8080")

	go func() {
		for {
			slog.Info("Active routines", "count", activeRoutines.Load())
			time.Sleep(5 * time.Second)
		}
	}()

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			slog.Error("Error accepting connection", "err", err)
			continue
		}

		go handleClient(clientConn)
	}
}

func handleClient(clientConn net.Conn) {
	defer activeRoutines.Add(-1)
	defer clientConn.Close()

	activeRoutines.Add(1)

	server := getNextBackend()
	conn := server.getConnection()
	var err error
	if conn.conn == nil {
		slog.Info("creating connection", "addr", server.addr)
		conn.conn, err = net.Dial("tcp", server.addr)
		if err != nil {
			slog.Error("Error connecting to backend", "addr", server.addr, "err", err)
			conn.engaged.Store(false)
			return
		}
	}
	defer conn.engaged.Store(false)

	slog.Info("Forwarding connection to backend", "addr", server.addr)
	done := make(chan struct{})

	// FIXME: Cannot re-use the same connection for second request.
	go func() {
		_, _ = io.Copy(conn.conn, clientConn)
		slog.Info("request done", "addr", server.addr)
		done <- struct{}{}
	}()

	go func() {
		_, _ = io.Copy(clientConn, conn.conn)
		slog.Info("response done", "addr", server.addr)
		done <- struct{}{}
	}()

	<-done
	<-done
}

func getNextBackend() *server {
	serverIndex := atomic.AddUint32(&currentServer, 1) % uint32(len(backendServers))
	return &backendServers[serverIndex]
}
