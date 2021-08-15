# coding: utf-8


import tools.dtime as dtime
import tg_api.tg_api as tg_api
import pogoda_api.pogoda_api as pogoda_api


def combine_msg_values(msg_data: dict, forecast: pogoda_api.Forecast,
                       synoptic: pogoda_api.Synoptic):
    v = tg_api.MessageRequestParams()
    v.set_chat_id(msg_data["from"]["id"])
    v.set_text(
        __compose_msg_text(forecast, __prepare_date(forecast.get_date()),
                           synoptic))

    return v


def __compose_msg_text(forecast: pogoda_api.Forecast, date: str,
                       synoptic: pogoda_api.Synoptic):
    night_forecast = ""
    night_forecast += __add_cloud_and_prec(forecast.get_night_cloud(),
                                           forecast.get_night_prec())
    night_forecast += __add_prec_common(forecast.get_night_prec_comm())
    night_forecast += __add_temp(forecast.get_night_temp())
    night_forecast += __add_temp_comm(forecast.get_night_temp_comm())
    night_forecast += __add_wind_direction_and_speed(
        forecast.get_night_wind_direction(), forecast.get_night_wind_speed())
    night_forecast += __add_wind_comm(forecast.get_night_wind_comm())

    day_forecast = ""
    day_forecast += __add_cloud_and_prec(forecast.get_day_cloud(),
                                         forecast.get_day_prec())
    day_forecast += __add_prec_common(forecast.get_day_prec())
    day_forecast += __add_temp(forecast.get_day_temp())
    day_forecast += __add_temp_comm(forecast.get_day_temp_comm())
    day_forecast += __add_wind_direction_and_speed(
        forecast.get_day_wind_direction(), forecast.get_day_wind_speed())
    day_forecast += __add_wind_comm(forecast.get_day_wind_comm())

    author = __add_author(synoptic)

    forecast_text = "_НОЧЬ на %s_\n%s\n\n_ДЕНЬ_\n%s%s" % (
        date, night_forecast, day_forecast, author)

    return forecast_text


def __add_cloud_and_prec(cloud: str, prec: str):
    if cloud != "" and prec != "":
        return "*%s%s, %s.*" % (
            cloud[0].capitalize(), cloud[1:].lower(), prec.lower())
    else:
        if cloud != "":
            return "*%s%s.*" % (cloud[0].capitalize(), cloud[1:].lower())
        elif prec != "":
            return "*%s%s.*" % (prec[0].capitalize(), prec[1:].lower())

    return ""


def __add_prec_common(prec: str):
    if prec != "":
        return " *%s%s*.\n" % (prec[0].capitalize(), prec[1:].lower())

    return "\n"


def __add_temp(temp):
    if temp != "":
        return "Температура *%s˚C*" % (temp.replace(",", "...", -1))

    return ""


def __add_temp_comm(temp: str):
    if temp != "":
        return ", *%sC*.\n" % (temp.replace(",", "...", -1))

    return ".\n"


def __add_wind_direction_and_speed(direction: str, speed: str):
    if direction != "" and speed != "":
        return "Ветер *%s*, *%s м/с*. " % (direction, speed)

    return ""


def __add_wind_comm(wind: str):
    if wind != "":
        return "*%s*." % wind

    return ""


def __add_author(synoptic: pogoda_api.Synoptic):
    if synoptic is not None:
        return "\n\n_%s%s %s %s_" % (
            synoptic.get_position()[0].capitalize(),
            synoptic.get_position()[1:],
            synoptic.get_first_name(), synoptic.get_last_name())
    return ""


def __prepare_date(date: str):
    date = dtime.convert_date_to_words(date)
    date = dtime.eng_month_to_rus(date)
    date = dtime.eng_dweek_to_rus(date)

    return date
