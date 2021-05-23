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

		updates, err := tg_api.GetUpdates(botConn.AccessToken, values)
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
	region := "0"
	district := 0
	for {
		messageData := <-channel

		if dbase != nil {
			go updateDB(dbase, messageData.From)
		}

		if district == 0 {
			var err error
			region, district, err = handlingDistrictSelection(botConn.AccessToken, messageData)
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
			ok, err := handlingDateSelection(botConn, pogodaApiConn, region, district, messageData)
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

func handlingDistrictSelection(accessToken string, messageData tg_api.Message) (string, int, error) {
	m := ""
	region := "0"
	district := 0
	if region, district, m = checkDistrict(messageData.Text); district > 0 {
		timezone := getTimezone(region)
		dates := getDates(timezone)
		m = fmt.Sprintf("%s Для получения прогноза погоды выберите дату из списка:"+
			"\n\nПрогноз на *%s* - /%s\nПрогноз на *%s* - /%s"+
			"\nПрогноз на *%s* - /%s\nПрогноз на *%s* - /%s",
			m, dates[0][1], dates[0][0],
			dates[1][1], dates[1][0],
			dates[2][1], dates[2][0],
			dates[3][1], dates[3][0])
		if err := sendHint(accessToken, m, messageData.Chat.ID); err != nil {
			return "0", 0, err
		}
	} else {
		if err := sendHint(accessToken, m, messageData.Chat.ID); err != nil {
			return "0", 0, err
		}
	}

	return region, district, nil
}

func handlingDateSelection(botConn tools.BotConn, pogodaApiConn tools.PogodaApiConn,
	region string, district int, messageData tg_api.Message) (bool, error) {
	ok := false
	dateRecognized, err := checkDate(messageData.Text)
	if err != nil {
		return false, err
	}
	timezone := getTimezone(region)
	dates := getDates(timezone)
	if dateRecognized {
		date := insertDashes([]rune(messageData.Text)[1:])
		forecast, err := pogoda_api.GetForecast(pogodaApiConn.PogodaApiURL, region, date)
		if err != nil {
			return false, err
		}
		localForecast := selectLocalForecast(forecast, district)
		forecastAvailable := checkWeatherForecast(localForecast)
		if forecastAvailable {
			dateInWords := ""
			indexOfDateInWords := findDateInWords(dates, string([]rune(messageData.Text)[1:]))
			if indexOfDateInWords > -1 {
				dateInWords = dates[indexOfDateInWords][1]
			}
			var authorSignature string
			if localForecast.Author != 0 {
				synoptic, err := pogoda_api.GetSynoptic(pogodaApiConn.PogodaApiURL, region,
					date, localForecast.Author)
				if err != nil {
					return false, err
				}
				authorSignature = makeAuthorSignature(synoptic)
			}
			forecastText := makeForecastMessage(localForecast, dateInWords, authorSignature)
			if err := sendWeatherForecast(botConn, messageData.Chat.ID, forecastText); err != nil {
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
			"Для получения прогноза погоды выберите дату из списка:"+
			"\n\nПрогноз на *%s* - /%s\nПрогноз на *%s* - /%s"+
			"\nПрогноз на *%s* - /%s\nПрогноз на *%s* - /%s",
			dates[0][1], dates[0][0],
			dates[1][1], dates[1][0],
			dates[2][1], dates[2][0],
			dates[3][1], dates[3][0])
		if err := sendHint(botConn.AccessToken, m, messageData.Chat.ID); err != nil {
			return false, err
		}
	}

	return ok, nil
}

func insertDashes(date []rune) string {
	var d []rune
	d = append(d, []rune(date)[:4]...)
	d = append(d, '-')
	d = append(d, []rune(date)[4:6]...)
	d = append(d, '-')
	d = append(d, []rune(date)[6:]...)
	return string(d)
}

func selectLocalForecast(forecast pogoda_api.Forecast, district int) pogoda_api.Weather {
	switch district {
	case 182:
		return forecast.OrenburgOblast
	case 111:
		return forecast.Orenburg
	case 106:
		return forecast.Buzuluk
	case 112:
		return forecast.Orsk
	case 183:
		return forecast.PenzaOblast
	case 154:
		return forecast.Penza
	case 184:
		return forecast.SamaraOblast
	case 1:
		return forecast.Samara
	case 9:
		return forecast.Tolyatti
	case 8:
		return forecast.Syzran
	case 185:
		return forecast.SaratovOblast
	case 38:
		return forecast.Saratov
	case 186:
		return forecast.UlyanovskOblast
	case 80:
		return forecast.Ulyanovsk
	}
	return pogoda_api.Weather{}
}

func checkDistrict(message string) (string, int, string) {
	switch true {
	case message == "/orenburg_oblast":
		return "1", 182, "Запрос прогноза погоды по *Оренбургской области*."
	case message == "/orenburg":
		return "1", 111, "Запрос прогноза погоды по *Оренбургу*."
	case message == "/buzuluk":
		return "1", 106, "Запрос прогноза погоды по *Бузулуку*."
	case message == "/orsk":
		return "1", 112, "Запрос прогноза погоды по *Орску*."
	case message == "/penza_oblast":
		return "2", 183, "Запрос прогноза погоды по *Пензенской области*."
	case message == "/penza":
		return "2", 154, "Запрос прогноза погоды по *Пензе*."
	case message == "/samara_oblast":
		return "3", 184, "Запрос прогноза погоды по *Самарской области*."
	case message == "/samara":
		return "3", 1, "Запрос прогноза погоды по *Самаре*."
	case message == "/tolyatti":
		return "3", 9, "Запрос прогноза погоды по *Тольятти*."
	case message == "/syzran":
		return "3", 8, "Запрос прогноза погоды по *Сызрани*."
	case message == "/saratov_oblast":
		return "4", 185, "Запрос прогноза погоды по *Саратовской области*."
	case message == "/saratov":
		return "4", 38, "Запрос прогноза погоды по *Саратову*."
	case message == "/ulyanovsk_oblast":
		return "5", 186, "Запрос прогноза погоды по *Ульяновской области*."
	case message == "/ulyanovsk":
		return "5", 80, "Запрос прогноза погоды по *Ульяновску*."
	}
	return "0", 0, "Для выбора региона/города введите «/» («слэш», без кавычек)..."
}

func checkDate(message string) (bool, error) {
	if len([]rune(message)) == 9 {
		matched, err := regexp.MatchString("/[0-9]{8}", message)
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

func findDateInWords(dates [][]string, dateInDigits string) int {
	for i, d := range dates {
		if d[0] == dateInDigits {
			return i
		}
	}
	return -1
}

func sendHint(accessToken, m string, chatId int) error {
	values := url.Values{
		"chat_id":    {strconv.Itoa(chatId)},
		"text":       {m},
		"parse_mode": {"Markdown"},
	}
	if err := tg_api.SendMessage(accessToken, values); err != nil {
		return err
	}
	return nil
}

func sendWeatherForecast(botConn tools.BotConn, chatID int, forecastText string) error {
	values := url.Values{
		"chat_id":    {strconv.Itoa(chatID)},
		"text":       {forecastText},
		"parse_mode": {"Markdown"},
	}
	if err := tg_api.SendMessage(botConn.AccessToken, values); err != nil {
		return err
	}
	return nil
}

func makeAuthorSignature(synoptic pogoda_api.Synoptic) string {
	return fmt.Sprintf("Прогноз составил(а): %s %s %s",
		synoptic.Position, synoptic.FirstName, synoptic.LastName)
}

func getTimezone(region string) string {
	switch region {
	case "1":
		return "Asia/Yekaterinburg"
	case "2":
		return "Europe/Moscow"
	case "3":
		return "Europe/Samara"
	case "4":
		return "Europe/Samara"
	case "5":
		return "Europe/Samara"
	}
	return ""
}

func getDates(timezone string) [][]string {
	ut := time.Now().Unix()
	days := []int64{1, 86400, 172800, 259200}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Printf("%s\n%s\n", err, debug.Stack())
	}
	var dates [][]string
	for _, d := range days {
		var t time.Time
		if loc != nil {
			t = time.Unix(ut+d, 0).In(loc)
		} else {
			t = time.Unix(ut+d, 0)
		}
		date := []string{
			dateInDigits(t),
			engDayOfWeekToRus(engMonthToRus(dateInWords(t))),
		}
		dates = append(dates, date)
	}
	return dates
}

func dateInDigits(t time.Time) string {
	dateFormat := "20060102"
	date := t.Format(dateFormat)
	return date
}

func dateInWords(t time.Time) string {
	dateFormat := "2 January, Monday"
	date := t.Format(dateFormat)
	return date
}

func engMonthToRus(date string) string {
	endMonths := []string{
		"January", "February",
		"March", "April", "May",
		"June", "July", "August",
		"September", "October", "November",
		"December",
	}

	rusMonths := map[string]string{
		"January":   "января",
		"February":  "февраля",
		"March":     "марта",
		"April":     "апреля",
		"May":       "мая",
		"June":      "июня",
		"July":      "июля",
		"August":    "августа",
		"September": "сентября",
		"October":   "октября",
		"November":  "ноября",
		"December":  "декабря",
	}

	for _, m := range endMonths {
		if strings.Contains(date, m) {
			date = strings.Replace(date, m, rusMonths[m], 1)
			break
		}
	}

	return date
}

func engDayOfWeekToRus(date string) string {
	engDayweeks := []string{
		"Monday",
		"Tuesday",
		"Wednesday",
		"Thursday",
		"Friday",
		"Saturday",
		"Sunday",
	}

	rusDayweeks := map[string]string{
		"Monday":    "понедельник",
		"Tuesday":   "вторник",
		"Wednesday": "среда",
		"Thursday":  "четверг",
		"Friday":    "пятница",
		"Saturday":  "суббота",
		"Sunday":    "воскресенье",
	}

	for _, m := range engDayweeks {
		if strings.Contains(date, m) {
			date = strings.Replace(date, m, rusDayweeks[m], 1)
			break
		}
	}

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