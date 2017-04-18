package main

import (
	"fmt"
	"log"
	"sync"

	"gopkg.in/telegram-bot-api.v4"
)

func (ul *UserList) AddUser(u *User) {
	(*ul)[u.ID] = u
	return
}

func (chat *Chat) CreateListener() chan *Action {
	chat.Lock()
	defer chat.Unlock()

	c := make(chan *Action)
	chat.Listeners = append(chat.Listeners, c)
	return c
}

func (chat *Chat) DeleteListener(c chan *Action) {
	chat.Lock()
	defer chat.Unlock()

	close(c)
	a := chat.Listeners
	for i, cR := range a {
		if c == cR {
			a[i] = a[len(a)-1]
			a = a[:len(a)-1]
			chat.Listeners = a
			return
		}
	}
}

func (chat *Chat) SendToListeners(a *Action) {
	chat.Lock()
	defer chat.Unlock()

	for _, c := range chat.Listeners {
		c <- a
	}
	log.Printf("[server] Sent to %v listeners: %+v", len(chat.Listeners), a)
}

func (chat *Chat) Delete() {
	chat.Lock()
	defer chat.Unlock()

	for _, c := range chat.Listeners {
		close(c)
	}
	delete(Chats, chat.ID)
	return
}

func (chat *Chat) GetUser(userRaw *tgbotapi.User) *User {
	chat.Lock()
	defer chat.Unlock()

	uid := userRaw.ID

	user, exists := chat.Users[uid]
	if !exists {
		user = &User{
			ID:        uid,
			Firstname: userRaw.FirstName,
			Lastname:  userRaw.LastName,
			Username:  userRaw.UserName,
		}
		chat.Users[uid] = user
	}
	return user
}

func (chat *Chat) LockAndExec(f func()) func() {
	return func() {
		chat.Lock()
		defer chat.Unlock()
		f()
	}
}

func (user *User) SprintName() string {
	if user != nil {
		return fmt.Sprintf("%v %v", user.Firstname, user.Lastname)
	} else {
		return ""
	}
}

type VoteStats map[VotePair]*Vote

type MapClbDataToVote struct {
	Effect int
	Emoji  string
}

type VoteResult struct {
	Score int
	N     int
}

type Vote map[string]int

type VotePair struct {
	Voter *User
	Votee *User
}

type UserList map[int]*User

type User struct {
	ID        int
	Firstname string
	Lastname  string
	Username  string
	Posted    bool
}

type Action struct {
	Type     string
	From     *User
	Message  *Message
	Callback *Callback
}

type Message struct {
	ID         int
	From       *User
	ReplyToMsg *Message
	Text       string
}

type Callback struct {
	ID   string
	Data string
}

type Chat struct {
	sync.Mutex
	ID                int64
	Users             UserList
	Queue             UserList
	Brawl             UserList
	Listeners         []chan *Action
	RecentTexts       []string
	MaxBrawlUserCount int
}

type Story struct {
	Text []string
	Freq int
}

type Data struct {
	Stories []*Story
}
