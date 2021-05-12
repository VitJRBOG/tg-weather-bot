package bot

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/VitJRBOG/TelegramWeatherBot/internal/db"
	"github.com/VitJRBOG/TelegramWeatherBot/internal/pogoda_api"
	"github.com/VitJRBOG/TelegramWeatherBot/internal/tg_api"
	"github.com/VitJRBOG/TelegramWeatherBot/internal/tools"
)

func Start(botConn tools.BotConn, pogodaApiConn tools.PogodaApiConn, dbConn tools.DBConn) {
	checkingChats(botConn, pogodaApiConn, dbConn)
}

func checkingChats(botConn tools.BotConn, pogodaApiConn tools.PogodaApiConn, dbConn tools.DBConn) {
	var dbase *sql.DB
	if (dbConn != tools.DBConn{}) {
		var err error
		dbase, err = db.Connect(dbConn)
		if err != nil {
			log.Printf("%s\n%s\n", err, debug.Stack())
		}
	}
	handlers := make(map[int]chan tg_api.Message)

	for {
		values := url.Values{
			"offset":  {strconv.Itoa(botConn.UpdatesOffset)},
			"timeout": {strconv.Itoa(botConn.Timeout)},
		}

		updates, err := tg_api.GetUpdates(botConn.AccessToken, "getUpdates", values)
		if err != nil {
			if strings.Contains(err.Error(), "error 401: Unauthorized") {
				log.Panicf("%s\n%s\n", err, debug.Stack())
			}
			log.Printf("%s\n%s\n", err, debug.Stack())
			continue
		}

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
			go handlingRequest(dbase, handlers[updates[i].Message.From.ID], botConn, pogodaApiConn)
			handlers[updates[i].Message.From.ID] <- updates[i].Message
		}

		err = updateOffset(&botConn, updates[len(updates)-1].UpdateID+1)
		if err != nil {
			log.Panicf("%s\n%s\n", err, debug.Stack())
		}
	}
}

func updateOffset(botConn *tools.BotConn, newOffset int) error {
	botConn.UpdatesOffset = newOffset
	err := botConn.UpdateFile()
	if err != nil {
		return err
	}
	return nil
}

func handlingRequest(dbase *sql.DB, channel chan tg_api.Message,
	botConn tools.BotConn, pogodaApiConn tools.PogodaApiConn) {
	district := 0
	for {
		messageData := <-channel

		if dbase != nil {
			go updateDB(dbase, messageData.From)
		}

		if district == 0 {
			var err error
			district, err = handlingDistrictSelection(botConn.AccessToken, messageData)
			if err != nil {
				log.Printf("%s\n%s\n", err, debug.Stack())
				m := "При выборе региона/города произошла ошибка."
				if err := sendMessageAboutError(m, botConn.AccessToken, messageData.From.ID); err != nil {
					log.Printf("%s\n%s\n", err, debug.Stack())
				}
				channel = nil
				return
			}
		} else {
			ok, err := handlingDateSelection(botConn, pogodaApiConn, district, messageData)
			if err != nil {
				log.Printf("%s\n%s\n", err, debug.Stack())
				m := "При получении прогноза произошла ошибка."
				if err := sendMessageAboutError(m, botConn.AccessToken, messageData.From.ID); err != nil {
					log.Printf("%s\n%s\n", err, debug.Stack())
				}
				channel = nil
				return
			}
			if ok {
				channel = nil
				break
			}
		}
	}
}

func updateDB(dbase *sql.DB, sender tg_api.User) {
	var user db.User

	user.UserID = sender.ID

	users, err := user.SelectFrom(dbase)
	if err != nil {
		log.Printf("%s\n%s\n", err, debug.Stack())
		return
	}

	if len(users) == 0 {
		user.Name = fmt.Sprintf("%s %s",
			sender.FirstName, sender.LastName)
		user.Username = sender.Username
		user.RequestCount = 1
		_, _, err := user.InsertInto(dbase)
		if err != nil {
			log.Printf("%s\n%s\n", err, debug.Stack())
			return
		}
	} else {
		users[0].RequestCount++
		_, _, err := users[0].Update(dbase)
		if err != nil {
			log.Printf("%s\n%s\n", err, debug.Stack())
			return
		}
	}
}

func handlingDistrictSelection(accessToken string, messageData tg_api.Message) (int, error) {
	m := ""
	district := 0
	if district, m = checkDistrict(messageData.Text); district > 0 {
		m = fmt.Sprintf("%s Для получения прогноза погоды отправьте дату "+
			"нужного прогноза в формате ГГГГ-ММ-ДД.\nНапример (без кавычек): «%s».",
			m, unixTimestampToHumanReadableFormat(time.Now().Unix()))
		if err := sendHint(accessToken, m, messageData.Chat.ID); err != nil {
			return 0, err
		}
	} else {
		if err := sendHint(accessToken, m, messageData.Chat.ID); err != nil {
			return 0, err
		}
	}

	return district, nil
}

func handlingDateSelection(botConn tools.BotConn, pogodaApiConn tools.PogodaApiConn,
	district int, messageData tg_api.Message) (bool, error) {
	ok := false
	dateRecognized, err := checkDate(messageData.Text)
	if err != nil {
		return false, err
	}
	if dateRecognized {
		forecast, err := pogoda_api.GetForecast(pogodaApiConn.PogodaApiURL, "1", messageData.Text)
		if err != nil {
			return false, err
		}
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
			if err := sendWeatherForecast(botConn, localForecast, messageData.Chat.ID); err != nil {
				return false, err
			}
		} else {
			if err := sendMessageAboutForecastUnavailable(botConn, messageData); err != nil {
				return false, err
			}
		}
		ok = true
	} else {
		m := fmt.Sprintf("Дата не распознана. "+
			"Для получения прогноза погоды отправьте дату "+
			"нужного прогноза в формате ГГГГ-ММ-ДД.\nНапример (без кавычек): «%s».",
			unixTimestampToHumanReadableFormat(time.Now().Unix()))
		if err := sendHint(botConn.AccessToken, m, messageData.Chat.ID); err != nil {
			return false, err
		}
	}

	return ok, nil
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

