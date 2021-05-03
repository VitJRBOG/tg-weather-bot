package pogoda_api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func GetForecast(pogodaApiURL, region, date string) Forecast {
	values := map[string]string{
		"region": region,
		"date":   date,
	}

	u := makeURL(pogodaApiURL, "GetForecast", values)
	rawData := sendRequest(u)

	forecast := parseForecastData(rawData)
	return forecast
}

func makeURL(pogodaApiURL, method string, values map[string]string) string {
	u := fmt.Sprintf("%s/forecast/api/%s", pogodaApiURL, method)

	firstItem := true
	for k, v := range values {
		if firstItem {
			u += fmt.Sprintf("?%s=%s", k, v)
			firstItem = false
		} else {
			u += fmt.Sprintf("&%s=%s", k, v)
		}
	}

	return u
}

type Forecast struct {
	// TODO: уточнить города
	Orenburg       Weather `json:"111"` // но это не точно, возможно Бузулук
	Buzuluk        Weather `json:"106"` // но это не точно, возможно Оренбург
	Orsk           Weather `json:"112"`
	OrenburgOblast Weather `json:"182"`
}

type Weather struct {
	Date             string `json:"date"`
	NightCloud       string `json:"nightcloud"`
	NightPrec        string `json:"nightprec"`
	NightPrecComm    string `json:"nightpreccomm"`
	NightPrecVision  bool   `json:"nightprecvision"`
	NightWindDirrect string `json:"nightwinddirrect"`
	NightWindSpeed   string `json:"nightwindspeed"`
	NightWindComm    string `json:"nightwindcomm"`
	NightTemp        string `json:"nighttemp"`
	NightTempComm    string `json:"nighttempcomm"`
	NightCommonComm  string `json:"nightcommoncomm"`
	DayCloud         string `json:"daycloud"`
	DayPrec          string `json:"dayprec"`
	DayPrecVision    bool   `json:"dayprecvision"`
	DayWindDirrect   string `json:"daywinddirrect"`
	DayWindSpeed     string `json:"daywindspeed"`
	DayWindComm      string `json:"daywindcomm"`
	DayTemp          string `json:"daytemp"`
	DayTempComm      string `json:"daytempcomm"`
	DayCommonComm    string `json:"daycommoncomm"`
}

func parseForecastData(rawData []byte) Forecast {
	var forecast Forecast
	err := json.Unmarshal(rawData, &forecast)
	if err != nil {
		panic(err.Error())
	}
	return forecast
}

func sendRequest(u string) []byte {
	response, err := http.Get(u)
	if err != nil {
		panic(err.Error())
	}

	defer func() {
		err := response.Body.Close()
		if err != nil {
			panic(err.Error())
		}
	}()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err.Error())
	}
	return body
}
