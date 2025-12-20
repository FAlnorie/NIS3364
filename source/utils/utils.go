package utils

import (
	"encoding/base64"
	"encoding/json"
)

type MessageType string

const (
	LoginMessage     MessageType = "login"
	TextMessage      MessageType = "text"
	BroadcastMessage MessageType = "broadcast"
	ErrorMessage     MessageType = "error"
	AcceptMessage    MessageType = "accept"
	AskMessage       MessageType = "ask"
	LogoutMessage    MessageType = "logout"
)

type Message struct {
	Type     string `json:"type"`
	Content  string `json:"content"`
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
}

func Base64ToJSON(b64 string, v interface{}) error {
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func JSONToBase64(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}