func checkDate(message string) (bool, error) {
	if len([]rune(message)) == 10 {
		matched, err := regexp.MatchString("[0-9]{4}-[0-9]{2}-[0-9]{2}", message)
		if err != nil {
			return false, err
		}
		return matched, nil
	}

	return false, nil
}

func checkWeatherForecast(localForecast pogoda_api.Weather) bool {
	return localForecast.Date != ""
}

func sendHint(accessToken, m string, chatId int) error {
	values := url.Values{
		"chat_id": {strconv.Itoa(chatId)},
		"text":    {m},
	}
	if err := tg_api.SendMessage(accessToken, values); err != nil {
		return err
	}
	return nil
}

func sendWeatherForecast(botConn tools.BotConn, localForecast pogoda_api.Weather, chatID int) error {
	f := ""
	f = makeNightForecastMessage(f, localForecast)
	f = makeDayForecastMessage(f, localForecast)

	values := url.Values{
		"chat_id":    {strconv.Itoa(chatID)},
		"text":       {f},
		"parse_mode": {"Markdown"},
	}
	if err := tg_api.SendMessage(botConn.AccessToken, values); err != nil {
		return err
	}
	return nil
}

func makeNightForecastMessage(f string, localForecast pogoda_api.Weather) string {
	if localForecast.NightCloud != "" && localForecast.NightPrec != "" {
		f += fmt.Sprintf("Ночью *%s*, *%s*.", localForecast.NightCloud, localForecast.NightPrec)
	} else {
		switch true {
		case localForecast.NightCloud != "":
			f += fmt.Sprintf("Ночью *%s*.", localForecast.NightCloud)
		case localForecast.NightPrec != "":
			f += fmt.Sprintf("Ночью *%s*.", localForecast.NightPrec)
		}
	}

	if localForecast.NightPrecComm != "" {
		f += fmt.Sprintf(" *%s*.\n", localForecast.NightPrecComm)
	} else {
		f += "\n"
	}

	if localForecast.NightTemp != "" {
		f += fmt.Sprintf("Температура ночью *%s°C*",
			strings.ReplaceAll(localForecast.NightTemp, ",", "..."))
	}

	if localForecast.NightTempComm != "" {
		f += fmt.Sprintf(", *%sC*.\n",
			strings.ToLower(strings.ReplaceAll(localForecast.NightTempComm, ",", "...")))
	} else {
		f += ".\n"
	}

	if localForecast.NightWindDirrect != "" && localForecast.NightWindSpeed != "" {
		f += fmt.Sprintf("Ветер *%s*, *%s м/с*. ",
			localForecast.NightWindDirrect, localForecast.NightWindSpeed)
	}

	if localForecast.NightWindComm != "" {
		f += fmt.Sprintf("*%s*.\n", localForecast.NightWindComm)
	} else {
		f += "\n"
	}

	return f
}

func makeDayForecastMessage(f string, localForecast pogoda_api.Weather) string {
	if localForecast.DayCloud != "" && localForecast.DayPrec != "" {
		f += fmt.Sprintf("\nДнем *%s*, *%s*.", localForecast.DayCloud, localForecast.DayPrec)
	} else {
		switch true {
		case localForecast.DayCloud != "":
			f += fmt.Sprintf("Днем *%s*.", localForecast.DayCloud)
		case localForecast.DayPrec != "":
			f += fmt.Sprintf("Днем *%s*.", localForecast.DayPrec)
		}
	}

	if localForecast.DayPrecComm != "" {
		f += fmt.Sprintf(" *%s*.\n", localForecast.DayPrecComm)
	} else {
		f += "\n"
	}

	if localForecast.DayTemp != "" {
		f += fmt.Sprintf("Температура днем *%s°C*",
			strings.ReplaceAll(localForecast.DayTemp, ",", "..."))
	}

	if localForecast.DayTempComm != "" {
		f += fmt.Sprintf(", *%sC*.\n",
			strings.ToLower(strings.ReplaceAll(localForecast.DayTempComm, ",", "...")))
	} else {
		f += ".\n"
	}

	if localForecast.DayWindDirrect != "" && localForecast.DayWindSpeed != "" {
		f += fmt.Sprintf("Ветер *%s*, *%s м/с*. ",
			localForecast.DayWindDirrect, localForecast.DayWindSpeed)
	}

	if localForecast.DayWindComm != "" {
		f += fmt.Sprintf("*%s*.", localForecast.DayWindComm)
	}

	return f
}

func unixTimestampToHumanReadableFormat(ut int64) string {
	t := time.Unix(ut, 0)
	dateFormat := "2006-01-02"
	date := t.Format(dateFormat)
	return date
}

func sendMessageAboutForecastUnavailable(botConn tools.BotConn, messageData tg_api.Message) error {
	m := "Не удалось получить прогноз на указанную дату."
	values := url.Values{
		"chat_id": {strconv.Itoa(messageData.Chat.ID)},
		"text":    {m},
	}
	if err := tg_api.SendMessage(botConn.AccessToken, values); err != nil {
		return err
	}
	return nil
}

func sendMessageAboutError(m, accessToken string, chatID int) error {
	m += " Попробуйте позже..."
	values := url.Values{
		"chat_id": {strconv.Itoa(chatID)},
		"text":    {m},
	}
	if err := tg_api.SendMessage(accessToken, values); err != nil {
		return err
	}
	return nil
}
