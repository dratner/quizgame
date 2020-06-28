package main

import (
	"encoding/json"
	"fmt"
	"github.com/rs/xid"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type Conf struct {
	Addr     string
	GamePath string
	LogFile  string
}

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
var Config Conf

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

	presp := &PlayerResp{TimerHtml: "", ScoreHtml: "", GameHtml: fmt.Sprintf("You made a new game with access code %s. Invite your friends! Click Join Game to continue.", ac), State: StateSetup, Payload: ac}

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

	//log.Printf("Req handler.")

	decoder := json.NewDecoder(r.Body)

	var preq *PlayerReq
	err := decoder.Decode(&preq)

	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
		return
	}

	if _, ok := Games[preq.AccessCode]; !ok {
		log.Printf("Error: Game with access code %s does not exist.", preq.AccessCode)
		http.Error(w, fmt.Sprintf("Error: Game with access code %s does not exist.", preq.AccessCode), http.StatusInternalServerError)
		return
	}

	//log.Printf("Handling %s request to game %s.", preq.RequestType, preq.AccessCode)

	preq.RespChan = make(chan PlayerResp)

	Games[preq.AccessCode] <- *preq

	var presp PlayerResp

	presp = <-preq.RespChan

	if preq.RequestType == ReqTypeEnd {
		delete(Games, preq.AccessCode)
	}

	jsonOut, err := json.Marshal(presp)

	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
		return
	}

	w.Write(jsonOut)
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Please include a configuration file (e.g. ./quizgame conf/conf.json)")
		return
	}

	f := os.Args[1]

	file, err := ioutil.ReadFile(f)

	if err != nil {
		log.Fatal(err)
	}

	Config := Conf{}

	err = json.Unmarshal([]byte(file), &Config)

	if err != nil {
		log.Fatalf("Error opening conf file %v", err)
	}

	if Config.LogFile != "" {
		flog, err := os.OpenFile(Config.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Error opening log file: %v", err)
		}
		defer flog.Close()

		log.SetOutput(flog)
		log.Println("Log redirected.")
	}

	log.Println("Config loaded.")

	Games = make(map[string]chan PlayerReq)

	http.Handle("/", new(htmlHandler))
	http.Handle("/new", new(newHandler))
	http.Handle("/req", new(reqHandler))

	log.Fatal(http.ListenAndServe(Config.Addr, nil))
}
