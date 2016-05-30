package main

import (
	"fmt"
	"strings"
)

func checkHelp(msg, source, room string, dp chan<- *dispatcherRequest) bool {
	logger.Debug.Println("Checking help command...")

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

func showHelp(cmdPrefix string) (helpMsg string) {

	helpMsg += "Here is what I can do:\n"
	for cmd, msg := range help {
		helpMsg += fmt.Sprintf("%s - %s\n", cmd, msg)
	}

	return
}
