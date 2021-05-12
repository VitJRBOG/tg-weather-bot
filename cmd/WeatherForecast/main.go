package main

import (
	"log"
	"runtime/debug"

	"github.com/VitJRBOG/TelegramWeatherBot/internal/bot"
	"github.com/VitJRBOG/TelegramWeatherBot/internal/tools"
)

func main() {
	botConn, err := tools.GetBotConnectionData()
	if err != nil {
		log.Panicf("%s\n%s\n", err, debug.Stack())
	}
	pogodaApiConn, err := tools.GetPogodaAPIConnectionData()
	if err != nil {
		log.Panicf("%s\n%s\n", err, debug.Stack())
	}
	dbConn, err := tools.GetDBConnectionData()
	if err != nil {
		log.Printf("%s\n%s\n", err, debug.Stack())
	}
	bot.Start(botConn, pogodaApiConn, dbConn)
}
