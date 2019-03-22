package handler

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/mochisuna/linebot-sample/config"
	"github.com/mochisuna/linebot-sample/domain"
)

const (
	ActionEventOpen   = "open"
	ActionEventClose  = "close"
	ActionEventJoin   = "join"
	ActionEventLeave  = "leave"
	ActionEventHelp   = "help"
	ActionEventVote   = "vote"
	ActionEventStart  = "start"
	ActionEventFinish = "finish"
	ActionEventCancel = "cancel"
)

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
				case ActionEventJoin:
					response = linebot.NewTextMessage("3")
				case ActionEventLeave:
					response = linebot.NewTextMessage("4")
				case ActionEventHelp:
					response = linebot.NewTextMessage("5")
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
func (s *Server) getMessageFollowAction(ctx context.Context, req *linebot.Event) linebot.SendingMessage {
	ownerID := domain.OwnerID(req.Source.UserID)
	requestID := middleware.GetReqID(ctx)
	profile, err := s.Bot.GetProfile(req.Source.UserID).Do()
	if err != nil {
		log.Fatalf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("プロフィール参照時にエラーが発生しました")
	}
	log.Printf("%v| DisplayName = %#v", requestID, profile.DisplayName)

	ref, err := s.CallbackService.Follow(ctx, ownerID)
	log.Printf("%v| %#v", requestID, ref)
	if err != nil {
		log.Fatalf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("登録時にエラーが発生しました")
	}
	return linebot.NewTextMessage(profile.DisplayName + "様。\n登録ありがとうございます。")
}

func (s *Server) getMessageOpenEvent(ctx context.Context, req *linebot.Event) linebot.SendingMessage {
	requestID := middleware.GetReqID(ctx)
	ownerID := domain.OwnerID(req.Source.UserID)
	ref, err := s.CallbackService.ReferEventStatus(ownerID, domain.EVENT_OPEN)
	if err != nil {
		if err != sql.ErrNoRows {
			return linebot.NewTextMessage("イベント参照時にエラーが発生しました")
		}
	} else if ref.OwnerID == ownerID {
		return linebot.NewTextMessage("あなたが主催のイベントが開催中です")
	}
	_, err = s.CallbackService.ReferEventStatus(ownerID, domain.EVENT_STABDBY)
	if err != nil {
		if err == sql.ErrNoRows {
			// スタンバイ状態ですら存在しない場合はイベントを作成
			_, err = s.CallbackService.RegisterEvent(ctx, ownerID)
			if err != nil {
				log.Fatalf("%v| error reason: %#v", requestID, err.Error())
				return linebot.NewTextMessage("イベントスタンバイ時にエラーが発生しました")
			}
		} else {
			return linebot.NewTextMessage("イベント参照時にエラーが発生しました")
		}
	}

	return linebot.NewTemplateMessage(
		"start event",
		linebot.NewConfirmTemplate(
			"イベントを開催しますか？",
			linebot.NewMessageAction("開催する", ActionEventStart),
			linebot.NewMessageAction("戻る", ActionEventCancel),
		),
	)
}

func (s *Server) getMessageStartEvent(ctx context.Context, req *linebot.Event) linebot.SendingMessage {
	requestID := middleware.GetReqID(ctx)
	ownerID := domain.OwnerID(req.Source.UserID)
	_, err := s.CallbackService.UpdateEventStatus(ctx, ownerID, domain.EVENT_OPEN)
	if err != nil {
		log.Fatalf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("ステータス更新時にエラーが発生しました")
	}
	return linebot.NewTextMessage("イベントを開催しました")
}

func (s *Server) getMessageCloseEvent(ctx context.Context, req *linebot.Event) linebot.SendingMessage {
	requestID := middleware.GetReqID(ctx)
	ownerID := domain.OwnerID(req.Source.UserID)
	_, err := s.CallbackService.ReferEventStatus(ownerID, domain.EVENT_OPEN)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Fatalf("%v| error reason: %#v", requestID, err.Error())
			return linebot.NewTextMessage("イベント参照時にエラーが発生しました")
		}
		return linebot.NewTextMessage("あなたはまだイベントを主催していません")
	}
	return linebot.NewTemplateMessage(
		"start event",
		linebot.NewConfirmTemplate(
			"イベントを終了しますか？",
			linebot.NewMessageAction("終了する", ActionEventFinish),
			linebot.NewMessageAction("戻る", ActionEventCancel),
		),
	)
}

func (s *Server) getMessageFinishEvent(ctx context.Context, req *linebot.Event) linebot.SendingMessage {
	requestID := middleware.GetReqID(ctx)
	ownerID := domain.OwnerID(req.Source.UserID)
	_, err := s.CallbackService.UpdateEventStatus(ctx, ownerID, domain.EVENT_CLOSED)
	if err != nil {
		log.Fatalf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("ステータス更新時にエラーが発生しました")
	}
	return linebot.NewTextMessage("イベントを終了しました")
}
