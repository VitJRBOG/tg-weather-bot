package main

import (
	"github.com/VitJRBOG/TelegramWeatherBot/bot"
	"github.com/VitJRBOG/TelegramWeatherBot/tools"
)

func main() {
	cfg := tools.GetConfig()
	bot.Start(cfg)
}
