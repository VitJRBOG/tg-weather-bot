# coding: utf-8


import json
import requests


class Forecast(object):
    def __init__(self):
        self.__date = ""
        self.__night_cloud = ""
        self.__night_prec = ""
        self.__night_prec_comm = ""
        self.__night_prec_vision = False
        self.__night_wind_direction = ""
        self.__night_wind_speed = ""
        self.__night_wind_comm = ""
        self.__night_temp = ""
        self.__night_temp_comm = ""
        self.__night_common_comm = ""
        self.__day_cloud = ""
        self.__day_prec = ""
        self.__day_prec_comm = ""
        self.__day_prec_vision = False
        self.__day_wind_direction = ""
        self.__day_wind_speed = ""
        self.__day_wind_comm = ""
        self.__day_temp = ""
        self.__day_temp_comm = ""
        self.__day_common_comm = ""
        self.__author = 0

    def set_date(self, new_date: str):
        self.__date = new_date

    def set_night_cloud(self, new_night_cloud: str):
        self.__night_cloud = new_night_cloud

    def set_night_prec(self, new_night_prec: str):
        self.__night_prec = new_night_prec

    def set_night_prec_comm(self, new_night_prec_comm: str):
        self.__night_prec_comm = new_night_prec_comm

    def set_night_prec_vision(self, new_night_prec_vision: bool):
        self.__night_prec_vision = new_night_prec_vision

    def set_night_wind_direction(self, new_night_wind_direction: str):
        self.__night_wind_direction = new_night_wind_direction

    def set_night_wind_speed(self, new_night_wind_speed: str):
        self.__night_wind_speed = new_night_wind_speed

    def set_night_wind_comm(self, new_night_wind_comm: str):
        self.__night_wind_comm = new_night_wind_comm

    def set_night_temp(self, new_night_temp: str):
        self.__night_temp = new_night_temp

    def set_night_temp_comm(self, new_night_temp_comm: str):
        self.__night_temp_comm = new_night_temp_comm

    def set_night_common_comm(self, new_night_common_comm: str):
        self.__night_common_comm = new_night_common_comm

    def set_day_cloud(self, new_day_cloud: str):
        self.__day_cloud = new_day_cloud

    def set_day_prec(self, new_day_prec: str):
        self.__day_prec = new_day_prec

    def set_day_prec_comm(self, new_day_prec_comm: str):
        self.__day_prec_comm = new_day_prec_comm

    def set_day_prec_vision(self, new_day_prec_vision: bool):
        self.__day_prec_vision = new_day_prec_vision

    def set_day_wind_direction(self, new_day_wind_direction: str):
        self.__day_wind_direction = new_day_wind_direction

    def set_day_wind_speed(self, new_day_wind_speed: str):
        self.__day_wind_speed = new_day_wind_speed

    def set_day_wind_comm(self, new_day_wind_comm: str):
        self.__day_wind_comm = new_day_wind_comm

    def set_day_temp(self, new_day_temp: str):
        self.__day_temp = new_day_temp

    def set_day_temp_comm(self, new_day_temp_comm: str):
        self.__day_temp_comm = new_day_temp_comm

    def set_day_common_comm(self, new_day_common_comm: str):
        self.__day_common_comm = new_day_common_comm

    def set_author(self, new_author: int):
        self.__author = new_author

    def get_date(self):
        return self.__date

    def get_night_cloud(self):
        return self.__night_cloud

    def get_night_prec(self):
        return self.__night_prec

    def get_night_prec_comm(self):
        return self.__night_prec_comm

    def get_night_prec_vision(self):
        return self.__night_prec_vision

    def get_night_wind_direction(self):
        return self.__night_wind_direction

    def get_night_wind_speed(self):
        return self.__night_wind_speed

    def get_night_wind_comm(self):
        return self.__night_wind_comm

    def get_night_temp(self):
        return self.__night_temp

    def get_night_temp_comm(self):
        return self.__night_temp_comm

    def get_night_common_comm(self):
        return self.__night_common_comm

    def get_day_cloud(self):
        return self.__day_cloud

    def get_day_prec(self):
        return self.__day_prec

    def get_day_prec_comm(self):
        return self.__day_prec_comm

    def get_day_prec_vision(self):
        return self.__day_prec_vision

    def get_day_wind_direction(self):
        return self.__day_wind_direction

    def get_day_wind_speed(self):
        return self.__day_wind_speed

    def get_day_wind_comm(self):
        return self.__day_wind_comm

    def get_day_temp(self):
        return self.__day_temp

    def get_day_temp_comm(self):
        return self.__day_temp_comm

    def get_day_common_comm(self):
        return self.__day_common_comm

    def get_author(self):
        return self.__author


