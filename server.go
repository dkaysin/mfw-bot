package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	"gopkg.in/telegram-bot-api.v4"
)

const (
	MIN_PHOTO_SIZE = 1000 // in pixels
	whURL          = "https://mfw-bot.herokuapp.com"
	whExtPort      = "443"
)

var (
	whIntPort = os.Getenv("PORT")
)

func APIKey() string {
	return os.Getenv("MFWBOT_API_KEY")
}

func getChat(chatID int64) *Chat {
	chat, exists := Chats[chatID]
	if !exists {
		chat = &Chat{
			ID:    chatID,
			Users: UserList{},
			Queue: UserList{},
			Brawl: UserList{},
		}
		Chats[chatID] = chat
		go chatActor(chat, chat.CreateListener()) //[TODO] Rewrite as method of Chat?
	}
	return chat
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

func handleRequest(r Actioner) {
	c := getChat(r.GetChatID())
	a := &Action{}
	r.GetAction(a, c)
	if a.Type == "" {
		return
	}
	c.SendToListeners(a)
}

var (
	Dict  = Data{}
	Chats = make(map[int64]*Chat)
	Bot   *tgbotapi.BotAPI
)

func main() {

	rand.Seed(time.Now().Unix())
	GetData(&Dict)

	var updates <-chan tgbotapi.Update
	if os.Getenv("GET_UPDATE_METHOD") == "webhook" {
		updates = connectWebHook()
	} else {
		updates = connectLongPolling()
	}

	for update := range updates {

		var a Actioner

		if msg := update.Message; msg != nil {
			a = Actioner(TGMessage(*msg))
		}
		if clb := update.CallbackQuery; clb != nil {
			a = Actioner(TGCallback(*clb))
		}
		if a != nil {
			go handleRequest(a)
		}
	}
}
