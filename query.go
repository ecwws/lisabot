package main

import (
	"errors"
)

type query struct {
	Type    string        `json:"type"`
	Source  string        `json:"source"`
	To      string        `json:"to"`
	Command *commandBlock `json:"command"`
	Message *messageBlock `json:"message"`
}

func (q *query) validate() error {
	switch {
	case q.Type == "command" && q.Command == nil:
		return errors.New("Missing command block")
	case q.Type == "message" && q.Message == nil:
		return errors.New("Missing message block")
	case q.Type != "command" && q.Type != "message":
		return errors.New("Invalid query type")
	default:
		return nil
	}
}

func (q *query) checkEngagement() error {
	switch {
	case q.Type != "command":
		return errors.New("Invalid query type for engegement: " + q.Type)
	case q.Command == nil:
		return errors.New("Empty command block")
	case q.Command.Action != "engage":
		return errors.New("First command is not engagement: " + q.Command.Action)
	case q.Command.Type != "adapter" && q.Command.Type != "responder":
		return errors.New("Unsupported engagement type: " + q.Command.Type)
	default:
		return nil
	}
}
