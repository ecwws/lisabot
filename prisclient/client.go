package prisclient

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/priscillachat/logging"
	"io"
	"net"
	"time"
)

type Client struct {
	raw      net.Conn
	decoder  *json.Decoder
	encoder  *json.Encoder
	SourceId string
	Logger   *logging.PrisLogger
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
	Message   string `json:"message,omitempty"`
	From      string `json:"from,omitempty"`
	Room      string `json:"room,omitempty"`
	Mentioned bool   `json:"mentioned,omitempty"`
}

type Query struct {
	Type    string        `json:"type,omitempty"`
	Source  string        `json:"source,omitempty"`
	To      string        `json:"to,omitempty"`
	Command *CommandBlock `json:"command,omitempty"`
	Message *MessageBlock `json:"message,omitempty"`
}

func RandomId() string {
	b := make([]byte, 8)
	io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%x", b)
}

func NewClient(host, port string,
	logger *logging.PrisLog) (*Client, error) {

	pris := new(Client)

	pris.Logger = logger

	conn, err := net.Dial("tcp", host+":"+port)

	if err != nil {
		return pris, err
	}

	pris.raw = conn
	pris.decoder = json.NewDecoder(pris.raw)
	pris.encoder = json.NewEncoder(pris.raw)

	return pris, nil
}

func (pris *Client) Engage(clienttype, sourceid string) error {
	if clienttype != "adapter" && clienttype != "responder" {
		return errors.New("client type has to be adapter or responder")
	}

	q := Query{
		Type:   "command",
		Source: sourceid,
		Command: &CommandBlock{
			Id:     RandomId(),
			Action: "engage",
			Type:   clienttype,
			Time:   time.Now().Unix(),
		},
	}

	pris.SourceId = sourceid

	return pris.encoder.Encode(&q)
}

func (pris *Client) listen(out chan<- *Query) {
	var q *Query
	for {
		q = new(Query)
		err := pris.decoder.Decode(q)

		if err != nil {
			fmt.Println(err)
			if err.Error() == "EOF" {
				out <- &Query{
					Type:   "command",
					Source: "pris",
					Command: &CommandBlock{
						Action: "disengage",
					},
				}
				break
			}
		}

		if pris.ValidateQuery(q) {
			pris.Logger.Debug.Println("Query received:", *q)
			if q.Type == "message" {
				pris.Logger.Debug.Println("Message:", q.Message.Message)
				pris.Logger.Debug.Println("From:", q.Message.From)
				pris.Logger.Debug.Println("Room:", q.Message.Room)
			}
			out <- q
		} else {
			pris.Logger.Error.Println("Invalid query from server(?!!):", q)
		}
	}
}

func (pris *Client) Run(toPris <-chan *Query, fromPris chan<- *Query) {

	go pris.listen(fromPris)
	var q *Query
	for {
		q = <-toPris
		if pris.ValidateQuery(q) {
			pris.encoder.Encode(q)
		} else {
			pris.Logger.Error.Println("Invalid query:", q)
		}
	}
}

func (pris *Client) ValidateQuery(q *Query) bool {
	switch {
	case q.Type == "command" && q.Command != nil:
		return true
	case q.Type == "message" && q.Message != nil && q.Message.Room != "":
		return true
	}
	return false
}
