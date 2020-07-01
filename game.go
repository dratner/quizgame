package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/xid"
	"io/ioutil"
	"log"
	"math/rand"
	"sort"
	"strconv"
	"time"
)

//The form for State is enum+(Question Xid)
const (
	StateSetup        = "setup"
	StateTemp         = "temp"
	StatePoseQuestion = "pose"
	StateOfferAnswers = "answer"
	StateShowResults  = "show"
	StateFinal        = "final"
	ReqTypeJoin       = "join"
	ReqTypeStart      = "start"
	ReqTypePoll       = "poll"
	ReqTypeEnd        = "end"
	ReqTypeNext       = "next"
	ReqTypeTimeout    = "timeout"
	ReqTypeSubmit     = "submit"
	ReqTypeAnswer     = "answer"
	CorrectAnswer     = "correct"
)

type Player struct {
	Xid        string
	Name       string
	AccessCode string
	Score      int
}

type Answer struct {
	User   string
	Answer string
}

type Game struct {
	Xid             string
	GameID          int
	AccessCode      string
	Players         map[string]*Player
	Questions       []Question
	State           string
	CurrentQuestion Question
	finalHtml       string
}

// === USER FUNCTIONS ===

// Start a new game
func (g *Game) Start() {
	err := g.FromFile("games/g2.json")
	if err != nil {
		log.Printf("Error: %s", err)
		return
	}
	log.Printf("Game file loaded with %d questions.", len(g.Questions))
	g.PlayQuestion()
}

// Process a user submission of a potential answer
func (g *Game) Submit(id, payload string) {

	// No duplicates.
	if g.CurrentQuestion.HasSubmitted(id) {
		log.Printf("User already submitted a suggestion.")
		return
	}

	log.Printf("Not a duplicate.")

	a := Answer{User: id, Answer: payload}
	g.CurrentQuestion.Answers = append(g.CurrentQuestion.Answers, a)

	log.Printf("We've got %d total submissions.", len(g.CurrentQuestion.Answers))

	// Do we have all the answers? This is one way to change state.

	log.Println("Submission recorded.")

	if len(g.CurrentQuestion.Answers) == len(g.Players) {
		log.Printf("We've got all the submissions.")
		g.CloseQuestionSubmissions()
	}
}

// Process a user guess at which submission is correct
func (g *Game) Answer(id, payload string) {
	if g.CurrentQuestion.HasAnswered(id) {
		log.Printf("User already submitted an answer.")
		return
	}

	log.Println("Not a duplicate.")

	val, err := strconv.Atoi(payload)
	if err != nil {
		return
	}

	g.CurrentQuestion.Guesses[id] = val

	log.Printf("We've got %d total answers.", len(g.CurrentQuestion.Answers))

	log.Println("Answer recorded.")

	// Do we have all the answers?
	if len(g.CurrentQuestion.Guesses) == len(g.Players) {
		log.Printf("We've got all the answers.")
		g.CloseGuessSubmissions()
	} else {
		log.Println("We're still waiting for more answers.")
	}
}

// === UTIL FUNCTIONS ===

func (g *Game) FinalResult() string {

	m := make(map[string]int)

	for _, p := range g.Players {
		m[p.Name] = p.Score
	}

	type kv struct {
		K string
		V int
	}

	var ss []kv
	for k, v := range m {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].V > ss[j].V
	})

	html := fmt.Sprintf("<p><strong>The winner is %s!</strong></p><p>", ss[0].K)

	for _, kv := range ss {
		html += fmt.Sprintf("%s got %d points.<br />\n", kv.K, kv.V)
	}

	return html + "<br />"

}

