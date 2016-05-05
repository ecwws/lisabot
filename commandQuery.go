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
	logger.Debug.Println("Id: ", c.Id)
	logger.Debug.Println("Action: ", c.Action)
	logger.Debug.Println("Type: ", c.Type)
	logger.Debug.Println("Time: ", c.Time)
}
