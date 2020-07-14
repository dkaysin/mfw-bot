package main

import (
	"fmt"
	"log"
	"time"

	"gopkg.in/telegram-bot-api.v4"
)

func ChatActor(chat *Chat, c chan *Action) {

	defer log.Printf("[server] Exiting chatActor goroutine for chat %v", chat.ID)
	defer chat.Delete()

	log.Printf("[server] Starting chatActor goroutine for chat %v", chat.ID)

	timer := time.NewTimer(CHAT_TIMEOUT)

	for {
		select {
		case action := <-c:

			usersCount, _ := Bot.GetChatMembersCount(tgbotapi.ChatConfig{chat.ID, ""})
			chat.MaxBrawlUserCount = MaxInt(
				MIN_BRAWL_USERS,
				MinInt(
					MAX_BRAWL_USERS,
					int((float64(usersCount-1)*COEFF_BRAWL_USERS))+1,
				),
			)

			actionMap := mapActionChat(action, chat)
			f, exists := actionMap[action.Type]
			if exists {
				f()
			}
			// if action.Type != "default" {
			timer.Reset(CHAT_TIMEOUT)
			// }
		case <-timer.C:
			log.Printf("[server] Chat timeouted")
			return
		}
	}
}

func mapActionChat(action *Action, chat *Chat) map[string]func() {

	user := action.From
	uID := user.ID

	var (
		start = func() {
			Bot.Send(GetTxtMsg(chat.ID, fmt.Sprintf("Hello, I am 'My Face When...'-bot.")))
			time.Sleep(DELAY_TEXT)
			Bot.Send(GetTxtMsg(chat.ID, fmt.Sprintf("I'm dying to play a game with you")))
			time.Sleep(DELAY_TEXT)
			Bot.Send(GetTxtMsg(chat.ID, "_[AGITATED BEEP]_"))
			time.Sleep(DELAY_TEXT)

			msg := GetTxtMsg(chat.ID, "Just type /fight or choose one of the options below:")

			kb := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						"Start the fight!",
						"fight"),
					tgbotapi.NewInlineKeyboardButtonData(
						"Help",
						"help"),
				))
			msg.ReplyMarkup = kb
			Bot.Send(msg)
		}

		help = func() {
			Bot.Send(GetTxtMsg(chat.ID, fmt.Sprintf("Unfortunately, the help has not been implemented yet")))
			time.Sleep(DELAY_TEXT)
			Bot.Send(GetTxtMsg(chat.ID, fmt.Sprintf("So... Bad luck!")))
		}

		debug = func() {
			log.Printf("[debug] Current queue: %+v\n", chat.Queue)
			log.Printf("[debug] Current brawl: %+v\n", chat.Brawl)
			log.Printf("[debug] Chats object: %+v\n", Chats)
		}

		data = func() {
			log.Printf("[debug] Dict: %+v\n", Dict)
		}

		quit = func() {
			_, existsQ := chat.Queue[uID]
			if existsQ {
				delete(chat.Queue, uID)
				d := chat.MaxBrawlUserCount - len(chat.Queue)
				log.Printf("[bot] %v more fighter(s) required", d)
			}

			_, existsB := chat.Brawl[uID]
			if existsB {
				chat.Brawl[uID].Posted = false
				if len(chat.Brawl) < MIN_BRAWL_USERS {
					chat.DeleteListener(chat.BrawlChan)
				}
				delete(chat.Brawl, uID)
			}

		}

		fight = func() {

			if len(chat.Brawl) == 0 {

				_, existsQ := chat.Queue[uID]
				if !existsQ {
					d := chat.MaxBrawlUserCount - len(chat.Queue)
					chat.Queue.AddUser(user)
					if d <= 1 {
						Bot.Send(GetTxtMsg(chat.ID, fmt.Sprintf("All right! We've got ourselves a party, aren't we?")))
						time.Sleep(DELAY_TEXT)
						Bot.Send(GetTxtMsg(chat.ID, "_[EXCITED BEEP]_"))
						time.Sleep(DELAY_TEXT)
						log.Printf("[bot] Party is ready")
						if len(chat.Brawl) == 0 {
							n := 0
							for userInListID, userInList := range chat.Queue {
								n++
								chat.Brawl.AddUser(userInList)
								delete(chat.Queue, userInListID)
								if n >= chat.MaxBrawlUserCount {
									break
								}
							}
							chat.BrawlChan = chat.CreateListener()
							go BrawlActor(chat, chat.BrawlChan)
							chat.BrawlChan <- action
						} else {
							log.Printf("[bot] Wait for brawl to finish")
						}
					} else {
						if d-1 == 1 {
							Bot.Send(GetTxtMsg(chat.ID, fmt.Sprintf("We need another volunteer")))
						} else {
							Bot.Send(GetTxtMsg(chat.ID, fmt.Sprintf("We need another volunteer. %v, in fact", d-1)))
						}
						time.Sleep(DELAY_TEXT)
						log.Printf("[bot] %v more fighter(s) required", d-1)
					}
				}
			} else {
				log.Printf("Wait for brawl to finish")
			}
		}
	)

	var actionMap = map[string]func(){
		"start": start,
		"help":  help,
		"debug": debug,
		"data":  data,
		"quit":  quit,
		"fight": fight,
	}

	return actionMap

}
