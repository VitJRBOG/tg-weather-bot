# coding: utf-8


import tools.dtime as dtime
import tg_api.tg_api as tg_api
import pogoda_api.pogoda_api as pogoda_api


def combine_msg_values(msg_data: dict, forecast: pogoda_api.Forecast,
                       synoptic: pogoda_api.Synoptic) -> tg_api.MessageRequestParams:
    v = tg_api.MessageRequestParams()
    v.set_chat_id(msg_data["from"]["id"])

    msg_text = __compose_msg_text(
        forecast, __prepare_date(forecast.get_date()), synoptic)
    msg_text = __erase_redundant_spaces(msg_text)

    v.set_text(msg_text)

    return v


def __compose_msg_text(forecast: pogoda_api.Forecast, date: str,
                       synoptic: pogoda_api.Synoptic) -> str:
    night_forecast = ""
    if forecast.get_night_prec_vision():
        night_forecast += __add_cloud_and_prec(forecast.get_night_cloud(),
                                               forecast.get_night_prec())
    else:
        night_forecast += __add_cloud_and_prec(forecast.get_night_cloud(), "")
    night_forecast += __add_prec_common(forecast.get_night_prec_comm())
    night_forecast += __add_temp(forecast.get_night_temp())
    night_forecast += __add_temp_comm(forecast.get_night_temp_comm())
    night_forecast += __add_wind_direction_and_speed(
        forecast.get_night_wind_direction(), forecast.get_night_wind_speed())
    night_forecast += __add_wind_comm(forecast.get_night_wind_comm())

    day_forecast = ""
    if forecast.get_day_prec_vision():
        day_forecast += __add_cloud_and_prec(
            forecast.get_day_cloud(), forecast.get_day_prec())
    else:
        day_forecast += __add_cloud_and_prec(forecast.get_day_cloud(), "")
    day_forecast += __add_prec_common(forecast.get_day_prec_comm())
    day_forecast += __add_temp(forecast.get_day_temp())
    day_forecast += __add_temp_comm(forecast.get_day_temp_comm())
    day_forecast += __add_wind_direction_and_speed(
        forecast.get_day_wind_direction(), forecast.get_day_wind_speed())
    day_forecast += __add_wind_comm(forecast.get_day_wind_comm())

    author = __add_author(synoptic)

    forecast_text = "_???????? ???? %s_\n%s\n\n_????????_\n%s%s" % (
        date, night_forecast, day_forecast, author)

    return forecast_text


def __add_cloud_and_prec(cloud: str, prec: str) -> str:
    if cloud != "" and prec != "":
        return "*%s%s, %s.*" % (
            cloud[0].capitalize(), cloud[1:].lower(), prec.lower())
    else:
        if cloud != "":
            return "*%s%s.*" % (cloud[0].capitalize(), cloud[1:].lower())
        elif prec != "":
            return "*%s%s.*" % (prec[0].capitalize(), prec[1:].lower())

    return ""


def __add_prec_common(prec: str) -> str:
    if prec != "":
        return " *%s%s*.\n" % (prec[0].capitalize(), prec[1:].lower())

    return "\n"


def __add_temp(temp: str) -> str:
    if temp != "":
        return "?????????????????????? *%s??C*" % (temp.replace(",", "...", -1))

    return ""


def __add_temp_comm(temp: str) -> str:
    if temp != "":
        return ", *%sC*.\n" % (temp.replace(",", "...", -1))

    return ".\n"


def __add_wind_direction_and_speed(direction: str, speed: str) -> str:
    if direction != "" and speed != "":
        return "?????????? *%s*, *%s ??/??*. " % (direction, speed)

    return ""


def __add_wind_comm(wind: str) -> str:
    if wind != "":
        return "*%s*." % wind

    return ""


def __add_author(synoptic: pogoda_api.Synoptic) -> str:
    if synoptic.get_id() != 0:
        return "\n\n_?????????????? ????????????????(??): %s %s %s._" % (
            synoptic.get_position(),
            synoptic.get_first_name(), synoptic.get_last_name())
    return ""


def __erase_redundant_spaces(text: str) -> str:
    new_text = ""

    for i, _ in enumerate(text):
        if i == 0:
            new_text += text[i]
            continue
        else:
            if text[i] == " " and text[i-1] == " ":
                continue
        new_text += text[i]

    return new_text


def __prepare_date(date: str) -> str:
    date = dtime.convert_date_to_words(date)
    date = dtime.eng_month_to_rus(date)
    date = dtime.eng_dweek_to_rus(date)

    return date
