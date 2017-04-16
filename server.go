package main

import (
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"math/rand"
	"time"
	"os"
	"net/http"
)

const (
	MIN_PHOTO_SIZE  = 1000 // in pixels
)

func API_KEY() string {
	return os.Getenv("MFWBOT_API_KEY")
}

func connectWebHook() <-chan tgbotapi.Update {

	log.Printf("Using API key: %v", API_KEY())

	var updatesC <-chan tgbotapi.Update

	for {
		b, err := tgbotapi.NewBotAPI(API_KEY())
		if err != nil {
			log.Println("[server] No connection. Reconnecting after timeout...")
			time.Sleep(5 * time.Second)
		} else {
			Bot = b
			break
		}
	}

	log.Printf("[server] Authorized on account %s", Bot.Self.UserName)

	log.Printf("[server] Setting up a webhook on port %s", os.Getenv("PORT"))

	_, err := Bot.SetWebhook(tgbotapi.NewWebhook("https://mfw-bot.herokuapp.com:80"+"/"+Bot.Token))
	if err != nil {
		log.Fatal(err)
	}


	updatesC = Bot.ListenForWebhook("/" + Bot.Token)

	go http.ListenAndServe(":"+os.Getenv("PORT"), nil)

	return updatesC
}

func connectLongPolling() <-chan tgbotapi.Update {

	log.Printf("Using API key: %v", API_KEY())

	var updatesC <-chan tgbotapi.Update

	for {
		b, err := tgbotapi.NewBotAPI(API_KEY())
		if err != nil {
			log.Println("[server] No connection. Reconnecting after timeout...")
			time.Sleep(5 * time.Second)
		} else {
			Bot = b
			break
		}
	}

	log.Printf("[server] Authorized on account %s", Bot.Self.UserName)

	// Bot.Debug = true
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	for {
		u, err := Bot.GetUpdatesChan(updateConfig)
		if err != nil {
			log.Println("[server] Unable to set updates handle. Reconnecting after timeout...")
			time.Sleep(5 * time.Second)
		} else {
			updatesC = u
			break
		}
	}

	return updatesC
}

func getChat(chatID int64) *Chat {
	chat, exists := Chats[chatID]
	if !exists {
		chat = &Chat{
			Id:    chatID,
			Users: UserList{},
			Queue: UserList{},
			Brawl: UserList{},
		}
		Chats[chatID] = chat
		go chatActor(chat, chat.CreateListener()) //[TODO] Rewrite as method of Chat?
	}
	return chat
}

func handleMessage(msg *tgbotapi.Message) {

	log.Printf("[%s] Message received: %s", msg.From.UserName, msg.Text)

	action := Action{
		Msg: msg,
	}

	if entities := msg.Entities; entities != nil {
		for _, e := range *entities {
			switch e.Type {
			case "bot_command":
				cmd := parseCommandFromMsg(e.Offset, e.Length, msg.Text)
				switch cmd {
				case "start":
					action.Type = "start"
				case "debug":
					action.Type = "debug"
				case "data":
					action.Type = "data"
				case "quit":
					action.Type = "quit"
				case "fight":
					action.Type = "fight"
				default:
				}
			default:
			}
		}
	}

	if photos := msg.Photo; photos != nil {
		for _, p := range *photos {
			if p.Width*p.Height > MIN_PHOTO_SIZE {
				action.Type = "photo"
			}
		}
	}

	if action.Type != "" {
		chat := getChat(msg.Chat.ID)
		chat.SendToListeners(&action)
	}
}

func handleCallback(clb *tgbotapi.CallbackQuery) {

	log.Printf("[%s] Callback received: %s", clb.From.UserName, clb.ID)

	action := Action{
		Clb: clb,
	}

	switch clb.Data {
	case "fight":
		action.Type = "fight"
		clbCfg := tgbotapi.CallbackConfig{CallbackQueryID: action.Clb.ID}
		Bot.AnswerCallbackQuery(clbCfg)
	case "help":
		action.Type = "help"
	default:
		action.Type = "vote"
	}

	// log.Printf("%v   %v", clb.Message.Chat.ID, clb.From.UserName)

	if action.Type != "" {
		chat := getChat(clb.Message.Chat.ID)
		chat.SendToListeners(&action)
	}
}

var (
	Dict  = Data{}
	Chats = make(map[int64]*Chat)
	Bot   *tgbotapi.BotAPI
)

func main() {

	rand.Seed(time.Now().Unix())
	GetData(&Dict)

	updates := connectWebHook()

	for update := range updates {
		if msg := update.Message; msg != nil {
			go handleMessage(msg)
		}
		if clb := update.CallbackQuery; clb != nil {
			go handleCallback(clb)
		}
	}
}
