package handler

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"github.com/go-chi/chi/middleware"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/mochisuna/linebot-sample/domain"
)

// botのアクションのみを統括

// getMessageFollowAction はbotをフォローした際に実行されるアクション
func (s *Server) getMessageFollowAction(ctx context.Context, req *linebot.Event) linebot.SendingMessage {
	log.Println("called action.getMessageFollowAction")
	requestID := middleware.GetReqID(ctx)
	ownerID := domain.OwnerID(req.Source.UserID)
	profile, err := s.Bot.GetProfile(req.Source.UserID).Do()
	if err != nil {
		log.Printf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("プロフィール参照時にエラーが発生しました")
	}
	log.Printf("%v| DisplayName = %#v", requestID, profile.DisplayName)

	ref, err := s.CallbackService.Follow(ctx, ownerID)
	log.Printf("%v| %#v", requestID, ref)
	if err != nil {
		log.Printf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("登録時にエラーが発生しました")
	}
	return linebot.NewTextMessage(profile.DisplayName + "様。\n登録ありがとうございます。")
}

// isOwnerOfEvent は自分がオーナーのイベントがあるかどうかを返します
func (s *Server) isOwnerOfEvent(ownerID domain.OwnerID) (bool, error) {
	log.Println("called action.isOwnerOfEvent")
	_, err := s.CallbackService.GetEventByOwnerID(ownerID, domain.EVENT_OPEN)
	if err != nil {
		if err != sql.ErrNoRows {
			return false, err
		}
		// sql.ErrNoRows ならすぐに返却してよい
		return false, nil
	}
	return true, nil
}

