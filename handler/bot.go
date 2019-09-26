package handler

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/mochisuna/linebot-sample/config"
	"github.com/mochisuna/linebot-sample/domain"
)

const (
	ActionEventOpen        = "open"
	ActionEventClose       = "close"
	ActionEventList        = "list"
	ActionEventParticipate = "participate"
	ActionEventLeave       = "leave"
	ActionEventHelp        = "help"
	ActionEventVote        = "vote"
	ActionEventVoted       = "voted"
	ActionEventStart       = "start"
	ActionEventFinish      = "finish"
	ActionEventCancel      = "cancel"
)

// TODO ファイルから読み出すように変更
const HelpMessage = "このbotについて\nこのbotはLT会等で、参加者からアンケートを募集することを目的に作られています。\n\n以下のアクション一覧から利用したいコマンドを実行してください。"

type Line struct {
	Bot *linebot.Client
}

// New inject to domain services
func NewLineBot(config *config.Line) *Line {
	client, err := linebot.New(config.ChannelSecret, config.ChannelToken)
	if err != nil {
		log.Fatal(err)
	}

	return &Line{client}
}

func (s *Server) callback(w http.ResponseWriter, r *http.Request) {
	log.Println("callback")
	ctx := r.Context()
	reqests, err := s.Bot.ParseRequest(r)
	for _, req := range reqests {
		fmt.Printf("%#v\n", req)
		var response linebot.SendingMessage
		switch req.Type {
		case linebot.EventTypeMessage:
			switch message := req.Message.(type) {
			case *linebot.TextMessage:
				switch message.Text {
				// リッチメニューボタン
				case ActionEventOpen:
					response = s.getMessageOpenEvent(ctx, req)
				case ActionEventClose:
					response = s.getMessageCloseEvent(ctx, req)
				case ActionEventList:
					response = s.getMessageEvents(ctx, req)
				case ActionEventVote:
					response = s.getMessageVoteList(ctx, req)
				case ActionEventLeave:
					response = s.getMessageLeaveEvent(ctx, req)
				case ActionEventHelp:
					response = linebot.NewTextMessage(HelpMessage)
				// 確認処理ボタン
				case ActionEventStart:
					response = s.getMessageStartEvent(ctx, req)
				case ActionEventFinish:
					response = s.getMessageFinishEvent(ctx, req)
				case ActionEventCancel:
					response = linebot.NewTextMessage("処理を中断しました")
				default:
					if strings.Contains(message.Text, ActionEventParticipate) {
						splits := strings.Split(message.Text, " ")
						log.Println(message.Text)
						eventID := domain.EventID(splits[1])
						response = s.getMessageParticipateEvent(ctx, req, eventID)
					} else if strings.Contains(message.Text, ActionEventVoted) {
						splits := strings.Split(message.Text, " ")
						response = s.getMessageVoteEvent(ctx, req, splits[1])
					} else {
						response = linebot.NewTextMessage(message.Text)
					}
				}
			}
		case linebot.EventTypeFollow:
			response = s.getMessageFollowAction(ctx, req)
		}

		// 全処理をここで一括
		if _, err = s.Bot.ReplyMessage(req.ReplyToken, response).Do(); err != nil {
			fmt.Println(err)
		}
	}
}
