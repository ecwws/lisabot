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
	isAdapter := make(map[string]bool) // not adapter then responder

	for {
		req := <-request
		q := req.Query

		if err := q.validate(); err != nil {
			logger.Error.Println("Query failed to validate:", err)
			logger.Info.Println("Invalid query received:", q)
			continue
		}

		switch {
		case q.Type == "command":
			cmd := q.Command
			switch cmd.Action {
			case "engage":
				if req.Encoder == nil {
					logger.Error.Println(
						"No connection provided for engagement")
					logger.Error.Fatal("Bad code, check code ininitialize()")
				} else {
					if err := cmd.engageChk(q.Source, conf.Secret); err == nil {
						id := q.Source
						// no source identifier given, we'll use a random
						// source id
						if id == "" {
							id = generateId()
						}

						// source identifier collision, use a random source id
						// and keep generating until no collision is found
						for _, ok := connMap[id]; ok; _, ok = connMap[id] {
							id = generateId()
						}

						connMap[id] = req.Encoder

						if cmd.Type == "adapter" {
							isAdapter[id] = true
						} else {
							isAdapter[id] = false
						}

						if id != q.Source && q.Source != "" {
							logger.Warn.Println("Requester's source id already",
								"taken, assign new source ID: ", q.Source,
								"-->", id)
						}

						logger.Info.Println("Engagement accepted: ", id)
						req.EngageResp <- id
						close(req.EngageResp)

						req.Encoder.Encode(&query{
							Type:   "command",
							Source: "server",
							To:     id,
							Command: &commandBlock{
								Id:     generateId(),
								Action: "proceed",
								Data:   id,
							},
						})
					} else {
						logger.Error.Println("unable to valide engagement", err)
						req.EngageResp <- ""
						close(req.EngageResp)

						req.Encoder.Encode(&query{
							Type:   "command",
							Source: "server",
							To:     q.Source,
							Command: &commandBlock{
								Id:     generateId(),
								Action: "terminate",
								Data:   err.Error(),
							},
						})
					}
				}
			case "disengage":
				if q.Source != "" {
					delete(connMap, q.Source)
					delete(isAdapter, q.Source)
				}
				logger.Info.Println("Connection disengaged: ", q.Source)
			default:
				go cmd.handleCommand(q.Source, request)
			}
		case q.Type == "message":
			// message from an adapter won't have a "To" field
			if q.To != "" && q.To != "server" {
				logger.Debug.Println("Responder message received:", *q.Message)
				logger.Debug.Println("Query source:", q.Source)
				if encoder, ok := connMap[q.To]; ok {
					encoder.Encode(q)
				} else {
					logger.Error.Println("Cannot find adapter source for", q.To)
				}
			} else {
				logger.Debug.Println("Adapter message received:", *q.Message)
				go q.Message.handleMessage(q.Source, request)
			}
		default:
			logger.Error.Println("Unhandlabe message, bad client code")
		}
	}

	quitChan <- true
}
