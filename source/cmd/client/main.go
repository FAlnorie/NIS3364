package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"nis3364/utils"
)

type ChatClient struct {
	conn      net.Conn
	username  string
	OnMessage func(utils.Message)
	OnError   func(error)
}

func NewChatClient() *ChatClient {
	return &ChatClient{}
}

func (c *ChatClient) Connect(ip, port, username string) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", ip, port))
	if err != nil {
		return err
	}

	loginMsg := utils.Message{Type: "login", Content: username}
	encoded, _ := utils.JSONToBase64(loginMsg)
	conn.Write([]byte(encoded + "\n"))

	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		conn.Close()
		return fmt.Errorf("server closed connection")
	}

	var resp utils.Message
	if err := utils.Base64ToJSON(scanner.Text(), &resp); err != nil {
		conn.Close()
		return err
	}

	if resp.Type == "error" {
		conn.Close()
		return fmt.Errorf(resp.Content)
	}

	c.conn = conn
	c.username = username

	go c.listen(scanner)

	return nil
}

func (c *ChatClient) listen(scanner *bufio.Scanner) {
	defer c.conn.Close()
	for scanner.Scan() {
		var msg utils.Message
		if err := utils.Base64ToJSON(scanner.Text(), &msg); err != nil {
			log.Println("Decode error:", err)
			continue
		}
		if c.OnMessage != nil {
			c.OnMessage(msg)
		}
	}
	if c.OnError != nil {
		c.OnError(fmt.Errorf("connection lost"))
	}
}

func (c *ChatClient) Send(msg utils.Message) {
	if c.conn == nil {
		return
	}
	go func() {
		encoded, err := utils.JSONToBase64(msg)
		if err != nil {
			log.Println("Encode error:", err)
			return
		}
		c.conn.Write([]byte(encoded + "\n"))
	}()
}

func main() {
	gui := NewGUI()
	gui.Run()
}
