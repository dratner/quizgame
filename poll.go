package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
)

type GamePoll struct {
	ScoreHtml string
	TimerHtml string
	State     string
	GameHtml  string
}

type pollHandler struct {
}

func (h *pollHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	log.Printf("Handle score request.")

	decoder := json.NewDecoder(r.Body)

	var preq *PlayerReq
	err := decoder.Decode(&preq)

	presp := new(GamePoll)

	if _, ok := Games[preq.AccessCode]; ok {
		var payload []*Player
		for _, p := range Games[preq.AccessCode].Players {
			payload = append(payload, p)
			presp.ScoreHtml += p.Name + ": " + strconv.Itoa(p.Score) + "&nbsp;&nbsp;&nbsp;"
		}
		presp.State = Games[preq.AccessCode].GetState()
		if Games[preq.AccessCode].Timeout() < 0 {
			presp.TimerHtml = ""
		} else {
			presp.TimerHtml = fmt.Sprintf("%d seconds", Games[preq.AccessCode].Timeout())
		}
		presp.GameHtml = Games[preq.AccessCode].ShowGame(preq.Xid)
	} else {
		err = errors.New("A game with that access code does not exist.")
	}

	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
		return
	}

	jsonOut, err := json.Marshal(presp)

	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Handler successful.")

	// THIS IS DEBUG CODE ONLY

	for _, p := range Games[preq.AccessCode].Players {
		p.Score += rand.Int()
	}

	// DELETE FOR PROD

	w.Write(jsonOut)
}
