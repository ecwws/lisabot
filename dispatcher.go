package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
)

type dispatcherRequest struct {
	Query   *query
	Encoder *json.Encoder
}

func generateId() string {
	b := make([]byte, 8)
	io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%x", b)
}

func dispatcher(request <-chan *dispatcherRequest, quitChan chan int) {
	// inspect incoming request
	// if it's direct respond message, respond directly
	// if it's targeting specific connection id, patch to that connection
	// if it's operation to register pattern or command, perform registration

	connMap := make(map[string]*json.Encoder)

	for {
		req := <-request
		query := req.Query
		switch {
		case query.Type == "command" && query.Command != nil:
			cmd := query.Command
			switch {
			case cmd.Action == "engage" && cmd.Type == "adapter":
				if req.Encoder == nil {
					logerr.Println("No connection provided for adapter")
					panic("Dummy, you forgot to include connection data!")
				} else {
					id := query.Source
					if id == "" {
						id = generateId()
					}
					connMap[id] = req.Encoder
				}
			}
		case query.Type == "message" && query.Message != nil:
		default:
			logerr.Println("Malformed ", query.Type, " request, missing ",
				query.Type, " block. We shouldn't reach here, check your ",
				"query.validate() code")
		}
	}

	quitChan <- 1
}
