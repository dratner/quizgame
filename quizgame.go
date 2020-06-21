package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

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

	g := NewGame()
	Games[g.AccessCode] = g

	jsonOut, err := json.Marshal(g)

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
	decoder := json.NewDecoder(r.Body)

	var p *Player
	err := decoder.Decode(&p)

	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
		return
	}

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

	jsonOut, err := json.Marshal(p)

	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Created new player with xid %s.", p.Xid)
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
	//	http.Handle("/pollgame", pollGameHandler)
	//	http.Handle("/submitanswer", submitAnswerHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
	/*
		r := mux.NewRouter()
		r.HandleFunc("/newgame", newGameHandler)
		r.HandleFunc("/joingame", joinGameHandler)
		r.HandleFunc("/pollgame", pollGameHandler)
		r.HandleFunc("/submitanswer", submitAnswerHandler)

		http.Handle("/", r)

			fmt.Println("Mr Mark Twain")

			var err error

			g1 := new(Game)
			g1.AccessCode = "ABC123"
			err = g1.FromFile("g1.json")
			if err == nil {
				fmt.Println ("Loaded up and ready to go")
			} else {
				fmt.Print ("Error: %s", err )
			}

			dbSelect()

			fmt.Println("DB connected!")

			fmt.Println(g1.AccessCode)
	*/
}
