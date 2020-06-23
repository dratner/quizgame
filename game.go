package main

import (
	"fmt"
	"encoding/json"
	"errors"
	"github.com/rs/xid"
	"io/ioutil"
	"log"
	"strings"
)

//The form for State is enum+(Question Xid)
const (
	Setup		=  "setup"
	PoseQuestion = "pose"
	OfferAnsers  = "answer"
	ShowResults  = "show"
	Final        = "final"
)

type Player struct {
	Xid        string
	Name       string
	AccessCode string
	Score      int
}

type Answer struct {
	AnswerID int
	Answer   string
}

type Question struct {
	Xid     string
	Summary string
	First   string
	Last    string
}

func NewGame() *Game {
	xid := xid.New().String()
	accessCode := strings.ToUpper(xid[len(xid)-4 : len(xid)])
	return &Game{Xid: xid, AccessCode: accessCode}
}

type Game struct {
	Xid                string
	GameID             int         `json:GameID sql:game_id`
	AccessCode         string      `json:AccessCode sql:access_code`
	Players            []*Player   `json:Players sql:players`
	Questions          []*Question `json:Questions sql:questions`
	state              string      `json:State sql:state`
	CurrentQuestion 	*Question
}

func (g *Game) GetState() string {
	if g.CurrentQuestion == nil {
		return Setup
	}
	return g.state + g.CurrentQuestion.Xid
}

func (g *Game) Timeout() int {
	return 3
}

func (g *Game) ShowGame(token string) string {
	switch g.state {
	case PoseQuestion:
		return fmt.Sprintf("<p>%s</p><p>Your suggestion:</p>",g.CurrentQuestion.Summary)
	default:
		return ""
	}
}

func (g *Game) FromFile(f string) error {

	file, err := ioutil.ReadFile(f)

	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(file), &g)

	if err != nil {
		return err
	}

	for _, q := range g.Questions {
		q.Xid = xid.New().String()
	}

	return nil
}

// This function checks to see if the requestor has permission to view the game.
// For now, the token is just the xid.
func (g *Game) CheckPermission(token string) bool {
	for _, p := range g.Players {
		if p.Xid == token {
			return true
		}
	}
	return false
}

func (g *Game) AddPlayer(name string) (*Player, error) {
	for _, n := range g.Players {
		if n.Name == name {
			return nil, errors.New("Player with that name already in game")
		}
	}
	p := &Player{Xid: xid.New().String(), Name: name, AccessCode: g.AccessCode}
	g.Players = append(g.Players, p)
	return p, nil
}

func (g *Game) Setup() error {
	return g.Play()
}

func (g *Game) PlayQuestion(q *Question) error {
	g.state = PoseQuestion
	return nil
}

func (g *Game) Start() {
	err := g.FromFile("games/g1.json")
	if err != nil {
		log.Printf("Error: %s", err)
		return
	}
	log.Printf("Game file loaded.")
}

func (g *Game) Play() error {
	var err error

	for _, q := range g.Questions {
		g.CurrentQuestion = q
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
