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
	Html    string
	Payload interface{}
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

type newGameHandler struct {
}

func (h *newGameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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

type joinGameHandler struct {
}

func (h *joinGameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	log.Printf("Handle add player.")

	pr := new(PlayerResp)

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

		pr.Html = string(file)

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

	}

	pr.Payload = p

	jsonOut, err := json.Marshal(pr)

	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Handler successful.")

	w.Write(jsonOut)
}

func pollGameHandler(w http.ResponseWriter, r *http.Request) {
}

func submitAnswerHandler(w http.ResponseWriter, r *http.Request) {
}

var Games map[string]*Game

func main() {

	Games = make(map[string]*Game)

	http.Handle("/play", new(htmlHandler))
	http.Handle("/newgame", new(newGameHandler))
	http.Handle("/joingame", new(joinGameHandler))
	http.Handle("/score", new(scoreHandler))
	//	http.Handle("/pollgame", pollGameHandler)
	//	http.Handle("/submitanswer", submitAnswerHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
