package tools

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

type BotConn struct {
	AccessToken   string `json:"access_token"`
	UpdatesOffset int    `json:"updates_offset"`
	Timeout       int    `json:"timeout"`
}

func (c *BotConn) UpdateFile() error {
	path, err := getPath("configs/bot_conn.json")
	if err != nil {
		return err
	}

	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, data, 0644)
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
	path, err := getPath("configs/bot_conn.json")
	if err != nil {
		return BotConn{}, err
	}

	var c BotConn
	data, err := ioutil.ReadFile(path)
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
	path, err := getPath("configs/db_conn.json")
	if err != nil {
		return DBConn{}, err
	}

	var c DBConn
	data, err := ioutil.ReadFile(path)
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
	path, err := getPath("configs/pogoda_api_conn.json")
	if err != nil {
		return PogodaApiConn{}, err
	}

	var c PogodaApiConn
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return PogodaApiConn{}, err
	}
	err = json.Unmarshal(data, &c)
	if err != nil {
		return PogodaApiConn{}, err
	}

	return c, err
}

func getPath(localPath string) (string, error) {
	absPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}

	pathToPath := filepath.FromSlash(absPath + "/path.txt")

	ok, err := checkFileExistence(pathToPath)
	if err != nil {
		return "", err
	}

	if ok {
		path, err := readTextFile(pathToPath)
		if err != nil {
			return "", err
		}
		return strings.ReplaceAll(path, "\n", "") + "/" + localPath, nil
	}

	return filepath.FromSlash(absPath + "/" + localPath), nil
}

func checkFileExistence(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func readTextFile(path string) (string, error) {
	file, err := os.Open(path)
	defer func() {
		err := file.Close()
		if err != nil {
			log.Printf("%s\n%s\n", err, debug.Stack())
		}
	}()
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(file)

	var text string
	for scanner.Scan() {
		text += fmt.Sprintf("%v\n", scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return text, nil
}
