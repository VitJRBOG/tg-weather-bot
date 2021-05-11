package tools

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	AccessToken   string       `json:"access_token"`
	UpdatesOffset int          `json:"updates_offset"`
	Timeout       int          `json:"timeout"`
	PogodaApiURL  string       `json:"pogoda_api_url"`
	DBConnection  DBConnection `json:"db_connection"`
}

type DBConnection struct {
	Address  string `json:"address"`
	Login    string `json:"login"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
}

func (c *Config) UpdateUpdatesOffset(newDate int) error {
	c.UpdatesOffset = newDate

	content, err := json.Marshal(c)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("../configs/config.json", content, 0644)
	if err != nil {
		return err
	}

	return nil
}

func GetConfig() (Config, error) {
	var c Config
	content, err := ioutil.ReadFile("../configs/config.json")
	if err != nil {
		return Config{}, err
	}
	err = json.Unmarshal(content, &c)
	if err != nil {
		return Config{}, err
	}

	return c, err
}
