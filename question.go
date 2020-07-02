package main

import (
	"time"
)

type Question struct {
	Xid     string
	Summary string
	First   string
	Last    string
	Answers []Answer
	Guesses map[string]int
	Posed   time.Time
}

func (q *Question) HasSubmitted(id string) bool {
	for _, a := range q.Answers {
		if a.User == id {
			return true
		}
	}
	return false
}

func (q *Question) HasAnswered(id string) bool {
	if _, ok := q.Guesses[id]; ok {
		return true
	}
	return false
}