def get_forecast(pogoda_api_url: str, region: str, location: int, date: str):
    values = {
        "region": region,
        "date": date
    }

    u = __make_url(pogoda_api_url, "GetForecast", values)
    response, ok = __send_request(u)
    if ok and response != "{}":
        return __parse_forecast(json.loads(response), location)
    return None


def __parse_forecast(raw_data: dict, location: int):
    d = raw_data[str(location)]

    forecast = Forecast()
    forecast.set_date(d["date"])
    forecast.set_night_cloud(d["nightcloud"])
    forecast.set_night_prec(d["nightprec"])
    forecast.set_night_prec_comm(d["nightpreccomm"])
    forecast.set_night_prec_vision(d["nightprecvision"])
    forecast.set_night_wind_direction(d["nightwinddirrect"])
    forecast.set_night_wind_speed(d["nightwindspeed"])
    forecast.set_night_wind_comm(d["nightwindcomm"])
    forecast.set_night_temp(d["nighttemp"])
    forecast.set_night_temp_comm(d["nighttempcomm"])
    forecast.set_night_common_comm(d["nightcommoncomm"])
    forecast.set_day_cloud(d["daycloud"])
    forecast.set_day_prec(d["dayprec"])
    forecast.set_day_prec_comm(d["daypreccomm"])
    forecast.set_day_prec_vision(d["dayprecvision"])
    forecast.set_day_wind_direction(d["daywinddirrect"])
    forecast.set_day_wind_speed(d["daywindspeed"])
    forecast.set_day_wind_comm(d["daywindcomm"])
    forecast.set_day_temp(d["daytemp"])
    forecast.set_day_temp_comm(d["daytempcomm"])
    forecast.set_day_common_comm(d["daycommoncomm"])
    forecast.set_author(d["author"])

    return forecast


class Synoptic(object):
    def __init__(self):
        self.__id = 0
        self.__first_name = "Name"
        self.__last_name = "LastName"
        self.__position = "Position"
        self.__region = 0

    def set_id(self, new_id: int):
        self.__id = new_id

    def set_first_name(self, new_first_name: str):
        self.__first_name = new_first_name

    def set_last_name(self, new_last_name: str):
        self.__last_name = new_last_name

    def set_position(self, new_position: str):
        self.__position = new_position

    def set_region(self, new_region: int):
        self.__region = new_region

    def get_id(self):
        return self.__id

    def get_first_name(self):
        return self.__first_name

    def get_last_name(self):
        return self.__last_name

    def get_position(self):
        return self.__position

    def get_region(self):
        return self.__region


def get_synoptic_data(pogoda_api_url: str, region: str, synoptic_id: str):
    values = {
        "region": region
    }

    u = __make_url(pogoda_api_url, "GetSynopticList", values)
    response, ok = __send_request(u)
    if ok and response != "{}":
        return __parse_synoptic_data(json.loads(response), synoptic_id)
    return None


def __parse_synoptic_data(raw_data: dict, synoptic_id: str):
    synoptic = Synoptic()

    d = raw_data.get(synoptic_id)
    if d is not None:
        synoptic.set_id(raw_data[synoptic_id]["id"])
        synoptic.set_first_name(raw_data[synoptic_id]["first_name"])
        synoptic.set_last_name(raw_data[synoptic_id]["last_name"])
        synoptic.set_position(raw_data[synoptic_id]["position"])
        synoptic.set_region(raw_data[synoptic_id]["region"])

        return synoptic

    return None


def __make_url(pogoda_api_url: str, method: str, values: dict):
    u = "%sforecast/api/%s" % (pogoda_api_url, method)
    for i, v in enumerate(values.items()):
        if i == 0:
            u += "?"
        else:
            u += "&"
        u += "%s=%s" % (v[0], v[1])

    return u


def __send_request(u: str):
    foo = requests.get(u)
    if foo.status_code == 200:
        return foo.text, True
    else:
        print("Error: " + foo.text)
        return "", False
