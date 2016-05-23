package main

import (
	"strings"
)

type messageBlock struct {
	Message   string `json:"message,omitempty"`
	From      string `json:"from,omitempty"`
	Room      string `json:"room,omitempty"`
	Mentioned bool   `json:"mentioned,omitempty"`
	Stripped  string `json:"stripped,omitempty"`
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
		logger.Debug.Println("Prefix matched!")
		trimmed := strings.TrimLeft(m.Message[conf.prefixLen:], " ")
		triggerPassiveResponders(prefixPResponders, trimmed, source, m.Room,
			m.Stripped, m.Mentioned, dispatch)
	} else {
		logger.Debug.Println("No prefix match triggered!")
		triggerPassiveResponders(noPrefixPResponders, m.Message, source, m.Room,
			m.Stripped, m.Mentioned, dispatch)
	}
}
