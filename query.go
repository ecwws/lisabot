package main

type messageBlock struct {
	Message string `json:"message"`
	From    string `json:"from"`
	Room    string `json:"room"`
}

type query struct {
	Type    string        `json:"type"`
	Source  string        `json:"source"`
	Command *commandBlock `json:"command"`
	Message *messageBlock `json:"message"`
}

func (q *query) process() {

	switch {
	case q.Type == "command":
		if q.Command != nil {
			q.Command.handleCommand(q.Source)
		}
	case q.Type == "message":
		if q.Message != nil {
			q.Message.handleMessage(q.Source)
		}
	}

	if q.Type == "command" && q.Command != nil {
	}
}

func (q *query) validate() bool {
	switch {
	case q.Type == "command" && q.Command != nil:
		return true
	case q.Type == "message" && q.Message != nil:
		return true
	default:
		return false
	}
}
