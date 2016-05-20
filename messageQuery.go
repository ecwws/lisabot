package main

import (
	"os/exec"
	"regexp"
	"strings"
)

type messageBlock struct {
	Message   string `json:"message,omitempty"`
	From      string `json:"from,omitempty"`
	Room      string `json:"room,omitempty"`
	Mentioned bool   `json:"mentioned,omitempty"`
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
							To:     source,
							Message: &messageBlock{
								Message: strings.TrimRight(string(output), "\n"),
								Room:    m.Room,
							},
						},
					}
				}

				if !pr.FallThrough {
					break
				}
			}
		}
	} else {
		// for _, pr := range noPrefixPResponders {
		// }
	}
}
