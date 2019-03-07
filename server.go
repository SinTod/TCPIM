package main

import (
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"time"
)

// Server server
type Server struct {
	IP   string
	Port int
	// 在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的 channel
	Message chan string
}

// NewServer 创建一个 server
func NewServer(ip string, port int) *Server {
	server := &Server{
		IP:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// BroadCast 广播消息方法
func (s *Server) BroadCast(user *User, msg string) {
	broadMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	s.Message <- broadMsg

}

// MessageListener 消息监听
func (s *Server) MessageListener() {
	for {
		msg := <-s.Message
		s.mapLock.Lock()
		for _, user := range s.OnlineMap {
			user.C <- msg
		}
		s.mapLock.Unlock()
	}
}

// Handler 处理请求
func (s *Server) Handler(conn net.Conn) {
	// 用户上线 将用户加入到 onlineMap
	user := NewUser(conn, s)
	user.Online()

	isActive := make(chan bool)

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("read err", err)
				return
			}
			msg := string(buf[:n-1])

			user.SendMessage(msg)
			isActive <- true

		}
	}()

	// 超时踢下线
	for {
		select {
		case <-isActive:
			// 重置定时器
		case <-time.After(300 * time.Second):
			fmt.Println(user.Name + ",you are timeout")
			user.SendPrivateMessage("you are timeout\n")
			user.Offline()
			close(user.C)
			user.conn.Close()
			runtime.Goexit()
		}
	}

}

// Run 启动 server
func (s *Server) Run() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		fmt.Println("net.Listen err", err)
		return
	}
	// close socket
	defer listener.Close()

	// 启动 MessageListener
	go s.MessageListener()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept", err)
			continue
		}

		// do handler
		go s.Handler(conn)

	}

}
