package bot

import (
	"database/sql"
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
	checkingChats(cfg)
}

func checkingChats(cfg tools.Config) {
	dbase := db.Connect(cfg.DBConnection)
	handlers := make(map[int]chan tg_api.Message)

	for {
		values := url.Values{
			"offset":  {strconv.Itoa(cfg.UpdatesOffset)},
			"timeout": {strconv.Itoa(cfg.Timeout)},
		}

		updates := tg_api.GetUpdates(cfg.AccessToken, "getUpdates", values)

		if len(updates) == 0 {
			continue
		}

		for i := 0; i < len(updates); i++ {
			if channel, exist := handlers[updates[i].Message.From.ID]; exist {
				select {
				case channel <- updates[i].Message:
					continue
				default:
					delete(handlers, updates[i].Message.From.ID)
				}
			}
			handlers[updates[i].Message.From.ID] = make(chan tg_api.Message)
			go handlingRequest(dbase, handlers[updates[i].Message.From.ID], cfg)
			handlers[updates[i].Message.From.ID] <- updates[i].Message
		}

		cfg.UpdateUpdatesOffset(updates[len(updates)-1].UpdateID + 1)
	}
}

func handlingRequest(dbase *sql.DB, channel chan tg_api.Message, cfg tools.Config) {
	district := 0
	for {
		messageData := <-channel

		var user db.User

		user.UserID = messageData.From.ID

		if users := user.SelectByUserID(dbase); len(users) == 0 {
			user.Name = fmt.Sprintf("%s %s",
				messageData.From.FirstName, messageData.From.LastName)
			user.Username = messageData.From.Username
			user.RequestCount = 1
			user.InsertInto(dbase) // пропущена обработка возвращаемых значений
		} else {
			users[0].RequestCount++
			users[0].Update(dbase) // пропущена обработка возвращаемых значений
		}

		if district == 0 {
			district = handlingDistrictSelection(cfg.AccessToken, messageData)
		} else {
			ok := handlingDateSelection(cfg, district, messageData)
			if ok {
				channel = nil
				break
			}
		}
	}
}

func handlingDistrictSelection(accessToken string, messageData tg_api.Message) int {
	m := ""
	district := 0
	if district, m = checkDistrict(messageData.Text); district > 0 {
		m = fmt.Sprintf("%s Для получения прогноза погоды отправьте дату "+
			"нужного прогноза в формате ГГГГ-ММ-ДД.\nНапример (без кавычек): «%s».",
			m, unixTimestampToHumanReadableFormat(time.Now().Unix()))
		sendHint(accessToken, m, messageData.Chat.ID)
	} else {
		sendHint(accessToken, m, messageData.Chat.ID)
	}

	return district
}

func handlingDateSelection(cfg tools.Config, district int, messageData tg_api.Message) bool {
	ok := false
	if dateRecognized := checkDate(messageData.Text); dateRecognized {
		forecast := pogoda_api.GetForecast(cfg.PogodaApiURL, "1", messageData.Text)
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
			sendWeatherForecast(cfg, localForecast, messageData.Chat.ID)
		} else {
			sendMessageAboutForecastUnavailable(cfg, messageData)
		}
		ok = true
	} else {
		m := fmt.Sprintf("Дата не распознана. "+
			"Для получения прогноза погоды отправьте дату "+
			"нужного прогноза в формате ГГГГ-ММ-ДД.\nНапример (без кавычек): «%s».",
			unixTimestampToHumanReadableFormat(time.Now().Unix()))
		sendHint(cfg.AccessToken, m, messageData.Chat.ID)
	}

	return ok
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

func sendHint(accessToken, m string, chatId int) {
	values := url.Values{
		"chat_id": {strconv.Itoa(chatId)},
		"text":    {m},
	}
	tg_api.SendMessage(accessToken, values)
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
