package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"gopkg.in/telegram-bot-api.v4"
)

func Shuffle(inS []string) []string {
	s := make([]string, len(inS))
	copy(s, inS)
	for i := range s {
		j := rand.Intn(i + 1)
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func GetTxtMsg(cID int64, text string) tgbotapi.MessageConfig {
	var msg tgbotapi.MessageConfig
	if text != "" {
		msg = tgbotapi.NewMessage(cID, text)
		msg.ParseMode = tgbotapi.ModeMarkdown
	}
	return msg
}

func GetVoteMsg(msg tgbotapi.MessageConfig, mID int) tgbotapi.MessageConfig {
	kb := GetUpdVoteMarkup(make(map[string]int))
	msg.ReplyMarkup = kb
	msg.ReplyToMessageID = mID
	return msg
}

func GetUpdVoteMarkup(voteResults map[string]int) tgbotapi.InlineKeyboardMarkup {
	var kbRow []tgbotapi.InlineKeyboardButton

	for _, key := range VoteOrder {
		kbRow = append(
			kbRow,
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf(
					"%v %v",
					VoteMap[key].Emoji,
					voteResults[key],
				),
				key,
			))
	}

	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(kbRow...))
	return kb
}

func GetVoteResults(votee *User, voteStats *VoteStats) map[string]int {
	voteResults := make(map[string]int)
	for pairInList, voteInList := range *voteStats {
		if pairInList.Votee.ID == votee.ID {
			for key := range VoteMap {
				voteResults[key] = voteResults[key] + (*voteInList)[key]
			}
		}
	}
	return voteResults
}

func ConfIDenceRating(score, maxScore int) float64 {

	if maxScore == 0 {
		return 0
	}

	n := float64(maxScore + 1)
	p := float64(score+1) / n
	z := 1.281551565545

	left := p + 1/(2*n)*z*z
	right := z * math.Sqrt(p*(1-p)/n+z*z/(4*n*n))
	under := 1 + 1/n*z*z

	return (left - right) / under
}

