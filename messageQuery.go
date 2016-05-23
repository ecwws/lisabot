package main

import (
	"strings"
)

type messageBlock struct {
	Message   string `json:"message,omitempty"`
	From      string `json:"from,omitempty"`
	Room      string `json:"room,omitempty"`
	Mentioned bool   `json:"mentioned,omitempty"`
	Mention   string `json:"mention,omitempty"`
}

func (m *messageBlock) handleMessage(source string,
	dispatch chan<- *dispatcherRequest) {

	logger.Debug.Println("Message: ", m.Message)
	logger.Debug.Println("From: ", m.From)
	logger.Debug.Println("Room: ", m.Room)

	prefixMatch := false

	if len(m.Message) > conf.prefixLen &&
		m.Message[0:conf.prefixLen] == conf.Prefix {

		prefixMatch = true
	}

	if prefixMatch {
		trimmed := strings.TrimLeft(m.Message[conf.prefixLen:], " ")
		triggerPassiveResponders(prefixPResponders, trimmed, source, m.Room,
			m.Mention, m.Mentioned, dispatch)
	} else {
		triggerPassiveResponders(noPrefixPResponders, m.Message, source, m.Room,
			m.Mention, m.Mentioned, dispatch)
	}
}
