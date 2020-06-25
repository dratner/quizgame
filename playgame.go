package main

import (
	"log"
)

func Timeout(ch chan PlayerReq, id string) {
	req := PlayerReq{RequestType: "Timeout", Payload: id}
	ch <- req
}

func PlayGame(ch chan PlayerReq, id string, accesscode string) {

	var presp PlayerResp

	p := make(map[string]*Player)
	g := &Game{Xid: id, AccessCode: accesscode, State: StateSetup, Players: p}

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
		case ReqTypeSubmit:
			log.Printf("Getting submission from %s", req.Xid)
			g.Submit(req.Xid, req.Payload)
			presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: "Got it!", State: StateTemp, Payload: ""}
			req.RespChan <- presp
			break
		case ReqTypeAnswer:
			log.Printf("Accepting answer from %s", req.Xid)
			g.Answer(req.Xid, req.Payload)
			presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: "Got it!", State: StateTemp, Payload: ""}
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
