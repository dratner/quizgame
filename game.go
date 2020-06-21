package main

import (
	//"fmt"
	"encoding/json"
	"errors"
	"github.com/rs/xid"
	"io/ioutil"
	"strings"
)

type Player struct {
	Xid        string
	Name       string
	AccessCode string
}

type Answer struct {
	AnswerID int
	Answer   string
}

type Question struct {
	QuestionID      int
	Question        string
	Answers         []string
	CorrectAnswerID int
	Timeout         int
}

type GameState int

const (
	Signup GameState = iota + 1
	AskQuestion
	ShowAnswer
	ScoreGame
)

func NewGame() *Game {
	xid := xid.New().String()
	accessCode := strings.ToUpper(xid[len(xid)-4 : len(xid)])
	players := make(map[string]*Player)
	return &Game{Xid: xid, AccessCode: accessCode, Players: players}
}

type Game struct {
	Xid        string
	GameID     int                `json:GameID sql:game_id`
	AccessCode string             `json:AccessCode sql:access_code`
	Players    map[string]*Player `json:Players sql:players`
	Questions  []*Question        `json:Questions sql:questions`
	State      string             `json:State sql:state`
}

func (g *Game) FromFile(f string) error {

	file, err := ioutil.ReadFile(f)

	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(file), &g)
}

func (g *Game) AddPlayer(name string) (*Player, error) {
	for _, n := range g.Players {
		if n.Name == name {
			return nil, errors.New("Player with that name already in game")
		}
	}
	p := &Player{Xid: xid.New().String(), Name: name, AccessCode: g.AccessCode}
	g.Players[p.Xid] = p
	return p, nil
}

func (g *Game) Setup() error {
	return g.Play()
}

func (g *Game) PlayQuestion(q *Question) error {
	return nil
}

func (g *Game) Play() error {
	var err error

	for _, q := range g.Questions {
		err = g.PlayQuestion(q)
	}

	return err
}

type PlayerAnswer struct {
	GameID     int
	QuestionID int
	AnswerID   int
	PlayerID   int
}
