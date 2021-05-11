package tg_api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
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

func GetUpdates(accessToken, method string, values url.Values) ([]Update, error) {
	u := makeURL(accessToken, method)

	rawData, err := sendRequest(u, values)
	if err != nil {
		return []Update{}, err
	}

	updates, err := parseUpdates(rawData)
	if err != nil {
		return []Update{}, err
	}

	return updates, nil
}

func SendMessage(accessToken string, values url.Values) error {
	u := makeURL(accessToken, "sendMessage")
	_, err := sendRequest(u, values)
	if err != nil {
		return err
	}
	return nil
}

func makeURL(accessToken, method string) string {
	u := fmt.Sprintf("https://api.telegram.org/bot%s/%s", accessToken, method)

	return u
}

func parseUpdates(rawData []byte) ([]Update, error) {
	var data struct {
		OK          bool     `json:"ok"`
		Updates     []Update `json:"result"`
		ErrorCode   int      `json:"error_code"`
		Description string   `json:"description"`
	}
	err := json.Unmarshal(rawData, &data)
	if err != nil {
		return []Update{}, err
	}

	if data.OK {
		return data.Updates, nil
	}

	return []Update{}, fmt.Errorf("error %d: %s", data.ErrorCode, data.Description)
}

func sendRequest(u string, values url.Values) ([]byte, error) {
	response, err := http.PostForm(u, values)
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
