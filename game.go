package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/xid"
	"io/ioutil"
	"log"
	"strconv"
	//	"strings"
)

//The form for State is enum+(Question Xid)
const (
	StateSetup        = "setup"
	StateJoined       = "joined"
	StatePoseQuestion = "pose"
	StateOfferAnsers  = "answer"
	StateShowResults  = "show"
	StateFinal        = "final"
	ReqTypeJoin       = "join"
	ReqTypeStart      = "start"
	ReqTypePoll       = "poll"
	ReqTypeEnd        = "end"
	ReqTypeTimeout    = "timeout"
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

func PlayGame(ch chan PlayerReq, id string, accesscode string) {

	var presp PlayerResp

	g := &Game{Xid: id, AccessCode: accesscode, State: StateJoined}

	for {
		req := <-ch

		switch req.RequestType {
		case ReqTypeJoin:
			log.Printf("Adding player to game %s", g.AccessCode)
			p, err := g.AddPlayer(req.Payload)
			if err != nil {
				log.Printf("Error: %s", err)
			} else {
				file, _ := ioutil.ReadFile("wwwroot/startgame.html")
				presp = PlayerResp{TimerHtml: "", ScoreHtml: "", GameHtml: string(file), State: StateJoined, Payload: p.Xid}
				req.RespChan <- presp
			}
			break
		case ReqTypePoll:
			log.Printf("Polling game %s", g.AccessCode)
			if g.CheckPermission(req.Xid) {
				presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: g.ShowGame(req.Xid), State: g.State, Payload: ""}
				req.RespChan <- presp
			}
			break
		case ReqTypeStart:
			log.Printf("Starting game %s", g.AccessCode)
			g.Start()
			break
		case ReqTypeTimeout:
			log.Printf("Question timeout for game %s", g.AccessCode)
			break
		case ReqTypeEnd:
			log.Printf("Ending game %s", g.AccessCode)
			return
		default:
			log.Printf("Unidentified request.")
			break
		}
		presp = PlayerResp{}
		close(req.RespChan)
	}
}

type Game struct {
	Xid             string
	GameID          int         `json:GameID sql:game_id`
	AccessCode      string      `json:AccessCode sql:access_code`
	Players         []*Player   `json:Players sql:players`
	Questions       []*Question `json:Questions sql:questions`
	State           string      `json:State sql:state`
	CurrentQuestion *Question
}

func (g *Game) GetScores() string {
	payload := ""
	for _, p := range g.Players {
		payload += p.Name + ": " + strconv.Itoa(p.Score) + "&nbsp;&nbsp;&nbsp;"
	}
	return payload
}

func (g *Game) GetState() string {
	if g.CurrentQuestion == nil {
		return StateSetup
	}
	return g.State + g.CurrentQuestion.Xid
}

func (g *Game) GetTimer() string {
	return "3"
}

func (g *Game) ShowGame(token string) string {
	switch g.State {
	case StatePoseQuestion:
		return fmt.Sprintf("<p>%s</p><p>Your suggestion:</p>", g.CurrentQuestion.Summary)
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
	g.State = StatePoseQuestion
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
