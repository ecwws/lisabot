package lisaclient

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ecwws/lisabot/logging"
	"io"
	"net"
	"time"
)

type LisaClient struct {
	raw      net.Conn
	decoder  *json.Decoder
	encoder  *json.Encoder
	SourceId string
	Logger   *logging.LisaLog
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

func RandomId() string {
	b := make([]byte, 8)
	io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%x", b)
}

func NewClient(host, port string,
	logger *logging.LisaLog) (*LisaClient, error) {

	lisabot := new(LisaClient)

	lisabot.Logger = logger

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
			Id:     RandomId(),
			Action: "engage",
			Type:   clienttype,
			Time:   time.Now().Unix(),
		},
	}

	lisa.SourceId = sourceid

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
				out <- &Query{
					Type:   "command",
					Source: "lisa",
					Command: &CommandBlock{
						Action: "disengage",
					},
				}
				break
			}
		}

		if lisa.ValidateQuery(q) {
			lisa.Logger.Debug.Println("Query received:", *q)
			if q.Type == "message" {
				lisa.Logger.Debug.Println("Message:", q.Message.Message)
				lisa.Logger.Debug.Println("From:", q.Message.From)
				lisa.Logger.Debug.Println("Room:", q.Message.Room)
			}
			out <- q
		} else {
			lisa.Logger.Error.Println("Invalid query from server(?!!):", q)
		}
	}
}

func (lisa *LisaClient) Run(toLisa <-chan *Query, fromLisa chan<- *Query) {

	go lisa.listen(fromLisa)
	var q *Query
	for {
		q = <-toLisa
		if lisa.ValidateQuery(q) {
			lisa.encoder.Encode(q)
		} else {
			lisa.Logger.Error.Println("Invalid query:", q)
		}
	}
}

func (lisa *LisaClient) ValidateQuery(q *Query) bool {
	switch {
	case q.Type == "command" && q.Command != nil:
		return true
	case q.Type == "message" && q.Message != nil && q.Message.Room != "":
		return true
	}
	return false
}
