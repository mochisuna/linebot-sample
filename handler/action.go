package handler

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/go-chi/chi/middleware"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/mochisuna/linebot-sample/domain"
)

// botのアクションのみを統括

// getMessageFollowAction はbotをフォローした際に実行されるアクション
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

// isOwnerOfEvent は自分がオーナーのイベントがあるかどうかを返します
func (s *Server) isOwnerOfEvent(ownerID domain.OwnerID) (bool, error) {
	ref, err := s.CallbackService.ReferEventStatus(ownerID, domain.EVENT_OPEN)
	if err != nil {
		if err != sql.ErrNoRows {
			return false, err
		}
		// sql.ErrNoRows ならすぐに返却してよい
		return false, nil
	}
	return ref.OwnerID == ownerID, nil
}

// getMessageOpenEvent イベント開催アクション
func (s *Server) getMessageOpenEvent(ctx context.Context, req *linebot.Event) linebot.SendingMessage {
	requestID := middleware.GetReqID(ctx)
	ownerID := domain.OwnerID(req.Source.UserID)

	owned, err := s.isOwnerOfEvent(ownerID)
	if err != nil {
		log.Fatalf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("イベント参照時にエラーが発生しました")
	}
	if owned {
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
			log.Fatalf("%v| error reason: %#v", requestID, err.Error())
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
	owned, err := s.isOwnerOfEvent(ownerID)
	if err != nil {
		log.Fatalf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("イベント参照時にエラーが発生しました")
	}
	if owned {
		return linebot.NewTextMessage("あなたが主催のイベントが開催中です")
	}
	res, err := s.CallbackService.UpdateEventStatus(ctx, ownerID, domain.EVENT_OPEN)
	if err != nil {
		log.Fatalf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("ステータス更新時にエラーが発生しました")
	}
	msg := fmt.Sprintf("イベントを開催しました。\nイベント番号:\n%v\nを参加者に共有しましょう", res.ID)
	return linebot.NewTextMessage(msg)
}

func (s *Server) getMessageCloseEvent(ctx context.Context, req *linebot.Event) linebot.SendingMessage {
	requestID := middleware.GetReqID(ctx)
	ownerID := domain.OwnerID(req.Source.UserID)
	owned, err := s.isOwnerOfEvent(ownerID)
	if err != nil {
		log.Fatalf("%v| error reason: %#v", requestID, err.Error())
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
	requestID := middleware.GetReqID(ctx)
	ownerID := domain.OwnerID(req.Source.UserID)
	owned, err := s.isOwnerOfEvent(ownerID)
	if err != nil {
		log.Fatalf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("イベント参照時にエラーが発生しました")
	}
	if !owned {
		return linebot.NewTextMessage("あなたはまだイベントを主催していません")
	}
	_, err = s.CallbackService.UpdateEventStatus(ctx, ownerID, domain.EVENT_CLOSED)
	if err != nil {
		log.Fatalf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("ステータス更新時にエラーが発生しました")
	}
	return linebot.NewTextMessage("イベントを終了しました")
}

func (s *Server) getMessageParticipateEvent(ctx context.Context, req *linebot.Event) linebot.SendingMessage {
	requestID := middleware.GetReqID(ctx)
	userID := domain.UserID(req.Source.UserID)
	fmt.Println(userID)
	// owned, err := s.isOwnerOfEvent(domain.OwnerID(userID))
	// if err != nil {
	// 	log.Fatalf("%v| error reason: %#v", requestID, err.Error())
	// 	return linebot.NewTextMessage("イベント参照時にエラーが発生しました")
	// }
	// if owned {
	// 	return linebot.NewTextMessage("あなたが主催のイベントが開催中です")
	// }

	event, err := s.CallbackService.ReferEvent(domain.EventID(""))
	if err != nil {
		log.Fatalf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("イベント情報取得時にエラーが発生しました")
	}
	if event.Status == domain.EVENT_CLOSED {
		return linebot.NewTextMessage("このイベントはすでに終了しています")
	}
	if event.Status == domain.EVENT_CLOSED {
		return linebot.NewTextMessage("このイベントはすでに終了しています")
	}

	if err = s.CallbackService.ParticipateEvent(ctx, userID, domain.EventID("")); err != nil {
		log.Fatalf("%v| error reason: %#v", requestID, err.Error())
		return linebot.NewTextMessage("ステータス更新時にエラーが発生しました")
	}
	return linebot.NewTextMessage("イベントに参加しました")
}