func (g *Game) PlayQuestion() {

	// If there are no more questions to be asked, we are done.
	if len(g.Questions) == 0 {
		log.Println("All questions have been asked.")
		g.State = StateFinal
		return
	}

	// Move the next question to current and play it.
	g.CurrentQuestion = g.Questions[0]
	g.CurrentQuestion.Guesses = make(map[string]int)
	g.Questions = g.Questions[1:]
	g.State = StatePoseQuestion
	log.Printf("Posing question %s. %d questions remaining.", g.CurrentQuestion.Xid, len(g.Questions))
}

func (g *Game) CloseQuestionSubmissions() {

	// We have all the submissions from users or we timed out.

	if !g.CurrentQuestion.HasSubmitted(CorrectAnswer) {

		// Add the correct answer to the list of possible answers.

		ca := Answer{User: CorrectAnswer, Answer: g.CurrentQuestion.First}
		g.CurrentQuestion.Answers = append(g.CurrentQuestion.Answers, ca)

		// Shuffle the answers.

		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(g.CurrentQuestion.Answers), func(i, j int) {
			g.CurrentQuestion.Answers[i], g.CurrentQuestion.Answers[j] = g.CurrentQuestion.Answers[j], g.CurrentQuestion.Answers[i]
		})
	}
	g.State = StateOfferAnswers
}

func (g *Game) CloseGuessSubmissions() {

	// Score the results
	var correct int
	for id, answer := range g.CurrentQuestion.Answers {
		if answer.User == CorrectAnswer {
			correct = id
		}
	}
	for uid, guess := range g.CurrentQuestion.Guesses {
		if guess == correct {
			// Give everyone a point who got it right...
			g.Players[uid].Score++
		} else {
			// Give everyone a point who fooled someone...
			if uid == g.CurrentQuestion.Answers[guess].User {
				log.Printf("A user guessed their own answer. Lame-o.")
			} else {
				g.Players[g.CurrentQuestion.Answers[guess].User].Score++
			}
		}
	}
	g.State = StateShowResults
}

// === THESE ARE DISPLAY FUNCS ===

func (g *Game) GetScores() string {
	payload := ""
	for _, p := range g.Players {
		payload += p.Name + ": " + strconv.Itoa(p.Score) + "&nbsp;&nbsp;&nbsp;"
	}
	return payload
}

func (g *Game) GetState() string {
	if g.CurrentQuestion.Xid == "" {
		return g.State
	}
	return g.State + g.CurrentQuestion.Xid
}

func (g *Game) GetTimer() string {
	return ""
}

func (g *Game) ShowGame(token string) string {

	var html string

	switch g.State {
	case StatePoseQuestion:
		if g.CurrentQuestion.HasSubmitted(token) {
			html = "Waiting for others to submit their sentences."
		} else {
			html = fmt.Sprintf("<p>%s</p><p>Your first sentence:</p>", g.CurrentQuestion.Summary) +
				`	<p><textarea id="submission" placeholder="Your sentence" rows="6" cols="40"></textarea></p>
				<p><button onclick="submitGame()">Submit!</button></p>
			`
		}
		return html
	case StateOfferAnswers:
		if g.CurrentQuestion.HasAnswered(token) {
			html = "Wait for other to make their guesses."
		} else {
			html = `
				Answer List
		`
			for v, a := range g.CurrentQuestion.Answers {
				html += fmt.Sprintf(`<p><input type="radio" name="answer" value="%d">&nbsp;%s</p>`, v, a.Answer)
			}
			html += `<p><button onclick="answerGame()">Guess!</button></p>`
		}
		return html
	case StateShowResults:
		html = fmt.Sprintf(`
				<p>The correct answer was:<p><p>%s</p>
				<p><button onclick="nextGame()">Next book!</button></p>
		`, g.CurrentQuestion.First)
		return html
	case StateFinal:
		if g.finalHtml == "" {
			g.finalHtml = g.FinalResult()
		}
		return g.finalHtml
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
		q.Xid = xid.New().String()
		log.Printf("Loading quesion %s", q.Xid)
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
	g.Players[p.Xid] = p
	return p, nil
}
