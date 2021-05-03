package tg_api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Result struct {
	OK      bool     `json:"ok"`
	Updates []Update `json:"result"`
}

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
	ID int `json:"id"`
}

type Chat struct {
	ID int `json:"id"`
}

func GetUpdates(accessToken, method string, values url.Values) Result {
	u := makeURL(accessToken, method)
	rawData := sendRequest(u, values)
	result := parseUpdates(rawData)
	return result
}

func SendMessage(accessToken string, values url.Values) {
	u := makeURL(accessToken, "sendMessage")
	sendRequest(u, values)
}

func makeURL(accessToken, method string) string {
	u := fmt.Sprintf("https://api.telegram.org/bot%s/%s", accessToken, method)

	return u
}

func parseUpdates(rawData []byte) Result {
	var result Result
	err := json.Unmarshal(rawData, &result)
	if err != nil {
		panic(err.Error())
	}
	return result
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
