package tg_api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Update struct {
	UpdateID int     `json:"update_id"`
	Message  Message `json:"message"`
}

type Message struct {
	MessageID int    `json:"message_id"`
	From      User   `json:"from"`
	Chat      Chat   `json:"chat"`
	Date      int    `json:"date"`
	Text      string `json:"text"`
}

type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type Chat struct {
	ID int `json:"id"`
}

func GetUpdates(accessToken, method string, values url.Values) []Update {
	u := makeURL(accessToken, method)
	rawData := sendRequest(u, values)
	updates := parseUpdates(rawData)
	return updates
}

func SendMessage(accessToken string, values url.Values) {
	u := makeURL(accessToken, "sendMessage")
	sendRequest(u, values)
}

func makeURL(accessToken, method string) string {
	u := fmt.Sprintf("https://api.telegram.org/bot%s/%s", accessToken, method)

	return u
}

func parseUpdates(rawData []byte) []Update {
	var data struct {
		OK          bool     `json:"ok"`
		Updates     []Update `json:"result"`
		ErrorCode   int      `json:"error_code"`
		Description string   `json:"description"`
	}
	err := json.Unmarshal(rawData, &data)
	if err != nil {
		panic(err.Error())
	}

	if data.OK {
		return data.Updates
	}

	panic(fmt.Errorf("error %d: %s", data.ErrorCode, data.Description))
}

func sendRequest(u string, values url.Values) []byte {
	response, err := http.PostForm(u, values)
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
