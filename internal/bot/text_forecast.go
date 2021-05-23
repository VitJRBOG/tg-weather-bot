package bot

import (
	"fmt"
	"strings"

	"github.com/VitJRBOG/TelegramWeatherBot/internal/pogoda_api"
)

type weatherForecast struct {
	Text string
}

func (w *weatherForecast) setCloudAndPrecipitation(cloudCondition, precCondition string) {
	if cloudCondition != "" && precCondition != "" {
		cloudCondition = strings.ToUpper(string([]rune(cloudCondition)[0])) +
			strings.ToLower(string([]rune(cloudCondition)[1:]))
		w.Text += fmt.Sprintf("*%s*, *%s*.", cloudCondition, strings.ToLower(precCondition))
	} else {
		switch true {
		case cloudCondition != "":
			cloudCondition = strings.ToUpper(string([]rune(cloudCondition)[0])) +
				strings.ToLower(string([]rune(cloudCondition)[1:]))
			w.Text += fmt.Sprintf("*%s*.", cloudCondition)
		case precCondition != "":
			precCondition = strings.ToUpper(string([]rune(precCondition)[0])) +
				strings.ToLower(string([]rune(precCondition)[1:]))
			w.Text += fmt.Sprintf("*%s*.", precCondition)
		}
	}
}

func (w *weatherForecast) setPrecipitationCommon(precCondition string) {
	if precCondition != "" {
		precCondition = strings.ToUpper(string([]rune(precCondition)[0])) +
			strings.ToLower(string([]rune(precCondition)[1:]))
		w.Text += fmt.Sprintf(" *%s*.\n", precCondition)
	} else {
		w.Text += "\n"
	}
}

func (w *weatherForecast) setTemperature(tempCondition string) {
	if tempCondition != "" {
		tempCondition = strings.ReplaceAll(tempCondition, ",", "...")
		w.Text += fmt.Sprintf("Температура *%s°C*", tempCondition)
	}
}

func (w *weatherForecast) setTemperatureCommon(tempCondition string) {
	if tempCondition != "" {
		tempCondition = strings.ToLower(strings.ReplaceAll(tempCondition, ",", "..."))
		w.Text += fmt.Sprintf(", *%sC*.\n", tempCondition)
	} else {
		w.Text += ".\n"
	}
}

func (w *weatherForecast) setWindDirectionAndSpeed(windDirection, windSpeed string) {
	if windDirection != "" && windSpeed != "" {
		w.Text += fmt.Sprintf("Ветер *%s*, *%s м/с*. ", windDirection, windSpeed)
	}
}

func (w *weatherForecast) setWindCommon(windCondition string) {
	if windCondition != "" {
		w.Text += fmt.Sprintf("*%s*.", windCondition)
	}
}

func makeNightForecastMessage(localForecast pogoda_api.Weather) string {
	var w weatherForecast

	w.setCloudAndPrecipitation(localForecast.NightCloud, localForecast.NightPrec)
	w.setPrecipitationCommon(localForecast.NightPrecComm)
	w.setTemperature(localForecast.NightTemp)
	w.setTemperatureCommon(localForecast.NightTempComm)
	w.setWindDirectionAndSpeed(localForecast.NightWindDirrect, localForecast.NightWindSpeed)
	w.setWindCommon(localForecast.NightWindComm)

	return w.Text
}

func makeDayForecastMessage(localForecast pogoda_api.Weather) string {
	var w weatherForecast

	w.setCloudAndPrecipitation(localForecast.DayCloud, localForecast.DayPrec)
	w.setPrecipitationCommon(localForecast.DayPrecComm)
	w.setTemperature(localForecast.DayTemp)
	w.setTemperatureCommon(localForecast.DayTempComm)
	w.setWindDirectionAndSpeed(localForecast.DayWindDirrect, localForecast.DayWindSpeed)
	w.setWindCommon(localForecast.DayWindComm)

	return w.Text
}

func makeForecastMessage(localForecast pogoda_api.Weather, dateInWords, forecastAuthor string) string {
	nightDate := ""
	if dateInWords != "" {
		nightDate = fmt.Sprintf(" на %s", dateInWords)
	}
	signature := ""
	if forecastAuthor != "" {
		signature = fmt.Sprintf("\n\n_%s_", forecastAuthor)
	}
	f := fmt.Sprintf("_НОЧЬ%s_\n%s\n\n_ДЕНЬ_\n%s%s",
		nightDate,
		makeNightForecastMessage(localForecast), makeDayForecastMessage(localForecast), signature)

	return f
}
