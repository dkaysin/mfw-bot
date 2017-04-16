package main

import (
	"fmt"
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"time"
)

const (
	CHAT_TIMEOUT = 100 * time.Second
	RECENT_TEXT_MEMORY   = 3
	MAX_BRAWL_USERS = 4
	MIN_BRAWL_USERS = 2
	COEFF_BRAWL_USERS = 0.2
)

func chatActor(chat *Chat, c chan *Action) {

	defer log.Printf("[server] Exiting chatActor goroutine for chat %v", chat.Id)
	defer chat.Delete()

	log.Printf("[server] Starting chatActor goroutine for chat %v", chat.Id)

	timer := time.NewTimer(CHAT_TIMEOUT)

	for {
		select {
		case action := <-c:

			usersCount, _ := Bot.GetChatMembersCount(tgbotapi.ChatConfig{chat.Id,""})
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

	var user *User
	if action.Msg != nil {
		user = chat.GetUser(action.Msg.From)
	} else if action.Clb != nil {
		user = chat.GetUser(action.Clb.From)
	} else {
		return nil
	}
	uid := user.Id

	var (
		start = func() {
			Bot.Send(GetTxtMsg(chat.Id, fmt.Sprintf("Hello, I am 'My Face When...'-bot.")))
			time.Sleep(DELAY_TEXT)
			Bot.Send(GetTxtMsg(chat.Id, fmt.Sprintf("I'm dying to play a game with you")))
			time.Sleep(DELAY_TEXT)
			Bot.Send(GetTxtMsg(chat.Id, "_[AGITATED BEEP]_"))
			time.Sleep(DELAY_TEXT)

			msg := GetTxtMsg(chat.Id, "Just type /fight or choose one of the options below:")

			kb := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						"Enter the brawl",
						"fight"),
					tgbotapi.NewInlineKeyboardButtonData(
						"Help",
						"help"),
				))
			msg.ReplyMarkup = kb
			Bot.Send(msg)
		}

		help = func() {
			Bot.Send(GetTxtMsg(chat.Id, fmt.Sprintf("Heeelp! I need somebooody...")))
			time.Sleep(DELAY_TEXT)
			Bot.Send(GetTxtMsg(chat.Id, fmt.Sprintf("Heeelp! Not just anybooody...")))
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
			_, exists := chat.Queue[uid]
			if exists {
				delete(chat.Queue, uid)
				d := chat.MaxBrawlUserCount - len(chat.Queue)
				log.Printf("[bot] %v more fighter(s) required", d)
			}
		}

		fight = func() {

			if len(chat.Brawl) == 0 {

				_, exists := chat.Queue[uid]
				if !exists {
					d := chat.MaxBrawlUserCount - len(chat.Queue)
					chat.Queue.AddUser(user)
					if d <= 1 {
						Bot.Send(GetTxtMsg(chat.Id, fmt.Sprintf("All right! We've got ourselves a party, aren't we?")))
						time.Sleep(DELAY_TEXT)
						Bot.Send(GetTxtMsg(chat.Id, "_[EXCITED BEEP]_"))
						time.Sleep(DELAY_TEXT)
						log.Printf("[bot] Party is ready")
						if len(chat.Brawl) == 0 {
							n := 0
							for userInListId, userInList := range chat.Queue {
								n++
								chat.Brawl.AddUser(userInList)
								delete(chat.Queue, userInListId)
								if n >= chat.MaxBrawlUserCount {
									break
								}
							}
							c := chat.CreateListener()
							go brawlActor(chat, c)
							c <- action
						} else {
							log.Printf("[bot] Wait for brawl to finish")
						}
					} else {
						if d-1 == 1 {
							Bot.Send(GetTxtMsg(chat.Id, fmt.Sprintf("We need another volunteer")))
						} else {
							Bot.Send(GetTxtMsg(chat.Id, fmt.Sprintf("We need another volunteer. %v, in fact", d-1)))
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
		"help": help,
		"debug": debug,
		"data":  data,
		"quit":  quit,
		"fight": fight,
	}

	return actionMap

}
