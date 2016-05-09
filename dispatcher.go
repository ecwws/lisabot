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

func dispatcher(request <-chan *dispatcherRequest, quitChan chan bool) {
	// inspect incoming request
	// if it's direct respond message, respond directly
	// if it's targeting specific connection id, patch to that connection
	// if it's operation to register pattern or command, perform registration

	connMap := make(map[string]*json.Encoder)

	for {
		req := <-request
		query := req.Query
		switch {
		case query.Type == "command":
			cmd := query.Command
			switch cmd.Action {
			case "engage":
				if req.Encoder == nil {
					logger.Error.Println("No connection provided for adapter")
					panic("Dummy, you forgot to include connection data!")
				} else {
					id := query.Source

					if id == "" {
						id = generateId()
					}

					for _, ok := connMap[id]; ok; _, ok = connMap[id] {
						id = generateId()
					}

					connMap[id] = req.Encoder

					if id != query.Source && query.Source != "" {
						logger.Warn.Println("Requester's source id already",
							"taken, assign new source ID: ", query.Source,
							"-->", id)
					}

					logger.Info.Println("Engagement accepted: ", id)
					req.EngageResp <- id
					close(req.EngageResp)
				}
			case "disengage":
				if query.Source != "" {
					delete(connMap, query.Source)
				}
				logger.Info.Println("Connection disengaged: ", query.Source)
			default:
				cmd.handleCommand(query.Source)
			}
		case query.Type == "message":
		default:
			logger.Error.Println("Unhandlabe message, we shouldn't get here")
		}
	}

	quitChan <- true
}
