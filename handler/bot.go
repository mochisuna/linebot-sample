package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/mochisuna/linebot-sample/config"
)

const (
	ActionEventOpen        = "open"
	ActionEventClose       = "close"
	ActionEventParticipate = "participate"
	ActionEventLeave       = "leave"
	ActionEventHelp        = "help"
	ActionEventVote        = "vote"
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
				case ActionEventOpen:
					response = s.getMessageOpenEvent(ctx, req)
				case ActionEventClose:
					response = s.getMessageCloseEvent(ctx, req)
				case ActionEventStart:
					response = s.getMessageStartEvent(ctx, req)
				case ActionEventFinish:
					response = s.getMessageFinishEvent(ctx, req)
				case ActionEventParticipate:
					response = s.getMessageParticipateEvent(ctx, req)
				case ActionEventLeave:
					response = linebot.NewTextMessage("TODO イベントに参加している場合のみ、イベントから離脱できるように変更")
				case ActionEventHelp:
					response = linebot.NewTextMessage(HelpMessage)
				case ActionEventVote:
					response = linebot.NewTextMessage("6")
				case ActionEventCancel:
					response = linebot.NewTextMessage("処理を中断しました")
				default:
					response = linebot.NewTextMessage(message.Text)
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
