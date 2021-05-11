package main

import (
	"log"
	"runtime/debug"

	"github.com/VitJRBOG/TelegramWeatherBot/internal/bot"
	"github.com/VitJRBOG/TelegramWeatherBot/internal/tools"
)

func main() {
	cfg, err := tools.GetConfig()
	if err != nil {
		log.Panicf("%s\n%s\n", err, debug.Stack())
	}
	bot.Start(cfg)
}