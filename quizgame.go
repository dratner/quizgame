package main

import (
	"encoding/json"
	//	"errors"
	"fmt"
	"github.com/rs/xid"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type PlayerReq struct {
	Xid         string
	AccessCode  string
	RequestType string
	Payload     string
	RespChan    chan PlayerResp
}

type PlayerResp struct {
	TimerHtml string
	ScoreHtml string
	GameHtml  string
	State     string
	Payload   string
}

var Games map[string]chan PlayerReq

type htmlHandler struct {
}

func (h *htmlHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	file, err := ioutil.ReadFile("wwwroot/index.html")

	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
		return
	}

	w.Write(file)
}

type newHandler struct {
}

func (h *newHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	log.Printf("Creating new game.")

	ch := make(chan PlayerReq, 100)
	id := xid.New().String()
	ac := strings.ToUpper(id[len(id)-4 : len(id)])

	if _, ok := Games[ac]; ok {
		log.Printf("Error because game with access code %s already exists.", ac)
		http.Error(w, fmt.Sprintf("Error because game with access code %s already exists.", ac), http.StatusInternalServerError)
		return
	}

	Games[ac] = ch

	go PlayGame(ch, id, ac)

	presp := &PlayerResp{TimerHtml: "", ScoreHtml: "", GameHtml: fmt.Sprintf("You made a new game with access code %s. Invite your friends!", ac), State: StateSetup, Payload: ac}

	jsonOut, err := json.Marshal(presp)

	if err != nil {
		log.Printf("Error in JSON encoding: %s", err)
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Created new game with access code %s.", ac)

	w.Write(jsonOut)
}

type reqHandler struct {
}

func (h *reqHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	log.Printf("Req handler.")

	decoder := json.NewDecoder(r.Body)

	var preq *PlayerReq
	err := decoder.Decode(&preq)

	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
		return
	}

	if _, ok := Games[preq.AccessCode]; !ok {
		log.Printf("HERE")
		for a, _ := range Games {
			log.Printf("AC: %s", a)
		}
		log.Printf("PREQ %v", preq)
		log.Printf("Error: Game with access code %s does not exist.", preq.AccessCode)
		http.Error(w, fmt.Sprintf("Error: Game with access code %s does not exist.", preq.AccessCode), http.StatusInternalServerError)
		return
	}

	log.Printf("Handling %s request to game %s.", preq.RequestType, preq.AccessCode)

	preq.RespChan = make(chan PlayerResp)

	Games[preq.AccessCode] <- *preq

	var presp PlayerResp

	presp = <-preq.RespChan

	if presp.State != "" {
		jsonOut, err := json.Marshal(presp)

		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
			return
		}

		log.Printf("Handler successful.")

		w.Write(jsonOut)
	}

	log.Println("Request complete.")
}

/*
type startHandler struct {
}

func (h *startHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var preq *PlayerReq
	err := decoder.Decode(&preq)

	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
		return
	}

	if _, ok := Games[preq.AccessCode]; ok {
		if Games[preq.AccessCode].CheckPermission(preq.Xid) {
			log.Printf("Starting game with Access Code %s.", preq.AccessCode)
			Games[preq.AccessCode].Start()
		}
	} else {
		err = errors.New("A game with that access code does not exist.")
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
		return
	}
}





type joinHandler struct {
}

func (h *joinHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	log.Printf("Handle add player.")

	presp := new(PlayerResp)

	decoder := json.NewDecoder(r.Body)

	var p *Player
	err := decoder.Decode(&p)

	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
		return
	}

	// If the player isn't defined, serve the request html.
	if p.AccessCode == "" {
		log.Printf("No player data so serving request page.")
		file, err := ioutil.ReadFile("wwwroot/joingame.html")

		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
			return
		}

		presp.Html = string(file)

	} else {

		log.Printf("Got player with name %s for game with access code %s.", p.Name, p.AccessCode)

		if _, ok := Games[p.AccessCode]; ok {
			p, err = Games[p.AccessCode].AddPlayer(p.Name)
		} else {
			err = errors.New("A game with that access code does not exist.")
		}

		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
			return
		}
		log.Printf("Created new player with xid %s.", p.Xid)

		file, err := ioutil.ReadFile("wwwroot/startgame.html")

		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
			return
		}

		presp.Html = string(file)

	}

	presp.Payload = p

	jsonOut, err := json.Marshal(presp)

	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Handler successful.")

	w.Write(jsonOut)
}

var Games map[string]*Game
*/

func main() {

	Games = make(map[string]chan PlayerReq)

	http.Handle("/", new(htmlHandler))
	http.Handle("/new", new(newHandler))
	http.Handle("/req", new(reqHandler))

	/*http.Handle("/play", new(htmlHandler))
	http.Handle("/join", new(joinHandler))
	http.Handle("/start", new(startHandler))
	http.Handle("/poll", new(pollHandler))
	*/
	//	http.Handle("/submit", new(submitHandler))
	//	http.Handle("/guess", new(guessHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))

	/*
	   	g := new(Game)
	   	q1 := new(Question)
	   	q1.Summary = "This was a great book."
	   	q1.First = "In the beginning..."
	   	q1.Last = "...The end."
	   	q2 := new(Question)
	   	q2.Summary = "This was also great book."
	   	q2.First = "In the beginning..."
	   	q2.Last = "...The end."
	   	g.Questions = append(g.Questions,q1)
	   	g.Questions = append(g.Questions,q2)


	   	jsonOut, _ := json.Marshal(g.Questions)
	   	fmt.Printf("%s",string(jsonOut))
	   return
	*/
}
