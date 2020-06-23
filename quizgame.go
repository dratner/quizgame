package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type PlayerReq struct {
	Xid        string
	AccessCode string
	Payload    interface{}
}

type PlayerResp struct {
	Html          string
	TimeRemaining int
	Payload       interface{}
}

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

	log.Printf("Handle create new game.")

	pr := new(PlayerResp)

	g := NewGame()
	Games[g.AccessCode] = g

	pr.Payload = g
	pr.Html = fmt.Sprintf("You made a new game with access code %s. Invite your friends!", g.AccessCode)

	jsonOut, err := json.Marshal(pr)

	if err != nil {
		log.Printf("JSON error.")
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Created new game with xid %s.", g.Xid)
	log.Printf("Handler successful.")

	w.Write(jsonOut)
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

func main() {

	Games = make(map[string]*Game)
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
	http.Handle("/play", new(htmlHandler))
	http.Handle("/new", new(newHandler))
	http.Handle("/join", new(joinHandler))
	http.Handle("/start", new(startHandler))
	http.Handle("/poll", new(pollHandler))
	//	http.Handle("/submit", new(submitHandler))
	//	http.Handle("/guess", new(guessHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
