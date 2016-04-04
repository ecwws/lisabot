package main

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
