package main

import (
	"os/exec"
	"strconv"
	"strings"
)

func triggerPassiveResponders(responders []*passiveResponderConfig, trimmed,
	source, room, mention string, mentioned bool,
	dispatch chan<- *dispatcherRequest) {

ResponderLoop:
	for _, pr := range prefixPResponders {
		for _, rg := range pr.regex {
			matches := rg.FindAllStringSubmatch(trimmed, 1)
			if len(matches) == 0 {
				continue
			}
			match := matches[0]

			logger.Debug.Println("Match:", match)

			var output []byte
			var err error
			respond := false
			switch pr.Type {
			case "simple":
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

				logger.Debug.Println("Simple command match")
				if subbed {
					output, err = exec.Command(pr.Cmd, subArgs...).Output()
				} else {
					output, err = exec.Command(pr.Cmd, pr.Args...).Output()
				}
				respond = true
			case "compound":
				logger.Debug.Println("Compound command match")
			default:
				logger.Error.Println("Unsupported passive responder type:",
					pr.Type)
			}

			if err != nil {
				logger.Error.Println("Passive responder error:", err)
			} else if respond {
				logger.Debug.Println("Passive responder executed:",
					string(output))

				dispatch <- &dispatcherRequest{
					Query: &query{
						Type:   "message",
						Source: "Passive Responder: " + pr.Name,
						To:     source,
						Message: &messageBlock{
							Message: strings.TrimRight(string(output), "\n"),
							Room:    room,
						},
					},
				}
			}

			// one regex in the match is good, continue onto next responder
			continue ResponderLoop
		}
	}
}
