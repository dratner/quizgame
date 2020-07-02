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

		// A requests are serialized by the channel, so mutex not required.

		req := <-ch

		presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: "", State: g.GetState(), Payload: ""}

		if req.RequestType != ReqTypePoll {
			log.Printf("Processing %s request for game %s and player %s.", req.RequestType, req.AccessCode, req.Token)
		}

		switch req.RequestType {
		case ReqTypeJoin:
			p, err := g.AddPlayer(req.Payload)
			if err != nil {
				log.Printf("Error: %s", err)
			} else {
				presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: "", State: g.GetState(), Payload: p.Xid}
			}
			break
		case ReqTypePoll:
			if g.CheckPermission(req.Token) {
				presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: g.ShowGame(req.Token), State: g.GetState(), Payload: ""}
			}
			break
		case ReqTypeStart:
			g.Start()
			presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: "OK! Let's go!", State: StateTemp, Payload: ""}
			break
		case ReqTypeSubmit:
			g.Submit(req.Token, req.Payload)
			presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: "Got it!", State: StateTemp, Payload: ""}
			break
		case ReqTypeAnswer:
			g.Answer(req.Token, req.Payload)
			presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: "Got it!", State: StateTemp, Payload: ""}
			break
		case ReqTypeNext:
			if g.State == StateShowResults {
				g.PlayQuestion()
				if g.State == StatePoseQuestion {
					presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: "Getting next question!", State: StateTemp, Payload: ""}
				} else {
					presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: "All done!", State: StateTemp, Payload: ""}
				}
			}
			break
		case ReqTypeTimeout:
			//TODO
			break
		case ReqTypeEnd:
			presp = PlayerResp{}
			req.RespChan <- presp
			close(req.RespChan)
			return
		default:
			log.Printf("Unidentified request.")
			break
		}
		req.RespChan <- presp
		close(req.RespChan)
	}
}
