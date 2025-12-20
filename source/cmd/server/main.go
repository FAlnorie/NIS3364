package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"nis3364/utils"
	"strings"
)

type Client struct {
	User     string
	SendChan chan utils.Message
}

type RegisterReq struct {
	Client *Client
	Result chan bool
}

type Server struct {
	listener   net.Listener
	clients    map[string]*Client
	register   chan RegisterReq
	unregister chan *Client
	broadcast  chan utils.Message
}

func NewServer(port string) *Server {
	l, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	return &Server{
		listener:   l,
		clients:    make(map[string]*Client),
		register:   make(chan RegisterReq),
		unregister: make(chan *Client),
		broadcast:  make(chan utils.Message),
	}
}

func (s *Server) Run() {
	log.Printf("Server listening on %s", s.listener.Addr().String())
	defer s.listener.Close()

	go s.acceptLoop()

	for {
		select {
		case req := <-s.register:
			if _, exists := s.clients[req.Client.User]; exists {
				req.Result <- false
			} else {
				s.clients[req.Client.User] = req.Client
				req.Result <- true
				s.systemMessage(fmt.Sprintf("%s has joined the chat", req.Client.User))
			}

		case client := <-s.unregister:
			if _, ok := s.clients[client.User]; ok {
				delete(s.clients, client.User)
				close(client.SendChan)
				s.systemMessage(fmt.Sprintf("%s has left the chat", client.User))
			}

		case msg := <-s.broadcast:
			s.handleMessage(msg)
		}
	}
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go s.handleClient(conn)
	}
}

func (s *Server) handleMessage(msg utils.Message) {
	switch msg.Type {
	case "text":
		src, ok := s.clients[msg.Sender]
		if !ok {
			break
		}
		if msg.Sender == msg.Receiver {
			src.SendChan <- utils.Message{
				Type:    "error",
				Content: "Cannot send message to yourself.",
				Sender:  "System",
			}
			break
		}
		if dest, ok := s.clients[msg.Receiver]; ok {
			dest.SendChan <- msg
			acceptMsg := utils.Message{
				Type:    "accept",
				Content: fmt.Sprintf("Message \"%s\" delivered to user [%s].", msg.Content, msg.Receiver),
				Sender:  "System",
			}
			src.SendChan <- acceptMsg
		} else {
			src.SendChan <- utils.Message{
				Type:    "error",
				Content: fmt.Sprintf("User %s not found.", msg.Receiver),
				Sender:  "System",
			}
		}
	case "broadcast":
		for _, client := range s.clients {
			if client.User != msg.Sender {
				client.SendChan <- msg
			}
		}
		acceptMsg := utils.Message{
			Type:    "accept",
			Content: fmt.Sprintf("Broadcast message \"%s\" delivered.", msg.Content),
			Sender:  "System",
		}
		if src, ok := s.clients[msg.Sender]; ok {
			src.SendChan <- acceptMsg
		}
	case "ask":
		var users []string
		for user := range s.clients {
			users = append(users, user)
		}
		userList := "Current users: " + strings.Join(users, ", ")

		if src, ok := s.clients[msg.Sender]; ok {
			src.SendChan <- utils.Message{
				Type:    "text",
				Content: userList,
				Sender:  "System",
			}
		}
	}
}

func (s *Server) systemMessage(content string) {
	msg := utils.Message{
		Type:    "text",
		Content: content,
		Sender:  "System",
	}
	for _, client := range s.clients {
		client.SendChan <- msg
	}
}

func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()

	client, err := s.readLogin(conn)
	if err != nil {
		log.Printf("Handshake failed: %v", err)
		return
	}

	resChan := make(chan bool)
	s.register <- RegisterReq{Client: client, Result: resChan}
	if !<-resChan {
		errMsg := utils.Message{
			Type:    "error",
			Content: "Username already taken.",
			Sender:  "System",
		}
		encoded, _ := utils.JSONToBase64(errMsg)
		conn.Write([]byte(encoded + "\n"))
		return
	}

	welcome := utils.Message{
		Type:    "accept",
		Content: "Welcome " + client.User + "!",
		Sender:  "System",
	}
	encoded, _ := utils.JSONToBase64(welcome)
	conn.Write([]byte(encoded + "\n"))

	go s.clientWriter(conn, client.SendChan)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var msg utils.Message
		if err := utils.Base64ToJSON(scanner.Text(), &msg); err != nil {
			log.Printf("Decode error from %s: %v", client.User, err)
			continue
		}
		msg.Sender = client.User
		s.broadcast <- msg
	}

	s.unregister <- client
}

func (s *Server) readLogin(conn net.Conn) (*Client, error) {
	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		return nil, fmt.Errorf("connection closed during handshake")
	}

	var loginMsg utils.Message
	if err := utils.Base64ToJSON(scanner.Text(), &loginMsg); err != nil {
		return nil, err
	}

	if loginMsg.Type != "login" {
		return nil, fmt.Errorf("expected login message")
	}

	return &Client{
		User:     loginMsg.Content,
		SendChan: make(chan utils.Message, 16),
	}, nil
}

func (s *Server) clientWriter(conn net.Conn, ch <-chan utils.Message) {
	for msg := range ch {
		encoded, err := utils.JSONToBase64(msg)
		if err != nil {
			continue
		}
		conn.Write([]byte(encoded + "\n"))
	}
}

func main() {
	port := flag.String("port", ":8080", "server port")
	flag.Parse()
	if *port == "" {
		*port = ":8080"
	} else if !strings.HasPrefix(*port, ":") {
		*port = ":" + *port
	}
	srv := NewServer(*port)
	srv.Run()
}
