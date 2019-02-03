package main

import (
	"container/list"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
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

func removeSource(arl *list.List, source string) {
	for eAr := arl.Front(); eAr != nil; eAr = eAr.Next() {
		ar := eAr.Value.(*activeResponderConfig)
		logger.Debug.Println("Remove check, source:", ar.source)
		if ar.source == source {
			rAr := eAr
			eAr = eAr.Prev()
			logger.Debug.Println("Deregistering active responder:", ar.helpCmd)
			arl.Remove(rAr)
			if eAr == nil {
				break
			}
		}
	}
}

func deregister(source string) {
	logger.Debug.Println("Deregister started for:", source)
	removeSource(prefixAResponders, source)
	removeSource(noPrefixAResponders, source)
	removeSource(mentionAResponders, source)
	removeSource(unhandledAResponders, source)
}

func dispatcher(request chan *dispatcherRequest, quitChan chan bool) {
	// inspect incoming request
	// if it's direct respond message, respond directly
	// if it's targeting specific connection id, patch to that connection
	// if it's operation to register pattern or command, perform registration

	connMap := make(map[string]*json.Encoder)
	var adapters []*json.Encoder

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
							adapters = append(adapters, req.Encoder)
							logger.Debug.Println("New adapter registered:", id)
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
								Action: "proceed",
								Data:   id,
							},
						})
					} else {
						logger.Error.Println("Invalid engagement request", err)
						req.EngageResp <- ""
						close(req.EngageResp)

						req.Encoder.Encode(&query{
							Type:   "command",
							Source: "server",
							To:     q.Source,
							Command: &commandBlock{
								Action: "terminate",
								Data:   err.Error(),
							},
						})
					}
				}
			case "disengage":
				if q.Source != "" {
					delete(connMap, q.Source)
				}
				logger.Info.Println("Connection disengaged: ", q.Source)
				deregister(q.Source)
			case "register":
				logger.Debug.Println("Register command received:", cmd)
				if err := cmd.registerChk(); err == nil {
					ar := new(activeResponderConfig)
					ar.regex, err = regexp.Compile(cmd.Data)
					if err != nil {
						logger.Error.Println("Error compiling regex:", err)
						continue
					}
					ar.source = q.Source
					ar.id = cmd.Id
					ar.helpCmd = cmd.Array[0]
					ar.help = cmd.Array[1]
					for _, option := range cmd.Options {
						if option == "fallthrough" {
							ar.matchNext = true
						}
					}

					helpMsg := &helpInfo{
						helpCmd: ar.helpCmd,
						helpMsg: ar.help,
					}

					switch cmd.Type {
					case "prefix":
						help.PushBack(helpMsg)
						prefixAResponders.PushBack(ar)
					case "noprefix":
						helpMsg.noPrefix = true
						help.PushBack(helpMsg)
						noPrefixAResponders.PushBack(ar)
					case "mention":
						helpMsg.mention = true
						help.PushBack(helpMsg)
						mentionAResponders.PushBack(ar)
					case "unhandled":
						unhandledAResponders.PushBack(ar)
					}
					logger.Debug.Println("Active adapter registered:", ar)
				} else {
					logger.Error.Println("Invalid register command:", err)
				}
			case "user_request":
				fallthrough
			case "room_request":
				fallthrough
			case "info":
				if q.To != "" && q.To != "server" {
					logger.Debug.Println("Info command received from:",
						q.Source)
					logger.Debug.Println("Info command destined to:", q.To)
					logger.Debug.Println("Action:", cmd.Action)

					if encoder, ok := connMap[q.To]; ok {
						encoder.Encode(q)
					} else {
						logger.Error.Println("Destination doesn't exist:", q.To)
					}
				} else {
					logger.Error.Println("Missing destination for info request")
				}
			default:
				go cmd.handleCommand(q.Source, request)
			}
		case q.Type == "message":
			// message from an adapter won't have a "To" field
			if q.To != "" && q.To != "server" {
				logger.Debug.Println("Responder message received:", *q.Message)
				logger.Debug.Println("Query source:", q.Source)

				// some debug log
				debug, _ := json.Marshal(q)
				logger.Debug.Println("Query to be sent:", string(debug))

				// first check if it's a broadcast message
				if q.To == "*" {
					logger.Info.Println("Broadcast message received:", q)
					for _, adapterEncoder := range adapters {
						adapterEncoder.Encode(q)
					}
				} else if encoder, ok := connMap[q.To]; ok {
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
