# coding: utf-8
import json
import os
from typing import Tuple


class BotConf(object):

    def __init__(self):
        self.__access_token = "NoAccessToken"
        self.__updates_offset = 0
        self.__timeout = 25

    def set_access_token(self, access_token: str):
        self.__access_token = access_token

    def set_updates_offset(self, updates_offset: int):
        self.__updates_offset = updates_offset

    def set_timeout(self, timeout: int):
        self.__timeout = timeout

    def get_access_token(self) -> str:
        return self.__access_token

    def get_updates_offset(self) -> int:
        return self.__updates_offset

    def get_timeout(self) -> int:
        return self.__timeout


def upd_bot_conf(bot_conf: BotConf):
    path = __get_path()
    path = __check_last_char(path)

    values = {
        "access_token": bot_conf.get_access_token(),
        "updates_offset": bot_conf.get_updates_offset(),
        "timeout": bot_conf.get_timeout()
    }

    __write_file(path + "bot_conf.json", json.dumps(values, indent=4))


def get_bot_conf() -> Tuple[BotConf, bool]:
    path = __get_path()
    path = __check_last_char(path)
    text, ok = __read_file(path + "bot_conf.json")

    if ok:
        return _parse_bot_conf(text), ok
    return BotConf(), ok


def _parse_bot_conf(text: str) -> BotConf:
    json_loads = json.loads(text)

    bot_conf = BotConf()
    bot_conf.set_access_token(json_loads["access_token"])
    bot_conf.set_updates_offset(json_loads["updates_offset"])
    bot_conf.set_timeout(json_loads["timeout"])

    return bot_conf


def get_pogoda_api_url() -> Tuple[str, bool]:
    path = __get_path()
    path = __check_last_char(path)
    text, ok = __read_file(path + "pogoda_api.txt")

    if ok:
        text = __check_last_char(text)
        return text, ok

    return "", ok


def __get_path() -> str:
    with open("path.txt") as f:
        path = f.read()

    if not f.closed:
        f.close()

    return path


def __check_last_char(text: str) -> str:
    while True:
        if text[len(text)-1] == "\n":
            text = text[:-1]
        else:
            break

    if text[len(text) - 1] != "/":
        text += "/"

    return text


def __read_file(path: str) -> Tuple[str, bool]:
    text = ""
    exist = os.path.exists(path)

    if exist:
        with open(path) as f:
            text = f.read()

        if not f.closed:
            f.close()
    else:
        print("\nError: file \"\" is not exist.")

    return text, exist


def __write_file(path: str, text: str):
    with open(path, "w") as f:
        f.write(text)
        f.close()