func BrawlActor(chat *Chat, c chan *Action) {
	defer log.Printf("[server] Exiting brawlActor goroutine for chat %v", chat.ID)
	log.Printf("[server] Starting brawlActor goroutine for chat %v", chat.ID)

	var (
		text  string
		story *Story
	)

	storiesWeighted := []*Story{}
	for _, s := range Dict.Stories {
		for i := 0; i < s.Freq; i++ {
			storiesWeighted = append(storiesWeighted, s)
		}
	}

	for {
		story = storiesWeighted[rand.Intn(len(storiesWeighted))]
		text = story.Text[rand.Intn(len(story.Text))]
		inList := false
		for _, textInList := range chat.RecentTexts {
			if textInList == text {
				inList = true
			}
		}
		if inList == false {
			break
		}
	}

	chat.RecentTexts = append(chat.RecentTexts, text)
	chat.RecentTexts = chat.RecentTexts[MaxInt(0, len(chat.RecentTexts)-RECENT_TEXT_MEMORY):]

	wlcmTxt := "A new challenge for:\n"
	n := 0
	bullets := Shuffle(BulletsEmoji)
	for _, u := range chat.Brawl {
		e := bullets[n]
		n++
		wlcmTxt += fmt.Sprintf("%s %s\n", e, u.SprintName())
	}
	Bot.Send(GetTxtMsg(chat.ID, wlcmTxt))

	time.Sleep(DELAY_TEXT)
	Bot.Send(GetTxtMsg(chat.ID, fmt.Sprintf("My face when *%s*", text)))
	time.Sleep(DELAY_TEXT)
	Bot.Send(GetTxtMsg(chat.ID, fmt.Sprintf("Feel free to post your photos now")))

	timer1 := time.NewTimer(DELAY_PHOTO)
	timer2 := time.NewTimer(100)
	timer2.Stop()

	voteStats := VoteStats{}

	brk := false
	for !brk {
		select {
		case action, ok := <-c:
			if ok {
				if action.Type == "photo" {
					for _, user := range chat.Brawl {

						if user.ID == action.From.ID && !user.Posted {

							user.Posted = true
							Bot.Send(GetVoteMsg(
								GetTxtMsg(chat.ID, fmt.Sprintf("Please vote")),
								action.Message.ID),
							)

							timer1.Reset(DELAY_VOTES)
							timer2.Stop()
						}
					}
				}

				if action.Type == "vote" {

					voter := action.From
					votee := action.Message.ReplyToMsg.From
					var pair = VotePair{voter, votee}

					vote, exists := voteStats[pair]
					if !exists {
						vote = &Vote{}
						voteStats[pair] = vote
					}

					voteStats[pair] = &Vote{action.Callback.Data: 1}
					clbTxt := fmt.Sprintf("You %s-ed %s", VoteMap[action.Callback.Data].Emoji, votee.SprintName())

					voteResults := GetVoteResults(votee, &voteStats)

					editMsgCfg := tgbotapi.NewEditMessageReplyMarkup(
						chat.ID,
						action.Message.ID,
						GetUpdVoteMarkup(voteResults))
					Bot.Send(editMsgCfg)

					clbCfg := tgbotapi.CallbackConfig{
						CallbackQueryID: action.Callback.ID,
						Text:            clbTxt,
					}
					Bot.AnswerCallbackQuery(clbCfg)
				}
			} else {
				return
			}
		case <-timer1.C:
			log.Printf("[bot] First brawl timeout")
			Bot.Send(GetTxtMsg(chat.ID, fmt.Sprintf("Hurry up! You have %v seconds left", DELAY_LAST_VOTES_SEC)))
			timer2.Reset(DELAY_LAST_VOTES)
		case <-timer2.C:
			log.Printf("[bot] Final brawl timeout")
			brk = true
		}
	}

	Bot.Send(GetTxtMsg(chat.ID, "Time is up!"))
	time.Sleep(DELAY_TEXT)

	voteSummary := make(map[*User]*VoteResult)
	for pairInList, voteInList := range voteStats {
		for key, n := range *voteInList {
			if voteSummary[pairInList.Votee] == nil {
				voteSummary[pairInList.Votee] = &VoteResult{}
			}
			voteSummary[pairInList.Votee].N += n
			voteSummary[pairInList.Votee].Score += VoteMap[key].Effect * n
		}
	}

	var winner struct {
		user       *User
		score      int
		n          int
		confRating float64
	}
	winner.confRating = -1
	for user, resultInList := range voteSummary {
		confRating := ConfIDenceRating(resultInList.Score, resultInList.N)
		if confRating > winner.confRating {
			winner.user = user
			winner.score = resultInList.Score
			winner.n = resultInList.N
			winner.confRating = confRating
		}
	}

	if winner.user == nil {
		Bot.Send(GetTxtMsg(chat.ID, "OK, so nobody chose to participate this time. Oh well..."))
		time.Sleep(DELAY_TEXT)
		Bot.Send(GetTxtMsg(chat.ID, "_[SAD BEEP]_"))
	} else {
		if winner.score < winner.n-winner.score {
			if len(chat.Brawl) == 2 {
				Bot.Send(GetTxtMsg(chat.ID, "Well... This was a tough one. Both of you guys were quite shitty"))
			} else {
				Bot.Send(GetTxtMsg(chat.ID, "Well... This was a tough one. All of you guys were quite shitty"))
			}
			time.Sleep(DELAY_TEXT)
			Bot.Send(GetTxtMsg(chat.ID, "Please welcome a winner:"))
			time.Sleep(DELAY_TEXT)
			Bot.Send(GetTxtMsg(chat.ID, fmt.Sprintf("%v", winner.user.SprintName())))
			time.Sleep(DELAY_TEXT)
			Bot.Send(GetTxtMsg(chat.ID, "_[SARCASATIC BEEP]_"))
		} else {
			time.Sleep(DELAY_TEXT)
			Bot.Send(GetTxtMsg(chat.ID, "All right! That was a fair competition"))
			time.Sleep(DELAY_TEXT)
			Bot.Send(GetTxtMsg(chat.ID, "I am pleased to announce the winner. Please welcome:"))
			time.Sleep(DELAY_TEXT)
			Bot.Send(GetTxtMsg(chat.ID, fmt.Sprintf("%v", winner.user.SprintName())))
			time.Sleep(DELAY_TEXT)
			Bot.Send(GetTxtMsg(chat.ID, "_[TRIUMPHANT BEEP]_"))
		}

		log.Printf("[bot] Winner: %+v", winner)
	}

	for _, u := range chat.Brawl {
		u.Posted = false
	}
	chat.Brawl = UserList{}
	chat.DeleteListener(c)
	return
}
