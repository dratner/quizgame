package main

import (
	"fmt"
	"log"
	"regexp"
	"time"
)

func StripTags(content string) string {
	re := regexp.MustCompile(`<(.|\n)*?>`)
	return re.ReplaceAllString(content, "")
}

func TimeoutQuestion(ch chan PlayerReq, id string) {
	time.Sleep(QuestionTimeout * time.Second)
	req := PlayerReq{RequestType: ReqTypeTimeout, Payload: id, RespChan: make(chan PlayerResp, 2)}
	ch <- req
}

func TimeoutGame(ch chan PlayerReq) {
	time.Sleep(24 * time.Hour)
	req := PlayerReq{RequestType: ReqTypeEnd}
	ch <- req
}

func PlayGame(ch chan PlayerReq, id string, accesscode string) {

	var presp PlayerResp

	go TimeoutGame(ch)

	g := &Game{Xid: id, AccessCode: accesscode, State: StateSetup}

	for {

		// A requests are serialized by the channel, so mutex not required.

		req := <-ch

		presp = PlayerResp{TimerHtml: "", ScoreHtml: "", GameHtml: "", State: g.GetState(), ChatHtml: "", Payload: ""}

		if req.RequestType != ReqTypePoll {
			log.Printf("[Game %s] Processing %s request.", g.AccessCode, req.RequestType)
		}

		if req.Payload != "" {
			req.Payload = StripTags(req.Payload)
		}

		switch req.RequestType {
		case ReqTypeChat:
			if g.CheckPermission(req.Token) {
				g.PlayerChat.AddMessage(g.PlayerByXid(req.Token).Name, req.Payload)
				log.Printf("[Game %s] Chat request by player: %s", g.AccessCode, g.PlayerByXid(req.Token).Name)
				presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: g.ShowGame(req.Token), State: g.GetState(), Payload: "", ChatHtml: g.PlayerChat.GetMessages(req.Token)}
			}
			break
		case ReqTypeJoin:
			p, err := g.AddPlayer(req.Payload)
			if err != nil {
				log.Printf("[Game %s] Error: %s", g.AccessCode, err)
			} else {
				presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: "", State: g.GetState(), Payload: p.Xid}
			}
			break
		case ReqTypePoll:
			if g.CheckPermission(req.Token) {
				presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: g.ShowGame(req.Token), State: g.GetState(), Payload: "", ChatHtml: g.PlayerChat.GetMessages(req.Token)}
			}
			break
		case ReqTypeStart:
			g.Start()
			g.PlayerChat.AddMessage(AdminName, "Let's play!")
			if g.State == StatePoseQuestion {
				go TimeoutQuestion(ch, g.CurrentQuestion.Xid)
			}
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
				g.PlayerChat.AddMessage(AdminName, fmt.Sprintf("There are %d books left!", len(g.Questions)))
				g.PlayQuestion()
				if g.State == StatePoseQuestion {
					go TimeoutQuestion(ch, g.CurrentQuestion.Xid)
					presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: "Getting next question!", State: StateTemp, Payload: ""}
				} else {
					presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: "All done!", State: StateTemp, Payload: ""}
				}
			}
			break
		case ReqTypeTimeout:
			if g.State == StatePoseQuestion {
				if g.CurrentQuestion.Xid == req.Payload {
					log.Printf("[Game %s] Timeout for question %s", g.AccessCode, g.CurrentQuestion.Xid)
					g.CloseQuestionSubmissions()
					log.Printf("[Game %s] State %s", g.AccessCode, g.State)
				}
			}
			break
		case ReqTypeEnd:
			presp = PlayerResp{}
			req.RespChan <- presp
			close(req.RespChan)
			return
		default:
			log.Printf("[Game %s] Unidentified request.", g.AccessCode)
			break
		}
		req.RespChan <- presp
		close(req.RespChan)
	}
}
