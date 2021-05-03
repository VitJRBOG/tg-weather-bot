package bot

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/VitJRBOG/TelegramWeatherBot/pogoda_api"
	"github.com/VitJRBOG/TelegramWeatherBot/tg_api"
	"github.com/VitJRBOG/TelegramWeatherBot/tools"
)

func Start(cfg tools.Config) {
	checkingChat(cfg)
}

func checkingChat(cfg tools.Config) {
	for {
		values := url.Values{
			"offset":  {strconv.Itoa(cfg.UpdatesOffset)},
			"timeout": {strconv.Itoa(cfg.Timeout)},
		}

		updates := tg_api.GetUpdates(cfg.AccessToken, "getUpdates", values)

		if len(updates) == 0 {
			continue
		}

		for _, updateData := range updates {
			commandRecognized := checkUserCommand(updateData.Message.Text)
			if commandRecognized {
				forecast := pogoda_api.GetForecast(cfg.PogodaApiURL, "1", updateData.Message.Text)
				forecastAvailable := checkWeatherForecast(forecast)
				if forecastAvailable {
					sendWeatherForecast(cfg, forecast, updateData.Message.Chat.ID)
				} else {
					sendMessageAboutForecastUnavailable(cfg, updateData.Message)
				}
			} else {
				sendHint(cfg, updateData.Message)
			}
		}

		cfg.UpdateUpdatesOffset(updates[len(updates)-1].UpdateID + 1)
	}
}

func checkUserCommand(message string) bool {
	matched, err := regexp.MatchString("[0-9]{4}-[0-9]{2}-[0-9]{2}", message)
	if err != nil {
		panic(err.Error())
	}

	return matched
}

func sendHint(cfg tools.Config, messageData tg_api.Message) {
	m := fmt.Sprintf("Команда не распознана. Для получения прогноза погоды отправьте дату "+
		"нужного прогноза в формате ГГГГ-ММ-ДД.\nНапример (без кавычек): «%s».",
		unixTimestampToHumanReadableFormat(time.Now().Unix()))
	values := url.Values{
		"chat_id": {strconv.Itoa(messageData.Chat.ID)},
		"text":    {m},
	}
	tg_api.SendMessage(cfg.AccessToken, values)
}

func checkWeatherForecast(forecast pogoda_api.Forecast) bool {
	return forecast.OrenburgOblast.Date != ""
}

func sendWeatherForecast(cfg tools.Config, forecast pogoda_api.Forecast, chatID int) {
	f := fmt.Sprintf("Оренбургская область\nПрогноз погоды на: %s.\n\n"+
		"Ночью %s, %s. "+
		"Температура ночью %s. Ветер %s, %s м/с.\n\n"+
		"Днем %s, %s. "+
		"Температура днем %s. Ветер %s, %s м/с.",
		forecast.OrenburgOblast.Date,
		forecast.OrenburgOblast.NightCloud, forecast.OrenburgOblast.NightPrec,
		forecast.OrenburgOblast.NightTemp, forecast.OrenburgOblast.NightWindDirrect,
		forecast.OrenburgOblast.NightWindSpeed,
		forecast.OrenburgOblast.DayCloud, forecast.OrenburgOblast.DayPrec,
		forecast.OrenburgOblast.DayTemp, forecast.OrenburgOblast.DayWindDirrect,
		forecast.OrenburgOblast.DayWindSpeed)
	values := url.Values{
		"chat_id": {strconv.Itoa(chatID)},
		"text":    {f},
	}
	tg_api.SendMessage(cfg.AccessToken, values)
}

func unixTimestampToHumanReadableFormat(ut int64) string {
	t := time.Unix(ut, 0)
	dateFormat := "2006-01-02"
	date := t.Format(dateFormat)
	return date
}

func sendMessageAboutForecastUnavailable(cfg tools.Config, messageData tg_api.Message) {
	m := "Не удалось получить прогноз на указанную дату."
	values := url.Values{
		"chat_id": {strconv.Itoa(messageData.Chat.ID)},
		"text":    {m},
	}
	tg_api.SendMessage(cfg.AccessToken, values)
}
