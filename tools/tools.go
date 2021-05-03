package tools

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	AccessToken   string `json:"access_token"`
	UpdatesOffset int    `json:"updates_offset"`
	PogodaApiURL  string `json:"pogoda_api_url"`
}

func (c *Config) UpdateUpdatesOffset(newDate int) {
	c.UpdatesOffset = newDate

	content, err := json.Marshal(c)
	if err != nil {
		panic(err.Error())
	}
	ioutil.WriteFile("config.json", content, 0644)
}

func GetConfig() Config {
	var c Config
	content, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(err.Error())
	}
	err = json.Unmarshal(content, &c)
	if err != nil {
		panic(err.Error())
	}

	return c
}
