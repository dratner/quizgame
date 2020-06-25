package main

import (
	"log"
)

type Question struct {
	Xid     string
	Summary string
	First   string
	Last    string
	Answers []Answer
	Guesses map[string]int
}

func (q *Question) HasSubmitted(id string) bool {
	for _, a := range q.Answers {
		log.Printf("U %s", a.User)
		log.Printf("A %s", a.Answer)
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
