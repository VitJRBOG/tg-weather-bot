# coding: utf-8
import json
import requests


class UpdatesRequestParams(object):
    def __init__(self):
        self.__offset = 0
        self.__timeout = 0

    def set_offset(self, offset: int):
        self.__offset = offset

    def set_timeout(self, timeout: int):
        self.__timeout = timeout

    def get_offset(self):
        return self.__offset

    def get_timeout(self):
        return self.__timeout


def get_updates(access_token: str, params: UpdatesRequestParams):
    u = __make_url(access_token, "getUpdates")

    p = {
        "offset": params.get_offset(),
        "timeout": params.get_timeout()
    }

    data, ok = __send_request(u, p)
    if ok:
        return json.loads(data)


class MessageRequestParams(object):
    def __init__(self):
        self.__chat_id = 0
        self.__text = "Test message"
        self.__parse_mode = "Markdown"

    def set_chat_id(self, chat_id: int):
        self.__chat_id = chat_id

    def set_text(self, text: str):
        self.__text = text

    def set_parse_mode(self, parse_mode: str):
        self.__parse_mode = parse_mode

    def get_chat_id(self):
        return self.__chat_id

    def get_text(self):
        return self.__text

    def get_parse_mode(self):
        return self.__parse_mode


def send_message(access_token: str, params: MessageRequestParams):
    u = __make_url(access_token, "sendMessage")

    p = {
        "chat_id": params.get_chat_id(),
        "text": params.get_text(),
        "parse_mode": params.get_parse_mode()
    }

    data, ok = __send_request(u, p)
    if ok:
        return json.loads(data)


def __make_url(access_token: str, method: str):
    return "https://api.telegram.org/bot%s/%s" % (access_token, method)


def __send_request(u: str, p: dict):
    foo = requests.post(u, p)
    if foo.status_code == 200:
        return foo.text, True
    else:
        print("Error: " + foo.text)
        return "", False
