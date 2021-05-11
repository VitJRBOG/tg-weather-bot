package tools

import (
	"encoding/json"
	"io/ioutil"
)

type BotConn struct {
	AccessToken   string `json:"access_token"`
	UpdatesOffset int    `json:"updates_offset"`
	Timeout       int    `json:"timeout"`
}

func (c *BotConn) UpdateFile() error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("../configs/bot_conn.json", data, 0644)
	if err != nil {
		return err
	}

	return nil
}

type DBConn struct {
	Address  string `json:"address"`
	Login    string `json:"login"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
}

type PogodaApiConn struct {
	PogodaApiURL string `json:"pogoda_api_url"`
}

func GetBotConnectionData() (BotConn, error) {
	var c BotConn
	data, err := ioutil.ReadFile("../configs/bot_conn.json")
	if err != nil {
		return BotConn{}, err
	}
	err = json.Unmarshal(data, &c)
	if err != nil {
		return BotConn{}, err
	}

	return c, err
}

func GetDBConnectionData() (DBConn, error) {
	var c DBConn
	data, err := ioutil.ReadFile("../configs/db_conn.json")
	if err != nil {
		return DBConn{}, err
	}
	err = json.Unmarshal(data, &c)
	if err != nil {
		return DBConn{}, err
	}

	return c, err
}

func GetPogodaAPIConnectionData() (PogodaApiConn, error) {
	var c PogodaApiConn
	data, err := ioutil.ReadFile("../configs/pogoda_api_conn.json")
	if err != nil {
		return PogodaApiConn{}, err
	}
	err = json.Unmarshal(data, &c)
	if err != nil {
		return PogodaApiConn{}, err
	}

	return c, err
}
