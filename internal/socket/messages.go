package socket

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type ContentType string

const (
	ContentType_BroadcastMessage ContentType = "BoradcastMessage"
	ContentType_PrivateMessage   ContentType = "PrivateMessage"
)

type Message interface {
	ToJson() ([]byte, error)
}

type ClientMessage struct {
	ContentType ContentType            `json:"contentType"`
	Payload     map[string]interface{} `json:"payload"`
}

func (c *ClientMessage) parsePayload(s interface{}) error {
	stValue := reflect.ValueOf(s).Elem()
	sType := stValue.Type()
	for i := 0; i < sType.NumField(); i++ {
		field := sType.Field(i)
		if value, ok := c.Payload[field.Name]; ok {
			stValue.Field(i).Set(reflect.ValueOf(value))
		} else {
			return fmt.Errorf("failed to find property '%s'", field.Name)
		}
	}

	return nil
}

type BroadcastMessge struct {
	ContentType ContentType `json:"contentType"`
	SessionId   string      `json:"sessionId"`
	Content     string      `json:"content"`
	FromUser    string      `json:"fromUser"`
	Target      string      `json:"target"`
}

func (b *BroadcastMessge) ToJson() ([]byte, error) {
	return json.Marshal(b)
}

func NewBroadcastMessage(session string, content string, target string, fromUser string) BroadcastMessge {
	return BroadcastMessge{ContentType: ContentType_BroadcastMessage, SessionId: session, Content: content, Target: target, FromUser: fromUser}
}

type PrivateMessage struct {
	ContentType ContentType `json:"contentType"`
	Message     string      `json:"message"`
	SessionId   string      `json:"sessionId"`
	ToUser      string      `json:"toUser"`
	FromUser    string      `json:"fromUser"`
}

func (b *PrivateMessage) ToJson() ([]byte, error) {
	return json.Marshal(b)
}

func NewPrivateMessage(fromUser string, toUser string, session string, msg string) PrivateMessage {
	return PrivateMessage{ContentType: ContentType_PrivateMessage, Message: msg, FromUser: fromUser, ToUser: toUser, SessionId: session}
}
