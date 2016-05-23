package main

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func triggerPassiveResponders(responders []*passiveResponderConfig, message,
	source, room, from string, mentionMode bool,
	dispatch chan<- *dispatcherRequest) (matched bool) {

ResponderLoop:
	for _, pr := range responders {
		var patterns []*regexp.Regexp
		if mentionMode {
			logger.Debug.Println("Using mention pattern")
			patterns = pr.mRegex
		} else {
			logger.Debug.Println("Using regular pattern")
			patterns = pr.regex
		}

		for _, rg := range patterns {
			logger.Debug.Println("Trying to match:", pr.Name)
			logger.Debug.Println("Pattern:", *rg)

			matches := rg.FindAllStringSubmatch(message, 1)
			if len(matches) == 0 {
				continue
			}
			match := matches[0]

			logger.Debug.Println("Match:", match)

			var output []byte
			var err error

			var subArgs []string
			subbed := false
			// submatch, may need to substitute
			if len(match) > 1 && len(pr.substitute) > 0 {
				subArgs = make([]string, len(pr.Args))
				copy(subArgs, pr.Args)
				for i, _ := range pr.substitute {
					logger.Debug.Println("Try sub:", subArgs[i])

					matchIds :=
						subRegex.FindAllStringSubmatch(subArgs[i], -1)
					logger.Debug.Println("MatchIds:", matchIds)

					for _, matchId := range matchIds {
						logger.Debug.Println("MatchId:", matchId)

						mId, _ := strconv.Atoi(matchId[1])
						if mId < len(match)-1 {
							logger.Debug.Println("Subbed:", match[mId+1])

							subArgs[i] = strings.Replace(subArgs[i],
								"__"+matchId[1]+"__", match[mId+1], -1)
							subbed = true
						}
					}
				}
			}

			if subbed {
				output, err = exec.Command(pr.Cmd, subArgs...).Output()
			} else {
				output, err = exec.Command(pr.Cmd, pr.Args...).Output()
			}
			matched = true

			if err != nil {
				logger.Error.Println("Passive responder error:", err)
			} else {
				logger.Debug.Println("Passive responder executed:",
					string(output))

				request := dispatcherRequest{
					Query: &query{
						Type:   "message",
						Source: "Passive Responder: " + pr.Name,
						To:     source,
						Message: &messageBlock{
							Message: strings.Trim(string(output), " \n"),
							Room:    room,
						},
					},
				}

				if mentionMode {
					request.Query.Message.MentionNotify = []string{from}
				}

				dispatch <- &request
			}

			// one regex in the match is good, continue onto next responder
			continue ResponderLoop
		}
	}
	return
}
