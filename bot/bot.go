package bot

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/VitJRBOG/TelegramWeatherBot/pogoda_api"
	"github.com/VitJRBOG/TelegramWeatherBot/tg_api"
	"github.com/VitJRBOG/TelegramWeatherBot/tools"
)

func Start(cfg tools.Config) {
	checkingChat(cfg)
}

func checkingChat(cfg tools.Config) {
	values := url.Values{
		"offset":  {strconv.Itoa(cfg.UpdatesOffset)},
		"timeout": {"5"},
	}

	result := tg_api.GetUpdates(cfg.AccessToken, "getUpdates", values)

	if len(result.Updates) > 0 {
		sendWeatherForecast(cfg, result.Updates[0].Message.Chat)
		cfg.UpdateUpdatesOffset(result.Updates[len(result.Updates)-1].UpdateID)
	}
}

func sendWeatherForecast(cfg tools.Config, chat tg_api.Chat) {
	forecast := pogoda_api.GetForecast(cfg.PogodaApiURL, "1", "2021-5-4")
	f := fmt.Sprintf("Ожидается %s, %s. Температура %s. Ветер %s, %s м/с.",
		forecast.OrenburgOblast.DayCloud, forecast.OrenburgOblast.DayPrec,
		forecast.OrenburgOblast.DayTemp, forecast.OrenburgOblast.DayWindDirrect,
		forecast.OrenburgOblast.DayWindSpeed)
	values := url.Values{
		"chat_id": {strconv.Itoa(chat.ID)},
		"text":    {f},
	}
	tg_api.SendMessage(cfg.AccessToken, values)
}
