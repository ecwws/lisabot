package main

import (
	"fmt"
)

func (m *messageBlock) handleMessage(source string) {
	fmt.Println("Message: ", m.Message)
	fmt.Println("From: ", m.From)
	fmt.Println("Room: ", m.Room)
}
