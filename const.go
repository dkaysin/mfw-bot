package main

import (
	"os"
	"time"
)

const (
	MIN_PHOTO_SIZE   = 1000 // in pixels
	WEBHOOK_URL      = "https://mfw-bot.herokuapp.com"
	WEBHOOK_EXT_PORT = "443"

	CHAT_TIMEOUT       = 100 * time.Second
	RECENT_TEXT_MEMORY = 3
	MAX_BRAWL_USERS    = 4
	MIN_BRAWL_USERS    = 1
	COEFF_BRAWL_USERS  = 0.2

	DELAY_TEXT           = 1 * time.Second
	DELAY_PHOTO          = 60 * time.Second
	DELAY_VOTES_SEC      = 30 // in seconds
	DELAY_VOTES          = DELAY_VOTES_SEC * time.Second
	DELAY_LAST_VOTES_SEC = 15 // in seconds
	DELAY_LAST_VOTES     = DELAY_LAST_VOTES_SEC * time.Second
)

func APIKey() string {
	return os.Getenv("MFWBOT_API_KEY")
}

var (
	WEBHOOK_INT_PORT = os.Getenv("PORT")

	VoteMap = map[string]MapClbDataToVote{
		"a": MapClbDataToVote{
			Effect: 1,
			Emoji:  "\U0001F601",
		},
		"b": MapClbDataToVote{
			Effect: 1,
			Emoji:  "\U0001F525",
		},
		"c": MapClbDataToVote{
			Effect: 0,
			Emoji:  "\U0001F631",
		},
		"d": MapClbDataToVote{
			Effect: 0,
			Emoji:  "\U0001F4A9",
		},
	}

	VoteOrder = []string{"a", "b", "c", "d"}

	BulletsEmoji = []string{
		"\U0001F913",
		"\U0001F607",
		"\U0001F62C",
		"\U0001F634",
		"\U0001F644",
		"\U0001F60E",
		"\U0001F917",
	}
)
