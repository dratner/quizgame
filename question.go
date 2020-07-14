package main

import (
	"github.com/rs/xid"
	"math/rand"
	"time"
)

const (
	FirstLine = "first"
	LastLine  = "last"
)

type Question struct {
	Xid         string
	Title       string
	Author      string
	Summary     string
	First       string
	Last        string
	Kind        string
	Answers     []Answer
	Guesses     map[string]int
	Posed       time.Time
	Guessed     time.Time
	Results     time.Time
	ResultsHtml string // Caching final output
}

func (q *Question) Init() {

	// Get an ID

	q.Xid = xid.New().String()

	// Decide first or last

	rand.Seed(time.Now().UnixNano())
	v := rand.Int()
	ca := Answer{User: CorrectAnswer}
	if v%2 == 0 {
		q.Kind = FirstLine
		ca.Answer = q.First
	} else {
		q.Kind = LastLine
		ca.Answer = q.Last
	}

	// Seed the correct answer

	q.Answers = append(q.Answers, ca)
}

func (q *Question) Shuffle() {

	// Shuffle the answers.

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(q.Answers), func(i, j int) {
		q.Answers[i], q.Answers[j] = q.Answers[j], q.Answers[i]
	})
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
