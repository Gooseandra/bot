package main

import (
	"database/sql"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/go-yaml/yaml"
	_ "github.com/lib/pq"
	"log"
	"os"
	"strconv"
	"sync"
)

var BotAPI *tgbotapi.BotAPI

func main() {
	var mainMutex sync.Mutex
	var chats = map[int64]Chat{}
	var settings Settings
	bytes, fail := os.ReadFile(".yml")
	if fail != nil {
		log.Panic(fail.Error())
	}
	fail = yaml.Unmarshal([]byte(bytes), &settings)
	if fail != nil {
		log.Panic(fail.Error())
	}
	log.Println(settings)
	//fail = core.Db(settings.Database.Type, settings.Database.Arguments)
	//if fail != nil {
	//	log.Panic(fail.Error())
	//}
	BotAPI, fail = tgbotapi.NewBotAPI(settings.Telegram)
	if fail != nil {
		log.Panic(fail)
	}
	//rand.Seed(time.Now().UnixNano())
	update := tgbotapi.NewUpdate(0)
	update.Timeout = 9
	channel := BotAPI.GetUpdatesChan(update)

	db, err := sql.Open(settings.Database.Type, settings.Database.Arguments)
	if err != nil {
		log.Println(err.Error())
		return
	}
	for {
		message := <-channel
		mainMutex.Lock()
		chat, found := chats[message.FromChat().ID]
		if !found {
			chat = Chat{id: message.FromChat().ID, channel: make(chan tgbotapi.Update)}
			chats[message.FromChat().ID] = chat
			log.Println("Chat " + strconv.FormatInt(message.FromChat().ID, 10) + " created")
			go chat.routine(chats, &mainMutex, db)
		}
		mainMutex.Unlock()
		chat.channel <- message
	}
}
