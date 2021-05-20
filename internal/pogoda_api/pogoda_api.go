package pogoda_api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
)

func GetForecast(pogodaApiURL, region, date string) (Forecast, error) {
	values := map[string]string{
		"region": region,
		"date":   date,
	}

	u := makeURL(pogodaApiURL, "GetForecast", values)
	rawData, err := sendRequest(u)
	if err != nil {
		return Forecast{}, err
	}

	forecast, err := parseForecastData(rawData)
	if err != nil {
		return Forecast{}, err
	}
	return forecast, nil
}

func GetSynoptic(pogodaApiURL, region, date string, id int) (Synoptic, error) {
	values := map[string]string{
		"region": region,
	}

	u := makeURL(pogodaApiURL, "GetSynopticList", values)
	rawData, err := sendRequest(u)
	if err != nil {
		return Synoptic{}, err
	}

	synoptic, err := parseSynopticData(rawData, strconv.Itoa(id))
	if err != nil {
		return Synoptic{}, err
	}

	return synoptic, nil
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
	OrenburgOblast  Weather `json:"182"`
	Orenburg        Weather `json:"111"`
	Buzuluk         Weather `json:"106"`
	Orsk            Weather `json:"112"`
	PenzaOblast     Weather `json:"183"`
	Penza           Weather `json:"154"`
	SamaraOblast    Weather `json:"184"`
	Samara          Weather `json:"1"`
	Tolyatti        Weather `json:"9"`
	Syzran          Weather `json:"8"`
	SaratovOblast   Weather `json:"185"`
	Saratov         Weather `json:"38"`
	UlyanovskOblast Weather `json:"186"`
	Ulyanovsk       Weather `json:"80"`
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
	DayPrecComm      string `json:"daypreccomm"`
	DayPrecVision    bool   `json:"dayprecvision"`
	DayWindDirrect   string `json:"daywinddirrect"`
	DayWindSpeed     string `json:"daywindspeed"`
	DayWindComm      string `json:"daywindcomm"`
	DayTemp          string `json:"daytemp"`
	DayTempComm      string `json:"daytempcomm"`
	DayCommonComm    string `json:"daycommoncomm"`
	Author           int    `json:"author"`
}

func parseForecastData(rawData []byte) (Forecast, error) {
	var forecast Forecast
	err := json.Unmarshal(rawData, &forecast)
	if err != nil {
		return Forecast{}, err
	}
	return forecast, nil
}

type Synoptic struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Position  string `json:"position"`
}

func parseSynopticData(rawData []byte, id string) (Synoptic, error) {
	var f map[string]interface{}
	err := json.Unmarshal(rawData, &f)
	if err != nil {
		return Synoptic{}, nil
	}

	s := f[id].(map[string]interface{})

	var synoptic Synoptic
	synoptic.ID = int(s["id"].(float64))
	synoptic.FirstName = s["first_name"].(string)
	synoptic.LastName = s["last_name"].(string)
	synoptic.Position = s["position"].(string)

	return synoptic, nil
}

func sendRequest(u string) ([]byte, error) {
	response, err := http.Get(u)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Printf("%s\n\n%s\n", err, debug.Stack())
		}
	}()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