// getMessageOpenEvent イベント開催アクション
func (s *Server) getMessageOpenEvent(ctx context.Context, req *linebot.Event) linebot.SendingMessage {
	log.Println("called action.getMessageOpenEvent")
	requestID := middleware.GetReqID(ctx)
	ownerID := domain.OwnerID(req.Source.UserID)

	owned, err := s.isOwnerOfEvent(ownerID)
	if err != nil {
		log.Printf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("イベント参照時にエラーが発生しました")
	}
	if owned {
		return linebot.NewTextMessage("あなたが主催のイベントが開催中です")
	}
	user, err := s.CallbackService.GetParticipatedEvent(domain.UserID(ownerID))
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("%v| error reason: %#v", requestID, err.Error())
			return linebot.NewTextMessage("参加イベント照会時にエラーが発生しました")
		}
	}
	if user.IsParticipated {
		log.Printf("%v| error in participated event: %#v", requestID, user.EventID)
		return linebot.NewTextMessage("あなたは既に別のイベントに参加しています")
	}
	_, err = s.CallbackService.GetEventByOwnerID(ownerID, domain.EVENT_STABDBY)
	if err != nil {
		if err == sql.ErrNoRows {
			// スタンバイ状態ですら存在しない場合はイベントを作成
			_, err = s.CallbackService.RegisterEvent(ctx, ownerID)
			if err != nil {
				log.Printf("%v| error reason: %#v", requestID, err.Error())
				return linebot.NewTextMessage("イベントスタンバイ時にエラーが発生しました")
			}
		} else {
			log.Printf("%v| error reason: %#v", requestID, err.Error())
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
	log.Println("called action.getMessageStartEvent")
	requestID := middleware.GetReqID(ctx)
	ownerID := domain.OwnerID(req.Source.UserID)
	owned, err := s.isOwnerOfEvent(ownerID)
	if err != nil {
		log.Printf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("イベント参照時にエラーが発生しました")
	}
	if owned {
		return linebot.NewTextMessage("あなたが主催のイベントが開催中です")
	}
	res, err := s.CallbackService.UpdateEventStatus(ctx, ownerID, domain.EVENT_OPEN)
	if err != nil {
		log.Printf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("ステータス更新時にエラーが発生しました")
	}
	msg := fmt.Sprintf("イベントを開催しました。\nイベント番号:\n%v\nを参加者に共有しましょう", res.ID)
	return linebot.NewTextMessage(msg)
}

func (s *Server) getMessageCloseEvent(ctx context.Context, req *linebot.Event) linebot.SendingMessage {
	log.Println("called action.getMessageCloseEvent")
	requestID := middleware.GetReqID(ctx)
	ownerID := domain.OwnerID(req.Source.UserID)
	owned, err := s.isOwnerOfEvent(ownerID)
	if err != nil {
		log.Printf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("イベント参照時にエラーが発生しました")
	}
	if !owned {
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
	log.Println("called action.getMessageFinishEvent")
	requestID := middleware.GetReqID(ctx)
	ownerID := domain.OwnerID(req.Source.UserID)
	owned, err := s.isOwnerOfEvent(ownerID)
	if err != nil {
		log.Printf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("イベント参照時にエラーが発生しました")
	}
	if !owned {
		return linebot.NewTextMessage("あなたはまだイベントを主催していません")
	}
	_, err = s.CallbackService.UpdateEventStatus(ctx, ownerID, domain.EVENT_CLOSED)
	if err != nil {
		log.Printf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("ステータス更新時にエラーが発生しました")
	}
	return linebot.NewTextMessage("イベントを終了しました")
}

func (s *Server) getMessageEvents(ctx context.Context, req *linebot.Event) linebot.SendingMessage {
	log.Println("called action.getMessageEvents")
	requestID := middleware.GetReqID(ctx)
	userID := domain.UserID(req.Source.UserID)
	owned, err := s.isOwnerOfEvent(domain.OwnerID(userID))
	if err != nil {
		log.Printf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("イベント参照時にエラーが発生しました")
	}
	if owned {
		return linebot.NewTextMessage("あなたが主催のイベントが開催中です")
	}
	user, err := s.CallbackService.GetParticipatedEvent(userID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("%v| error reason: %#v", requestID, err.Error())
			return linebot.NewTextMessage("参加イベント情報取得時にエラーが発生しました")
		}
	}
	if user.IsParticipated {
		log.Printf("%v| error in participated event: %#v", requestID, user.EventID)
		return linebot.NewTextMessage("あなたは既にどこかのイベントに参加しています")
	}

	events, err := s.CallbackService.GetActiveEvents()
	if err != nil {
		log.Printf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("開催イベント情報取得時にエラーが発生しました")
	} else if len(events) < 1 {
		return linebot.NewTextMessage("開催中のイベントが存在しません")
	}

	actions := []linebot.TemplateAction{}
	for _, ev := range events {
		if ev.Status != domain.EVENT_OPEN {
			continue
		}
		action := linebot.NewMessageAction(
			string(ev.ID),
			ActionEventParticipate+" "+string(ev.ID),
		)
		actions = append(actions, action)
	}

	return linebot.NewTemplateMessage(
		"start event",
		linebot.NewButtonsTemplate(
			"",
			"開催中のイベント",
			"参加したいイベントIDを選んでください",
			actions...,
		),
	)
}

func (s *Server) getMessageParticipateEvent(ctx context.Context, req *linebot.Event, eventID domain.EventID) linebot.SendingMessage {
	log.Println("called action.getMessageParticipateEvent")
	requestID := middleware.GetReqID(ctx)
	userID := domain.UserID(req.Source.UserID)
	owned, err := s.isOwnerOfEvent(domain.OwnerID(userID))
	if err != nil {
		log.Printf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("イベント参照時にエラーが発生しました")
	}
	if owned {
		return linebot.NewTextMessage("あなたが主催のイベントが開催中です")
	}
	event, err := s.CallbackService.GetEventByEventID(eventID)
	if err != nil {
		log.Printf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("開催イベント情報取得時にエラーが発生しました")
	}
	if event.Status == domain.EVENT_STABDBY {
		return linebot.NewTextMessage("このイベントはまだ開催していません")
	}
	if event.Status == domain.EVENT_CLOSED {
		return linebot.NewTextMessage("このイベントはすでに終了しています")
	}
	user, err := s.CallbackService.GetParticipatedEvent(userID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("%v| error reason: %#v", requestID, err.Error())
			return linebot.NewTextMessage("参加イベント情報取得時にエラーが発生しました")
		}
	}
	if user.EventID == event.ID {
		log.Printf("%v| error in participated event: %#v", requestID, event.ID)
		return linebot.NewTextMessage("あなたは既にこのイベントに参加しています")
	} else if user.IsParticipated {
		log.Printf("%v| error in participated event: %#v", requestID, event.ID)
		return linebot.NewTextMessage("あなたは既に別のイベントに参加しています")
	}

	if err = s.CallbackService.ParticipateEvent(ctx, &userID, &eventID); err != nil {
		log.Printf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("イベント参加時にエラーが発生しました")
	}
	return linebot.NewTextMessage("イベントに参加しました")
}

