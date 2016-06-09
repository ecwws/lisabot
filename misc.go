package main

import (
	"fmt"
	"strings"
)

func checkHelp(msg, source, room string, dp chan<- *dispatcherRequest) bool {
	logger.Debug.Println("Checking help command:", msg)

	matches := conf.helpRegex.FindAllStringSubmatch(msg, 1)

	if len(matches) == 0 {
		logger.Debug.Println("No help match found")
		return false
	}

	section := ""
	if len(matches[0]) > 1 {
		section = matches[0][1]
	}

	dp <- &dispatcherRequest{
		Query: &query{
			Type:   "message",
			Source: "Internal: help",
			To:     source,
			Message: &messageBlock{
				Message: strings.Trim(showHelp(section), " \n"),
				Room:    room,
			},
		},
	}

	return true
}

func showHelp(cmdPrefix string) string {

	helpMsg := "Here is what I can do:\n"
	for helpE := help.Front(); helpE != nil; helpE = helpE.Next() {
		h := helpE.Value.(*helpInfo)

		if h.mention {
			helpMsg += fmt.Sprintf("(when mentioned) %s - %s\n", h.helpCmd,
				h.helpMsg)
		} else if h.noPrefix {
			helpMsg += fmt.Sprintf("%s - %s\n", h.helpCmd, h.helpMsg)
		} else {
			helpMsg += fmt.Sprintf("%s %s - %s\n", conf.Prefix, h.helpCmd,
				h.helpMsg)
		}
	}

	return helpMsg
}
