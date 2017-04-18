package main

import (
	"log"

	"gopkg.in/telegram-bot-api.v4"
)

type TGMessage tgbotapi.Message
type TGCallback tgbotapi.CallbackQuery

type Actioner interface {
	GetAction(a *Action, c *Chat)
	GetChatID() int64
}

func (msg TGMessage) GetAction(a *Action, c *Chat) {

	if entities := msg.Entities; entities != nil {
		for _, e := range *entities {
			switch e.Type {
			case "bot_command":
				f := e.Offset
				l := e.Length
				cmd := msg.Text[f+1 : f+l]
				switch cmd {
				case "start":
					a.Type = "start"
				case "debug":
					a.Type = "debug"
				case "data":
					a.Type = "data"
				case "quit":
					a.Type = "quit"
				case "fight":
					a.Type = "fight"
				}
			}
		}
	}

	if photos := msg.Photo; photos != nil {
		for _, p := range *photos {
			if p.Width*p.Height > MIN_PHOTO_SIZE {
				a.Type = "photo"
			}
		}
	}

	if a.Type == "" {
		return
	}

	a.From = c.GetUser(msg.From)
	a.Message = &Message{
		ID:   msg.MessageID,
		From: c.GetUser(msg.From),
		Text: msg.Text,
	}

	log.Printf("[%s] Message received: %s", a.From.Username, a.Message.Text)
	return
}

func (clb TGCallback) GetAction(a *Action, c *Chat) {

	switch clb.Data {
	case "fight":
		a.Type = "fight"
	case "help":
		a.Type = "help"
	default:
		for k, _ := range VoteMap {
			if k == clb.Data {
				a.Type = "vote"
			}
		}
	}

	if a.Type == "fight" || a.Type == "help" {
		clbCfg := tgbotapi.CallbackConfig{CallbackQueryID: clb.ID}
		Bot.AnswerCallbackQuery(clbCfg)
	}

	if a.Type == "" {
		return
	}

	a.From = c.GetUser(clb.From)
	a.Message = &Message{
		ID:   clb.Message.MessageID,
		From: c.GetUser(clb.Message.From),
		Text: clb.Message.Text,
	}
	a.Callback = &Callback{
		ID:   clb.ID,
		Data: clb.Data,
	}
	if reply := clb.Message.ReplyToMessage; reply != nil {
		a.Message.ReplyToMsg = &Message{
			ID:   reply.MessageID,
			From: c.GetUser(reply.From),
		}
	}

	log.Printf("[%s] Callback received: %s", clb.From.UserName, clb.ID)
	return

}

func (msg TGMessage) GetChatID() int64 {
	return msg.Chat.ID
}

func (clb TGCallback) GetChatID() int64 {
	return clb.Message.Chat.ID
}
