# coding: utf-8


import datetime
import threading
import queue
import re
from typing import Tuple
import tools.tools as tools
import tools.dtime as dtime
import tg_api.tg_api as tg_api
import pogoda_api.pogoda_api as pogoda_api
import bot.text_forecast as text_forecast

queues = {}


def start(bot_conf: tools.BotConf, pogoda_api_url: str):
    __chat_listening(bot_conf, pogoda_api_url)


def __chat_listening(bot_conf: tools.BotConf, pogoda_api_url: str):
    params = tg_api.UpdatesRequestParams()

    while True:
        params.set_offset(bot_conf.get_updates_offset())
        params.set_timeout(bot_conf.get_timeout())

        response = tg_api.get_updates(bot_conf.get_access_token(), params)

        if response == {}:
            try:
                raise Exception("ERROR: dict response is empty")
            except Exception:
                continue

        if len(response["result"]) > 0:
            for update_data in response["result"]:
                sender_id = update_data["message"]["from"]["id"]

                global queues

                q = queues.get(sender_id)
                if q is None:
                    new_q = queue.Queue()
                    t = threading.Thread(target=__user_message_processing,
                                         args=(
                                             new_q, update_data,
                                             pogoda_api_url,
                                             bot_conf,))
                    t.start()
                    queues.update({sender_id: new_q})
                else:
                    q.put(update_data)

                bot_conf = __upd_updates_offset(bot_conf,
                                                update_data["update_id"] + 1)


def __user_message_processing(q: queue.Queue,
                              update_data: dict,
                              pogoda_api_url: str,
                              bot_conf: tools.BotConf):
    global queues

    ok, region, location = __check_command(
        update_data["message"]["text"])

    if ok:
        text = "Запрос прогноза погоды по " + \
               "*%s*. Выберите дату из списка:\n\n%s" % (
                   __get_location_name(location), __list_of_dates_preparing())

        tg_api.send_message(bot_conf.get_access_token(),
                            __compose_hint_msg(
                                update_data["message"], text))

        while True:
            try:
                update_data = q.get(block=True, timeout=30)
            except queue.Empty:
                text = "Время ожидания ответа закончилось."
                tg_api.send_message(bot_conf.get_access_token(),
                                    __compose_hint_msg(update_data["message"],
                                                       text))
                queues.pop(update_data["message"]["from"]["id"])
                return

            if __check_date(update_data["message"]["text"]):
                date = __convert_date(update_data["message"]["text"][1:])
                msg_values = __forecast_preparing(pogoda_api_url,
                                                  region, location, date,
                                                  update_data["message"])
                tg_api.send_message(bot_conf.get_access_token(),
                                    msg_values)
                queues.pop(update_data["message"]["from"]["id"])
                return
            else:
                text = "Ошибка при вводе даты. " + \
                       "Пожалуйста, выберите дату из списка."
                tg_api.send_message(bot_conf.get_access_token(),
                                    __compose_hint_msg(
                                        update_data["message"], text))

    if not ok:
        text = "Ошибка при обработке команды. " + \
               "Пожалуйста, проверьте текст сообщения " + \
               "и повторите попытку."
        tg_api.send_message(bot_conf.get_access_token(),
                            __compose_hint_msg(
                                update_data["message"], text))
        queues.pop(update_data["message"]["from"]["id"])


def __check_command(text: str) -> Tuple[bool, str, int]:
    if text == "/orenburg_oblast":
        return True, "1", 182
    elif text == "/orenburg":
        return True, "1", 111
    elif text == "/buzuluk":
        return True, "1", 106
    elif text == "/orsk":
        return True, "1", 112
    elif text == "/penza_oblast":
        return True, "2", 183
    elif text == "/penza":
        return True, "2", 154
    elif text == "/samara_oblast":
        return True, "3", 184
    elif text == "/samara":
        return True, "3", 1
    elif text == "/tolyatti":
        return True, "3", 9
    elif text == "/syzran":
        return True, "3", 8
    elif text == "/saratov_oblast":
        return True, "4", 185
    elif text == "/saratov":
        return True, "4", 38
    elif text == "/ulyanovsk_oblast":
        return True, "5", 186
    elif text == "/ulyanovsk":
        return True, "5", 80
    else:
        return False, "0", 0


def __check_date(text: str) -> bool:
    if re.match("[0-9]{8}", text[1:]) is None:
        return False
    return True


def __convert_date(date_str: str) -> str:
    d = datetime.datetime.strptime(date_str, "%Y%m%d")

    return d.strftime("%Y-%m-%d")


def __get_location_name(location: int) -> str:
    if location == 182:
        return "Оренбургской области"
    elif location == 111:
        return "Оренбургу"
    elif location == 106:
        return "Бузулуку"
    elif location == 112:
        return "Орску"
    elif location == 183:
        return "Пензенской области"
    elif location == 154:
        return "Пензе"
    elif location == 184:
        return "Самарской области"
    elif location == 1:
        return "Самаре"
    elif location == 9:
        return "Тольятти"
    elif location == 8:
        return "Сызрани"
    elif location == 185:
        return "Саратовской области"
    elif location == 38:
        return "Саратову"
    elif location == 186:
        return "Ульяновской области"
    elif location == 80:
        return "Ульяновск"
    else:
        return "[region_error]"


def __list_of_dates_preparing() -> str:
    today = datetime.datetime.now()
    text = ""

    for i in range(4):
        d = today + datetime.timedelta(i)
        text += "Прогноз на *%s* - /%s\n" % (
            __prepare_date(d.strftime("%Y-%m-%d")), d.strftime("%Y%m%d"))

    return text


def __prepare_date(date: str) -> str:
    date = dtime.convert_date_to_words(date)
    date = dtime.eng_month_to_rus(date)
    date = dtime.eng_dweek_to_rus(date)

    return date


def __forecast_preparing(pogoda_api_url: str, region: str,
                         location: int, date: str,
                         msg_data: dict) -> tg_api.MessageRequestParams:
    forecast, ok = pogoda_api.get_forecast(pogoda_api_url, region, location,
                                           date)

    if ok:
        synoptic, ok = pogoda_api.get_synoptic_data(pogoda_api_url, region,
                                                    str(forecast.get_author()))

        if ok:
            msg_values = text_forecast.combine_msg_values(msg_data, forecast,
                                                          synoptic)
            return msg_values

    text = "При получении прогноза произошла ошибка. Попробуйте позже..."
    msg_values = __compose_hint_msg(msg_data, text)

    return msg_values


def __compose_hint_msg(msg_data: dict,
                       text: str) -> tg_api.MessageRequestParams:
    v = tg_api.MessageRequestParams()
    v.set_chat_id(msg_data["from"]["id"])
    v.set_text(text)

    return v


def __upd_updates_offset(bot_conf: tools.BotConf,
                         new_offset: int) -> tools.BotConf:
    bot_conf.set_updates_offset(new_offset)
    tools.upd_bot_conf(bot_conf)

    return bot_conf
