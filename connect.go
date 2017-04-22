package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/telegram-bot-api.v4"
)

func connectWebHook() <-chan tgbotapi.Update {

	log.Printf("Connecting via Webhook using API key: %v", APIKey())

	var updatesC <-chan tgbotapi.Update

	for {
		b, err := tgbotapi.NewBotAPI(APIKey())
		if err != nil {
			log.Println("[server] No connection. Reconnecting after timeout...")
			time.Sleep(5 * time.Second)
		} else {
			Bot = b
			break
		}
	}

	log.Printf("[server] Authorized on account %s", Bot.Self.UserName)

	log.Printf("[server] Setting up a webhook on port %v->%s", WEBHOOK_EXT_PORT, os.Getenv("PORT"))

	_, err := Bot.SetWebhook(tgbotapi.NewWebhook(WEBHOOK_URL + ":" + WEBHOOK_EXT_PORT + "/" + Bot.Token))
	if err != nil {
		log.Fatal(err)
	}

	updatesC = Bot.ListenForWebhook("/" + Bot.Token)

	go http.ListenAndServe(":"+WEBHOOK_INT_PORT, nil)

	return updatesC
}

func connectLongPolling() <-chan tgbotapi.Update {

	log.Printf("Connecting via Long polling using API key: %v", APIKey())

	var updatesC <-chan tgbotapi.Update

	for {
		b, err := tgbotapi.NewBotAPI(APIKey())
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

	Bot.RemoveWebhook()
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
