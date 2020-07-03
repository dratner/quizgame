package main

import (
	"fmt"
)

type Message struct {
	Body string
	Name string
}

type Chat struct {
	PlayerIndex map[string]int
	Messages    []Message
}

func (c *Chat) AddMessage(name, msg string) {
	c.Messages = append(c.Messages, Message{Body: msg, Name: name})
}

func (c *Chat) GetMessages(id string) string {

	if c.PlayerIndex == nil {
		c.PlayerIndex = make(map[string]int)
	}

	if id == "" {
		return ""
	}
	if len(c.Messages) <= c.PlayerIndex[id] {
		return ""
	}
	msgs := c.Messages[c.PlayerIndex[id]:]
	c.PlayerIndex[id] = len(c.Messages)

	html := ""

	for _, m := range msgs {
		html += fmt.Sprintf("<p><strong>%s:</strong> %s</p>", m.Name, m.Body)
	}

	return html
}
