package main

import (
	"log"
	"regexp"
	"time"
)

func StripTags(content string) string {
	re := regexp.MustCompile(`<(.|\n)*?>`)
	return re.ReplaceAllString(content, "")
}

func Timeout(ch chan PlayerReq, id string, kind string, secs int) {
	time.Sleep(time.Duration(secs) * time.Second)
	req := PlayerReq{RequestType: kind, Payload: id, RespChan: make(chan PlayerResp, 2)}
	ch <- req
}

func PlayGame(ch chan PlayerReq, id string, accesscode string) {

	var presp PlayerResp

	go Timeout(ch, "", ReqTypeEnd, TimeoutGame)

	g := &Game{Xid: id, AccessCode: accesscode, Ch: ch}
	g.SetState(StateSetup)

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
				presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: "Joining game...", State: StateTemp, Payload: p.Xid}
			}
			break
		case ReqTypePoll:
			if g.CheckPermission(req.Token) {
				presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: g.ShowGame(req.Token), State: g.GetState(), Payload: "", ChatHtml: g.PlayerChat.GetMessages(req.Token)}
			}
			break
		case ReqTypeStart:
			// Avoid the case of someone starting the game after someone else already has.
			if g.GetState() == StateSetup {
				g.Start()
				g.PlayerChat.AddMessage(AdminName, "Let's play!")
				presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: "OK! Let's go!", State: StateTemp, Payload: ""}
			}
			break
		case ReqTypeSubmit:
			if g.GetState() == StatePoseQuestion {
				if req.Payload != "" {
					g.Submit(req.Token, req.Payload)
					presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: "Got it!", State: StateTemp, Payload: ""}
				}
			}
			break
		case ReqTypeAnswer:
			g.Answer(req.Token, req.Payload)
			presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: "Got it!", State: StateTemp, Payload: ""}
			break
		case ReqTypeNext:
			if g.GetState() == StateShowResults {
				g.PlayQuestion()
				if g.GetState() == StatePoseQuestion {
					presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: "Getting next question!", State: StateTemp, Payload: ""}
				} else {
					presp = PlayerResp{TimerHtml: g.GetTimer(), ScoreHtml: g.GetScores(), GameHtml: "All done!", State: StateTemp, Payload: ""}
				}
			}
			break
		case ReqTypeTimeoutQuestion:
			if g.GetState() == StatePoseQuestion {
				if g.CurrentQuestion.Xid == req.Payload {
					log.Printf("[Game %s] Timeout for submitting question %s", g.AccessCode, g.CurrentQuestion.Xid)
					g.CloseQuestionSubmissions()
					log.Printf("[Game %s] State %s", g.AccessCode, g.GetState())
					g.PlayerChat.AddMessage(AdminName, "Time's up! Let's see what you came up with.")

				}
			}
			break
		case ReqTypeTimeoutAnswer:
			if g.GetState() == StateOfferAnswers {
				if g.CurrentQuestion.Xid == req.Payload {
					log.Printf("[Game %s] Timeout for guessing question %s", g.AccessCode, g.CurrentQuestion.Xid)
					g.CloseGuessSubmissions()
					log.Printf("[Game %s] State %s", g.AccessCode, g.GetState())
					g.PlayerChat.AddMessage(AdminName, "Time's up! Let's see what the answers were.")
				}
			}
			break
		case ReqTypeTimeoutDisplay:
			if g.GetState() == StateShowResults {
				log.Printf("[Game %s] Timeout for showing results for question %s", g.AccessCode, g.CurrentQuestion.Xid)
				g.PlayQuestion()
			}
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
