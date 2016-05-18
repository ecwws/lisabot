package main

import (
	"os/exec"
	"regexp"
	"strings"
)

type messageBlock struct {
	Message string `json:"message"`
	From    string `json:"from"`
	Room    string `json:"room"`
	To      string
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
		for _, pr := range prefixPResponders {
			if match, err := regexp.MatchString(pr.Match,
				trimmed); err == nil && match {

				logger.Debug.Println("Matched:", pr.Match)

				output, err := exec.Command(pr.Cmd, pr.Args...).Output()

				if err != nil {
					logger.Error.Println("Passive responder error:", err)
				} else {
					logger.Debug.Println("Passive responder executed:",
						string(output))

					dispatch <- &dispatcherRequest{
						Query: &query{
							Type:   "message",
							Source: "PR:" + pr.Match,
							Message: &messageBlock{
								Message: strings.TrimRight(string(output), "\n"),
								Room:    m.Room,
								To:      source,
							},
						},
					}
				}
			}
		}
	} else {
		// for _, pr := range noPrefixPResponders {
		// }
	}
}
