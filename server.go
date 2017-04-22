package main

import (
	"math/rand"
	"os"
	"time"

	"gopkg.in/telegram-bot-api.v4"
)

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
			a = Actioner((*TGMessage)(msg))
		}
		if clb := update.CallbackQuery; clb != nil {
			a = Actioner((*TGCallback)(clb))
		}
		if a == nil {
			continue
		}
		go HandleRequest(a)
	}
}
