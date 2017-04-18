package main

import (
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

	Dict  = Data{}
	Chats = make(map[int64]*Chat)
	Bot   *tgbotapi.BotAPI
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

func handleRequest(r Actioner) {
	c := getChat(r.GetChatID())
	a := &Action{}
	r.GetAction(a, c)
	if a.Type == "" {
		return
	}
	c.SendToListeners(a)
}

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
			a = Actioner((*TGMessage)(msg))
		}
		if clb := update.CallbackQuery; clb != nil {
			a = Actioner((*TGCallback)(clb))
		}
		if a != nil {
			go handleRequest(a)
		}
	}
}