func (s *Server) getMessageLeaveEvent(ctx context.Context, req *linebot.Event) linebot.SendingMessage {
	log.Println("called action.getMessageLeaveEvent")
	requestID := middleware.GetReqID(ctx)
	userID := domain.UserID(req.Source.UserID)
	user, err := s.CallbackService.GetParticipatedEvent(userID)
	if err != nil {
		log.Printf("%v| error reason: %#v", requestID, err.Error())
		if err == sql.ErrNoRows {
			return linebot.NewTextMessage("あなたはまだイベントに参加していません")
		}
		return linebot.NewTextMessage("参加イベント情報取得時にエラーが発生しました")
	}
	if err = s.CallbackService.LeaveEvent(ctx, &userID, &user.EventID); err != nil {
		log.Printf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("イベント参加時にエラーが発生しました")
	}
	return linebot.NewTextMessage("イベントから離脱しました")

}

func voteString(vote domain.VOTE_STATUS) string {
	ret := ""
	switch vote {
	case domain.GREAT:
		ret = "よさみが深い"
	case domain.GOOD:
		ret = "よし"
	case domain.NOT_GOOD:
		ret = "まぁまぁ"
	case domain.BAD:
		ret = "わろし"
	}
	return ret
}

// getMessageOpenEvent イベント開催アクション
func (s *Server) getMessageVoteList(ctx context.Context, req *linebot.Event) linebot.SendingMessage {
	log.Println("called action.getMessageVoteList")
	requestID := middleware.GetReqID(ctx)
	userID := domain.UserID(req.Source.UserID)
	owned, err := s.isOwnerOfEvent(domain.OwnerID(userID))
	if err != nil {
		log.Printf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("イベント参照時にエラーが発生しました")
	}
	if owned {
		return linebot.NewTextMessage("あなたが主催のイベントが開催中です")
	}
	_, err = s.CallbackService.GetParticipatedEvent(userID)
	if err != nil {
		log.Printf("%v| error reason: %#v", requestID, err.Error())
		if err == sql.ErrNoRows {
			return linebot.NewTextMessage("あなたはまだイベントに参加していません")
		}
		return linebot.NewTextMessage("参加イベント情報取得時にエラーが発生しました")
	}

	return linebot.NewTemplateMessage(
		"vote event",
		linebot.NewButtonsTemplate(
			"",
			"投票",
			"このイベントについて投票します",
			linebot.NewMessageAction(voteString(domain.GREAT), fmt.Sprintf("voted %#v", domain.GREAT)),
			linebot.NewMessageAction(voteString(domain.GOOD), fmt.Sprintf("voted %#v", domain.GOOD)),
			linebot.NewMessageAction(voteString(domain.NOT_GOOD), fmt.Sprintf("voted %#v", domain.NOT_GOOD)),
			linebot.NewMessageAction(voteString(domain.BAD), fmt.Sprintf("voted %#v", domain.BAD)),
		),
	)
}

// getMessageOpenEvent イベント開催アクション
func (s *Server) getMessageVoteEvent(ctx context.Context, req *linebot.Event, votes string) linebot.SendingMessage {
	log.Println("called action.getMessageVoteEvent")
	requestID := middleware.GetReqID(ctx)
	userID := domain.UserID(req.Source.UserID)

	vote, err := strconv.Atoi(votes)
	if err != nil {
		log.Printf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("参加イベント情報取得時にエラーが発生しました")
	}
	owned, err := s.isOwnerOfEvent(domain.OwnerID(userID))
	if err != nil {
		log.Printf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("イベント参照時にエラーが発生しました")
	}
	if owned {
		return linebot.NewTextMessage("あなたが主催のイベントが開催中です")
	}
	user, err := s.CallbackService.GetParticipatedEvent(userID)
	if err != nil {
		log.Printf("%v| error reason: %#v", requestID, err.Error())
		if err == sql.ErrNoRows {
			return linebot.NewTextMessage("あなたはまだイベントに参加していません")
		}
		return linebot.NewTextMessage("参加イベント情報取得時にエラーが発生しました")
	}
	status := domain.VOTE_STATUS(vote)
	err = s.CallbackService.VoteEvent(ctx, &userID, &user.EventID, status)
	if err != nil {
		log.Printf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("投票時にエラーが発生しました")
	}

	return linebot.NewTextMessage(voteString(status) + "に投票しました")
}
