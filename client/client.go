package lisabot

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"
)

type LisabotClient struct {
	raw     net.Conn
	decoder *json.Decoder
	encoder *json.Encoder
}

type CommandBlock struct {
	Id      string   `json:"id,omitempty"`
	Action  string   `json:"action,omitempty"`
	Type    string   `json:"type,omitempty"`
	Time    int64    `json:"time,omitempty"`
	Data    string   `json:"data,omitempty"`
	Array   []string `json:"array,omitempty"`
	Options []string `json:"options,omitempty"`
}

type MessageBlock struct {
	Message string `json:"message,omitempty"`
	From    string `json:"from,omitempty"`
	Room    string `json:"room,omitempty"`
}

type Query struct {
	Type    string        `json:"type,omitempty"`
	Source  string        `json:"source,omitempty"`
	Command *commandBlock `json:"command,omitempty"`
	Message *messageBlock `json:"message,omitempty"`
}

func Id() string {
	b := make([]byte, 8)
	io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%x", b)
}

func NewLisabotClient(host, port string) (*LisabotClient, error) {
	lisabot := new(LisabotClient)

	conn, err := net.Dial("tcp", host+":"+port)

	if err != nil {
		return lisabot, err
	}

	lisabot.raw = conn
	lisabot.decoder = json.NewDecoder(lisabot.raw)
	lisabot.encoder = json.NewEncoder(lisabot.raw)

	return lisabot, nil
}

func (lisa *LisabotClient) Engage() error {
	q := Query{
		Type:   "command",
		Source: "lisabot-hipchat",
		Command: &commandBlock{
			Id:     Id(),
			Action: "engage",
			Type:   "adapter",
			Time:   time.Now().Unix(),
		},
	}

	return lisa.encoder.Encode(&q)
}
