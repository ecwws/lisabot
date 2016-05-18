package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
)

type dispatcherRequest struct {
	Query      *query
	Encoder    *json.Encoder
	EngageResp chan<- string
}

func generateId() string {
	b := make([]byte, 8)
	io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%x", b)
}

func dispatcher(request chan *dispatcherRequest, quitChan chan bool) {
	// inspect incoming request
	// if it's direct respond message, respond directly
	// if it's targeting specific connection id, patch to that connection
	// if it's operation to register pattern or command, perform registration

	connMap := make(map[string]*json.Encoder)

	for {
		req := <-request
		q := req.Query

		if !q.validate() {
			logger.Error.Println("Invalid query received:", q)
			continue
		}

		switch {
		case q.Type == "command":
			cmd := q.Command
			switch cmd.Action {
			case "engage":
				if req.Encoder == nil {
					logger.Error.Println("No connection provided for adapter")
					panic("Dummy, you forgot to include connection data!")
				} else {
					id := q.Source

					// no source identifier given, we'll use a random source id
					if id == "" {
						id = generateId()
					}

					// source identifier collision, use a random source id
					for _, ok := connMap[id]; ok; _, ok = connMap[id] {
						id = generateId()
					}

					connMap[id] = req.Encoder

					if id != q.Source && q.Source != "" {
						logger.Warn.Println("Requester's source id already",
							"taken, assign new source ID: ", q.Source,
							"-->", id)
					}

					logger.Info.Println("Engagement accepted: ", id)
					req.EngageResp <- id
					close(req.EngageResp)
				}
			case "disengage":
				if q.Source != "" {
					delete(connMap, q.Source)
				}
				logger.Info.Println("Connection disengaged: ", q.Source)
			default:
				go cmd.handleCommand(q.Source, request)
			}
		case q.Type == "message":
			// message from an adapter won't have a "To" field
			if q.Message.To != "" {
				logger.Debug.Println("Responder message received:", *q.Message)
				if encoder, ok := connMap[q.Message.To]; ok {
					encoder.Encode(q)
				} else {
					logger.Error.Println("Cannot find adapter source for",
						q.Message.To)
				}
			} else {
				logger.Debug.Println("Adapter message received:", *q.Message)
				go q.Message.handleMessage(q.Source, request)
			}
		default:
			logger.Error.Println("Unhandlabe message, we shouldn't get here")
		}
	}

	quitChan <- true
}