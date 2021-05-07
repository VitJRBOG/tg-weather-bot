package bot

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/VitJRBOG/TelegramWeatherBot/db"
	"github.com/VitJRBOG/TelegramWeatherBot/pogoda_api"
	"github.com/VitJRBOG/TelegramWeatherBot/tg_api"
	"github.com/VitJRBOG/TelegramWeatherBot/tools"
)

func Start(cfg tools.Config) {
	checkingChat(cfg)
}

func checkingChat(cfg tools.Config) {
	dbase := db.Connect(cfg.DBConnection)

	waitingForDate := false
	district := 0
	m := ""

	for {
		values := url.Values{
			"offset":  {strconv.Itoa(cfg.UpdatesOffset)},
			"timeout": {strconv.Itoa(cfg.Timeout)},
		}

		updates := tg_api.GetUpdates(cfg.AccessToken, "getUpdates", values)

		if len(updates) == 0 {
			continue
		}

		for _, update := range updates {
			var user db.User

			user.UserID = update.Message.From.ID

			if users := user.SelectByUserID(dbase); len(users) == 0 {
				user.Name = fmt.Sprintf("%s %s",
					update.Message.From.FirstName, update.Message.From.LastName)
				user.Username = update.Message.From.Username
				user.RequestCount = 1
				user.InsertInto(dbase) // пропущена обработка возвращаемых значений
			} else {
				users[0].RequestCount++
				users[0].Update(dbase) // пропущена обработка возвращаемых значений
			}
		}

		if len(updates) > 1 {
			m = ""
			cfg.UpdateUpdatesOffset(updates[len(updates)-1].UpdateID + 1)
			continue
		}

		if waitingForDate {
			if dateRecognized := checkDate(updates[0].Message.Text); dateRecognized {
				forecast := pogoda_api.GetForecast(cfg.PogodaApiURL, "1", updates[0].Message.Text)
				var localForecast pogoda_api.Weather
				switch district {
				case 182:
					localForecast = forecast.OrenburgOblast
				case 111:
					localForecast = forecast.Orenburg
				case 106:
					localForecast = forecast.Buzuluk
				case 112:
					localForecast = forecast.Orsk
				}
				forecastAvailable := checkWeatherForecast(localForecast)
				if forecastAvailable {
					sendWeatherForecast(cfg, localForecast, updates[0].Message.Chat.ID)
				} else {
					sendMessageAboutForecastUnavailable(cfg, updates[0].Message)
				}
				waitingForDate = false
			} else {
				m = fmt.Sprintf("Дата не распознана. "+
					"Для получения прогноза погоды отправьте дату "+
					"нужного прогноза в формате ГГГГ-ММ-ДД.\nНапример (без кавычек): «%s».",
					unixTimestampToHumanReadableFormat(time.Now().Unix()))
				sendHint(cfg, m, updates[0].Message.Chat.ID)
			}
		} else {
			if district, m = checkDistrict(updates[0].Message.Text); district > 0 {
				m = fmt.Sprintf("%s Для получения прогноза погоды отправьте дату "+
					"нужного прогноза в формате ГГГГ-ММ-ДД.\nНапример (без кавычек): «%s».",
					m, unixTimestampToHumanReadableFormat(time.Now().Unix()))
				sendHint(cfg, m, updates[0].Message.Chat.ID)
				waitingForDate = true
			} else {
				sendHint(cfg, m, updates[0].Message.Chat.ID)
			}
		}

		m = ""
		cfg.UpdateUpdatesOffset(updates[len(updates)-1].UpdateID + 1)
	}
}

func checkDistrict(message string) (int, string) {
	switch true {
	case message == "/orenburg_oblast":
		return 182, "Запрос прогноза погоды по Оренбургской области."
	case message == "/orenburg":
		return 111, "Запрос прогноза погоды по Оренбургу."
	case message == "/buzuluk":
		return 106, "Запрос прогноза погоды по Бузулуку."
	case message == "/orsk":
		return 112, "Запрос прогноза погоды по Орску."
	}
	return 0, "Команда не распознана. " +
		"Для выбора региона/города введите «/» («слэш», без кавычек) " +
		"и выберите вариант из списка."
}

func checkDate(message string) bool {
	matched, err := regexp.MatchString("[0-9]{4}-[0-9]{2}-[0-9]{2}", message)
	if err != nil {
		panic(err.Error())
	}

	return matched
}

func checkWeatherForecast(localForecast pogoda_api.Weather) bool {
	return localForecast.Date != ""
}

func sendHint(cfg tools.Config, m string, chatId int) {
	values := url.Values{
		"chat_id": {strconv.Itoa(chatId)},
		"text":    {m},
	}
	tg_api.SendMessage(cfg.AccessToken, values)
}

func sendWeatherForecast(cfg tools.Config, localForecast pogoda_api.Weather, chatID int) {
	f := ""
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
