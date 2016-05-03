package main

type commandBlock struct {
	Id      string   `json:"id"`
	Action  string   `json:"action"`
	Type    string   `json:"type"`
	Time    int      `json:"time"`
	Data    string   `json:"data"`
	Array   []string `json:"array"`
	Options []string `json:"options"`
}

func (c *commandBlock) handleCommand(source string) {
	if debugOut {
		logstd.Println("Id: ", c.Id)
		logstd.Println("Action: ", c.Action)
		logstd.Println("Type: ", c.Type)
		logstd.Println("Time: ", c.Time)
	}
}
