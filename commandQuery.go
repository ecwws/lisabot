package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

type commandBlock struct {
	Id      string   `json:"id"`
	Action  string   `json:"action"`
	Type    string   `json:"type"`
	Time    int64    `json:"time"`
	Data    string   `json:"data"`
	Array   []string `json:"array"`
	Options []string `json:"options"`
}

func (c *commandBlock) handleCommand(source string,
	dispatch chan<- *dispatcherRequest) {

	logger.Debug.Println("Id: ", c.Id)
	logger.Debug.Println("Action: ", c.Action)
	logger.Debug.Println("Type: ", c.Type)
	logger.Debug.Println("Time: ", c.Time)
}

func (c *commandBlock) engageChk(source, secret string) error {
	if c.Type != "adapter" && c.Type != "responder" {
		return errors.New("Invalid client engagement type: " + c.Type)
	}

	if c.Data == "" {
		return errors.New("No auth data received")
	}

	diff := time.Now().Unix() - c.Time

	if diff > 10 || diff < 0 {
		return errors.New("Timestamp out of range")
	}

	decoded, err := base64.StdEncoding.DecodeString(c.Data)

	if err != nil {
		return err
	}

	authMsg := fmt.Sprintf("%d%s%s", c.Time, source, secret)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(authMsg))

	if !hmac.Equal(decoded, mac.Sum(nil)) {
		return errors.New("Incorrect auth code")
	}

	return nil

}
