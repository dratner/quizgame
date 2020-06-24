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
	StateTemp         = "temp"
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

	g := &Game{Xid: id, AccessCode: accesscode, State: StateSetup}

	for {
		req := <-ch

		presp = PlayerResp{}

		switch req.RequestType {
		case ReqTypeJoin:
			log.Printf("Adding player %s to game %s", req.Payload, g.AccessCode)
			p, err := g.AddPlayer(req.Payload)
			if err != nil {
				log.Printf("Error: %s", err)
			} else {
				presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: "", State: g.GetState(), Payload: p.Xid}
				req.RespChan <- presp
			}
			break
		case ReqTypePoll:
			log.Printf("Polling game %s", g.AccessCode)
			if g.CheckPermission(req.Xid) {
				presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: g.ShowGame(req.Xid), State: g.GetState(), Payload: ""}
				req.RespChan <- presp
			}
			break
		case ReqTypeStart:
			log.Printf("Starting game %s", g.AccessCode)
			g.Start()
			presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: "OK! Let's go!", State: StateTemp, Payload: ""}
			req.RespChan <- presp
			break
		case ReqTypeTimeout:
			log.Printf("Question timeout for game %s", g.AccessCode)
			break
		case ReqTypeEnd:
			log.Printf("Ending game %s", g.AccessCode)
			presp = PlayerResp{}
			req.RespChan <- presp
			return
		default:
			log.Printf("Unidentified request.")
			break
		}
		close(req.RespChan)
	}
}

/*
type Game struct {
	Xid             string
	GameID          int         `json:GameID`
	AccessCode      string      `json:AccessCode`
	Players         []*Player   `json:Players`
	Questions       []*Question `json:Questions`
	State           string      `json:State`
	CurrentQuestion *Question
}*/

type Game struct {
	Xid             string
	GameID          int
	AccessCode      string
	Players         []*Player
	Questions       []Question
	State           string
	CurrentQuestion Question
}

func (g *Game) GetScores() string {
	payload := ""
	for _, p := range g.Players {
		payload += p.Name + ": " + strconv.Itoa(p.Score) + "&nbsp;&nbsp;&nbsp;"
	}
	return payload
}

func (g *Game) GetState() string {
	if g.CurrentQuestion.Xid == "" {
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

	err = json.Unmarshal([]byte(file), &g.Questions)

	if err != nil {
		return err
	}

	log.Println("reading")
	for _, q := range g.Questions {
		log.Println("quesion")
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

func (g *Game) Start() {
	err := g.FromFile("games/g1.json")
	if err != nil {
		log.Printf("Error: %s", err)
		return
	}
	log.Printf("Game file loaded with %d questions.", len(g.Questions))
	_ = g.PlayQuestion()
}

func (g *Game) PlayQuestion() error {
	g.CurrentQuestion = g.Questions[0]
	g.Questions = g.Questions[1:]
	g.State = StatePoseQuestion
	log.Printf("Posing question %s. %d questions remaining.", g.CurrentQuestion.Xid, len(g.Questions))
	return nil
}
