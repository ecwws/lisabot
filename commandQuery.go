package main

import (
	"fmt"
)

func (c *commandBlock) handleCommand(source string) {
	fmt.Println("Id: ", c.Id)
	fmt.Println("Action: ", c.Action)
	fmt.Println("Type: ", c.Type)
	fmt.Println("Time: ", c.Time)
}
