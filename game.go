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
	StateSetup             = "setup"
	StateTemp              = "temp"
	StatePoseQuestion      = "pose"
	StateOfferAnswers      = "answer"
	StateShowResults       = "show"
	StateFinal             = "final"
	ReqTypeJoin            = "join"
	ReqTypeStart           = "start"
	ReqTypePoll            = "poll"
	ReqTypeEnd             = "end"
	ReqTypeNext            = "next"
	ReqTypeTimeoutQuestion = "questiontimeout"
	ReqTypeTimeoutAnswer   = "answertimeout"
	ReqTypeTimeoutDisplay  = "displaytimeout"
	ReqTypeSubmit          = "submit"
	ReqTypeAnswer          = "answer"
	ReqTypeChat            = "chat"
	CorrectAnswer          = "correct"
	TimeoutQuestion        = 180   // 3 minutes
	TimeoutAnswer          = 60    // 1 minute
	TimeoutDisplay         = 30    // 30 seconds
	TimeoutGame            = 86400 // 24 hours
	AdminName              = "Admin"
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
	Players         []*Player
	Questions       []Question
	DoneQuestions   []Question
	CurrentQuestion Question
	PlayerChat      Chat
	finalHtml       string
	Ch              chan PlayerReq
	state           string
}

// === USER FUNCTIONS ===

func (g *Game) PlayerByXid(id string) *Player {
	for _, p := range g.Players {
		if p.Xid == id {
			return p
		}
	}

	// Don't return a nil since the results are often piped. Return a zero value instead.
	return new(Player)
}

// Start a new game
func (g *Game) Start() {
	err := g.FromFile("games/g1.json")
	if err != nil {
		log.Printf("[Game %s] Error: %s", g.AccessCode, err)
		return
	}
	log.Printf("[Game %s] Game file loaded with %d questions.", g.AccessCode, len(g.Questions))
	g.PlayQuestion()
}

// Process a user submission of a potential answer
func (g *Game) Submit(id, payload string) {

	// No duplicates.
	if g.CurrentQuestion.HasSubmitted(id) {
		log.Printf("[Game %s] User already submitted a suggestion.", g.AccessCode)
		return
	}

	log.Printf("[Game %s] Not a duplicate.", g.AccessCode)

	a := Answer{User: id, Answer: payload}
	g.CurrentQuestion.Answers = append(g.CurrentQuestion.Answers, a)

	log.Printf("[Game %s] We've got %d total submissions.", g.AccessCode, len(g.CurrentQuestion.Answers))

	// Do we have all the answers? This is one way to change state.

	log.Printf("[Game %s] Submission recorded.", g.AccessCode)

	if len(g.CurrentQuestion.Answers) == len(g.Players) {
		log.Printf("[Game %s] We've got all the submissions.", g.AccessCode)
		g.CloseQuestionSubmissions()
	}
}

// Process a user guess at which submission is correct
func (g *Game) Answer(id, payload string) {
	if g.CurrentQuestion.HasAnswered(id) {
		log.Printf("[Game %s] User already submitted an answer.", g.AccessCode)
		return
	}

	log.Printf("[Game %s] Not a duplicate.", g.AccessCode)

	val, err := strconv.Atoi(payload)
	if err != nil {
		return
	}

	g.CurrentQuestion.Guesses[id] = val

	log.Printf("[Game %s] We've got %d total answers.", g.AccessCode, len(g.CurrentQuestion.Answers))

	log.Printf("[Game %s] Answer recorded.", g.AccessCode)

	// Do we have all the answers?
	if len(g.CurrentQuestion.Guesses) == len(g.Players) {
		log.Printf("[Game %s] We've got all the answers.", g.AccessCode)
		g.CloseGuessSubmissions()
	} else {
		log.Printf("[Game %s] We're still waiting for more answers.", g.AccessCode)
	}
}

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

	var html string

	if len(ss) > 1 {
		if ss[0].V == ss[1].V {
			html = "<p><strong>Tie game!</strong></p><p>"
		} else {
			html = fmt.Sprintf("<p><strong>The winner is %s!</strong></p><p>", ss[0].K)
		}
	} else {
		html = "<p><strong>Did you beat yourself?</strong></p><p>"
	}

	for _, kv := range ss {
		html += fmt.Sprintf("%s got %d points.<br />\n", kv.K, kv.V)
	}

	html += "</p><p>The books were:</p><p>"

	for _, q := range g.DoneQuestions {
		html += fmt.Sprintf("<i>%s</i><br />", q.Title)
	}

	return html

}

func (g *Game) PlayQuestion() {

	if g.CurrentQuestion.Xid != "" {
		g.DoneQuestions = append(g.DoneQuestions, g.CurrentQuestion)
	}

	// If there are no more questions to be asked, we are done.
	if len(g.Questions) == 0 {
		log.Printf("[Game %s] All questions have been asked.", g.AccessCode)
		g.SetState(StateFinal)
		return
	}

	g.PlayerChat.AddMessage(AdminName, fmt.Sprintf("There are %d books left.", len(g.Questions)))

	// Move the next question to current and play it.
	g.CurrentQuestion = g.Questions[0]
	g.CurrentQuestion.Guesses = make(map[string]int)
	g.CurrentQuestion.Posed = time.Now()
	g.Questions = g.Questions[1:]
	g.SetState(StatePoseQuestion)

	log.Printf("[Game %s] Posing question %s. %d questions remaining.", g.AccessCode, g.CurrentQuestion.Xid, len(g.Questions))
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
	g.SetState(StateOfferAnswers)
	g.CurrentQuestion.Guessed = time.Now()
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
			g.PlayerByXid(uid).Score++
		} else {
			// Give everyone a point who fooled someone...
			if uid == g.CurrentQuestion.Answers[guess].User {
				g.PlayerChat.AddMessage(AdminName, fmt.Sprintf("Come on, %s, don't guess your own answer. Lame-o.", g.PlayerByXid(uid).Name))
				log.Printf("[Game %s] A user guessed their own answer. Lame-o.", g.AccessCode)
			} else {
				g.PlayerByXid(g.CurrentQuestion.Answers[guess].User).Score++
			}
		}
	}
	g.SetState(StateShowResults)
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
	return g.state
}

