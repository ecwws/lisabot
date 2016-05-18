package main

type query struct {
	Type    string        `json:"type"`
	Source  string        `json:"source"`
	Command *commandBlock `json:"command"`
	Message *messageBlock `json:"message"`
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