package main

import (
	"strings"
)

type messageBlock struct {
	Message       string   `json:"message,omitempty"`
	From          string   `json:"from,omitempty"`
	Room          string   `json:"room,omitempty"`
	Mentioned     bool     `json:"mentioned,omitempty"`
	Stripped      string   `json:"stripped,omitempty"`
	MentionNotify []string `json:"mentionnotify,omitempty"`
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

		if !checkHelp(trimmed, source, m.Room, dispatch) {
			triggerPassiveResponders(prefixPResponders, trimmed, source, m.Room,
				m.From, false, dispatch)
		}
	} else {
		logger.Debug.Println("Non-prefix match triggered!")
		if !checkHelp(m.Message, source, m.Room, dispatch) {
			matched := triggerPassiveResponders(noPrefixPResponders, m.Message,
				source, m.Room, m.From, false, dispatch)
			if !matched && m.Mentioned {
				logger.Debug.Println("Mention match triggered!")
				triggerPassiveResponders(mentionPResponders, m.Stripped, source,
					m.Room, m.From, true, dispatch)
			}
		}
	}
}
