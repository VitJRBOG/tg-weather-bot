package main

import (
	"encoding/json"
	"io/ioutil"

	tgbotapi "gopkg.in/telegram-bot-api.v5"
)

const PathToConfig = "config.json"

func main() {
	var cfg config
	cfg.parseJSON(PathToConfig)
	var botCtrl botControl
	botCtrl.cfg = cfg
	var err error
	botCtrl.bot, err = tgbotapi.NewBotAPI(cfg.AccessToken)
	if err != nil {
		panic(err.Error())
	}
	botCtrl.chatListener()
}

type config struct {
	AccessToken string `json:"access_token"`
	LastMsgDate int    `json:"last_message_date"`
}

func (c *config) parseJSON(pathToFile string) {
	content, err := ioutil.ReadFile(pathToFile)
	if err != nil {
		panic(err.Error())
	}
	err = json.Unmarshal(content, c)
	if err != nil {
		panic(err.Error())
	}
}

func (c *config) updateLastMsgDate(newDate int) {
	c.LastMsgDate = newDate

	content, err := json.Marshal(c)
	if err != nil {
		panic(err.Error())
	}
	ioutil.WriteFile(PathToConfig, content, 0644)
}

type botControl struct {
	bot *tgbotapi.BotAPI
	cfg config
}

func (b *botControl) chatListener() {
	upd := tgbotapi.NewUpdate(0)

	updates, err := b.bot.GetUpdatesChan(upd)
	if err != nil {
		panic(err.Error())
	}
	for {
		var timeToFinish bool
		select {
		case update := <-updates:
			if update.Message.Date > b.cfg.LastMsgDate {
				timeToFinish = b.sendMessage(&update)
			}
		}
		if timeToFinish {
			break
		}
	}
}

func (b *botControl) sendMessage(u *tgbotapi.Update) bool {
	var timeToFinish bool
	var botMsgCfg tgbotapi.MessageConfig
	switch u.Message.Text {
	case "/start":
		botMsgCfg = tgbotapi.NewMessage(u.Message.Chat.ID, "I'm here!")
		timeToFinish = false
	case "a":
		botMsgCfg = tgbotapi.NewMessage(u.Message.Chat.ID, "b")
		timeToFinish = false
	case "b":
		botMsgCfg = tgbotapi.NewMessage(u.Message.Chat.ID, "c")
		timeToFinish = false
	case "c":
		botMsgCfg = tgbotapi.NewMessage(u.Message.Chat.ID, "d")
		timeToFinish = false
	case "Bye":
		botMsgCfg = tgbotapi.NewMessage(u.Message.Chat.ID, "I should go!")
		timeToFinish = true
	default:
		botMsgCfg = tgbotapi.NewMessage(u.Message.Chat.ID, "Nevermind...")
		timeToFinish = false
	}

	b.bot.Send(botMsgCfg)
	b.cfg.updateLastMsgDate(u.Message.Date)
	return timeToFinish
}
