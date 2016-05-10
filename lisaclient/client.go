package lisaclient

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

type LisaClient struct {
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
	Command *CommandBlock `json:"command,omitempty"`
	Message *MessageBlock `json:"message,omitempty"`
}

func Id() string {
	b := make([]byte, 8)
	io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%x", b)
}

func NewClient(host, port string) (*LisaClient, error) {
	lisabot := new(LisaClient)

	conn, err := net.Dial("tcp", host+":"+port)

	if err != nil {
		return lisabot, err
	}

	lisabot.raw = conn
	lisabot.decoder = json.NewDecoder(lisabot.raw)
	lisabot.encoder = json.NewEncoder(lisabot.raw)

	return lisabot, nil
}

func (lisa *LisaClient) Engage(clienttype, sourceid string) error {
	if clienttype != "adapter" && clienttype != "responder" {
		return errors.New("client type has to be adapter or responder")
	}

	q := Query{
		Type:   "command",
		Source: sourceid,
		Command: &CommandBlock{
			Id:     Id(),
			Action: "engage",
			Type:   clienttype,
			Time:   time.Now().Unix(),
		},
	}

	return lisa.encoder.Encode(&q)
}

func (lisa *LisaClient) listen(out chan<- *Query) {
	var q *Query
	for {
		q = new(Query)
		err := lisa.decoder.Decode(q)

		if err != nil {
			fmt.Println(err)
			if err.Error() == "EOF" {
				break
			}
		}

		out <- q
	}
}

func (lisa *LisaClient) Run(toLisa <-chan *Query, fromLisa chan<- *Query) {

	go lisa.listen(fromLisa)
	var q *Query
	for {
		q = <-toLisa
		lisa.encoder.Encode(&q)
	}
}
