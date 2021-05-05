package bot

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
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

		answering(cfg, updates)

		cfg.UpdateUpdatesOffset(updates[len(updates)-1].UpdateID + 1)
	}
}

func answering(cfg tools.Config, updates []tg_api.Update) {
	for _, updateData := range updates {
		commandRecognized := checkUserCommand(updateData.Message.Text)
		if commandRecognized {
			forecast := pogoda_api.GetForecast(cfg.PogodaApiURL, "1", updateData.Message.Text)
			forecastAvailable := checkWeatherForecast(forecast)
			if forecastAvailable {
				sendWeatherForecast(cfg, forecast.OrenburgOblast, updateData.Message.Chat.ID)
			} else {
				sendMessageAboutForecastUnavailable(cfg, updateData.Message)
			}
		} else {
			sendHint(cfg, updateData.Message)
		}
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

func sendWeatherForecast(cfg tools.Config, localForecast pogoda_api.Weather, chatID int) {
	f := fmt.Sprintf("Оренбургская область\nПрогноз погоды на: %s.\n\n", localForecast.Date)

	f = makeNightForecastMessage(f, localForecast)
	f = makeDayForecastMessage(f, localForecast)

	values := url.Values{
		"chat_id": {strconv.Itoa(chatID)},
		"text":    {f},
	}
	tg_api.SendMessage(cfg.AccessToken, values)
}

func makeNightForecastMessage(f string, localForecast pogoda_api.Weather) string {
	if localForecast.NightCloud != "" && localForecast.NightPrec != "" {
		f += fmt.Sprintf("Ночью %s, %s.", localForecast.NightCloud, localForecast.NightPrec)
	} else {
		switch true {
		case localForecast.NightCloud != "":
			f += fmt.Sprintf("Ночью %s.", localForecast.NightCloud)
		case localForecast.NightPrec != "":
			f += fmt.Sprintf("Ночью %s.", localForecast.NightPrec)
		}
	}

	if localForecast.NightPrecComm != "" {
		f += fmt.Sprintf(" %s.\n", localForecast.NightPrecComm)
	} else {
		f += "\n"
	}

	if localForecast.NightTemp != "" {
		f += fmt.Sprintf("Температура ночью %s°C",
			strings.ReplaceAll(localForecast.NightTemp, ",", "..."))
	}

	if localForecast.NightTempComm != "" {
		f += fmt.Sprintf(", %sC.\n",
			strings.ToLower(strings.ReplaceAll(localForecast.NightTempComm, ",", "...")))
	} else {
		f += ".\n"
	}

	if localForecast.NightWindDirrect != "" && localForecast.NightWindSpeed != "" {
		f += fmt.Sprintf("Ветер %s, %s м/с. ",
			localForecast.NightWindDirrect, localForecast.NightWindSpeed)
	}

	if localForecast.NightWindComm != "" {
		f += fmt.Sprintf("%s.\n", localForecast.NightWindComm)
	} else {
		f += "\n"
	}

	return f
}

func makeDayForecastMessage(f string, localForecast pogoda_api.Weather) string {
	if localForecast.DayCloud != "" && localForecast.DayPrec != "" {
		f += fmt.Sprintf("\nДнем %s, %s.", localForecast.DayCloud, localForecast.DayPrec)
	} else {
		switch true {
		case localForecast.DayCloud != "":
			f += fmt.Sprintf("Днем %s.", localForecast.DayCloud)
		case localForecast.DayPrec != "":
			f += fmt.Sprintf("Днем %s.", localForecast.DayPrec)
		}
	}

	if localForecast.DayPrecComm != "" {
		f += fmt.Sprintf(" %s.\n", localForecast.DayPrecComm)
	} else {
		f += "\n"
	}

	if localForecast.DayTemp != "" {
		f += fmt.Sprintf("Температура днем %s°C",
			strings.ReplaceAll(localForecast.DayTemp, ",", "..."))
	}

	if localForecast.DayTempComm != "" {
		f += fmt.Sprintf(", %sC.\n",
			strings.ToLower(strings.ReplaceAll(localForecast.DayTempComm, ",", "...")))
	} else {
		f += ".\n"
	}

	if localForecast.DayWindDirrect != "" && localForecast.DayWindSpeed != "" {
		f += fmt.Sprintf("Ветер %s, %s м/с. ",
			localForecast.DayWindDirrect, localForecast.DayWindSpeed)
	}

	if localForecast.DayWindComm != "" {
		f += fmt.Sprintf("%s.", localForecast.DayWindComm)
	}

	return f
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