func (g *Game) SetState(state string) {
	if state == g.state {
		return
	}
	switch state {
	case StateShowResults:
		g.CurrentQuestion.Results = time.Now()
		go Timeout(g.Ch, g.CurrentQuestion.Xid, ReqTypeTimeoutDisplay, TimeoutDisplay)
		break
	case StateOfferAnswers:
		g.CurrentQuestion.Guessed = time.Now()
		go Timeout(g.Ch, g.CurrentQuestion.Xid, ReqTypeTimeoutAnswer, TimeoutAnswer)
		break
	case StatePoseQuestion:
		g.CurrentQuestion.Posed = time.Now()
		go Timeout(g.Ch, g.CurrentQuestion.Xid, ReqTypeTimeoutQuestion, TimeoutQuestion)
		break
	}
	g.state = state
}

func (g *Game) GetTimer() string {

	var i int

	if g.state == StatePoseQuestion {
		i = TimeoutQuestion - int(time.Since(g.CurrentQuestion.Posed).Seconds())
	} else if g.state == StateOfferAnswers {
		i = TimeoutAnswer - int(time.Since(g.CurrentQuestion.Guessed).Seconds())
	} else if g.state == StateShowResults {
		i = TimeoutDisplay - int(time.Since(g.CurrentQuestion.Results).Seconds())
	} else {
		return ""
	}

	if i < 0 {
		i = 0
	}

	return fmt.Sprintf("%d seconds", i)
}

func (g *Game) ShowGame(token string) string {

	var html string

	switch g.state {
	case StateSetup:
		html = `<p><strong>Ready To Start?</strong></p>
        	<p>Is everyone here?</p>
			<button onclick="startGame()">Start Game!</button> 
			`
		return html
	case StatePoseQuestion:
		if g.CurrentQuestion.HasSubmitted(token) {
			html = "Waiting for others to submit their sentences..."
		} else {
			html = fmt.Sprintf("<p>%s</p>", g.CurrentQuestion.Summary) +
				`	<p><textarea id="submission" placeholder="Your first sentence" rows="6" cols="80"></textarea></p>
				<p><button onclick="submitGame()">Submit!</button></p>
			`
		}
		return html
	case StateOfferAnswers:
		if g.CurrentQuestion.HasAnswered(token) {
			html = "Waiting for others to make their guesses..."
		} else {
			html = "<strong>Your Choices:</strong>"

			for v, a := range g.CurrentQuestion.Answers {
				html += fmt.Sprintf(`<p><input type="radio" name="answer" value="%d">&nbsp;%s</p>`, v, a.Answer)
			}
			html += `<p><button onclick="answerGame()">Guess!</button></p>`
		}
		return html
	case StateShowResults:

		if g.CurrentQuestion.ResultsHtml == "" {

			type Result struct {
				Title     string
				Answer    string
				GuessedBy string
			}

			for aid, ans := range g.CurrentQuestion.Answers {
				r := Result{Answer: ans.Answer}

				if ans.User == CorrectAnswer {
					r.Title = "Correct Answer:"
				} else {
					r.Title = fmt.Sprintf("Suggestion From %s:", g.PlayerByXid(ans.User).Name)
				}

				for pxid, gid := range g.CurrentQuestion.Guesses {
					if gid == aid {
						if r.GuessedBy != "" {
							r.GuessedBy += ", "
						}
						r.GuessedBy += g.PlayerByXid(pxid).Name
					}
				}
				if r.GuessedBy == "" {
					r.GuessedBy = "No players guessed this one."
				}
				g.CurrentQuestion.ResultsHtml += fmt.Sprintf("<p><strong>%s</strong></p><p>%s</p><p><em>Guessed by: %s</em></p>", r.Title, r.Answer, r.GuessedBy)
			}
			g.CurrentQuestion.ResultsHtml += `<p><button onclick="nextGame()">Next book!</button></p>`
		}
		return g.CurrentQuestion.ResultsHtml
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

	log.Printf("[Game %s] Load questions from file.", g.AccessCode)

	file, err := ioutil.ReadFile(f)

	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(file), &g.Questions)

	if err != nil {
		return err
	}

	log.Printf("[Game %s] Reading...", g.AccessCode)

	for i, _ := range g.Questions {
		g.Questions[i].Xid = xid.New().String()
	}

	log.Printf("[Game %s] Loaded %d quesions.", g.AccessCode, len(g.Questions))

	log.Printf("[Game %s] Shuffling...", g.AccessCode)

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(g.Questions), func(i, j int) {
		g.Questions[i], g.Questions[j] = g.Questions[j], g.Questions[i]
	})

	log.Printf("[Game %s] Cutting...", g.AccessCode)

	if len(g.Questions) > 7 {
		g.Questions = g.Questions[0:7]
	}

	log.Printf("[Game %s] %d questions Ready.", g.AccessCode, len(g.Questions))

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
