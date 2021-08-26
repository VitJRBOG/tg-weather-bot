# coding: utf-8


import datetime


def convert_date_to_words(date_str: str) -> str:
    d = datetime.datetime.strptime(date_str, "%Y-%m-%d")

    return d.strftime("%-d %B, %A")


def eng_month_to_rus(date: str) -> str:
    months = {
        "January": "января",
        "February": "февраля",
        "March": "марта",
        "April": "апреля",
        "May": "мая",
        "June": "июня",
        "July": "июля",
        "August": "августа",
        "September": "сентября",
        "October": "октября",
        "November": "ноября",
        "December": "декабря"
    }

    for k, v in months.items():
        if date.find(k) != -1:
            date = date.replace(k, v, 1)
            break

    return date


def eng_dweek_to_rus(date: str) -> str:
    days_of_the_week = {
        "Monday": "понедельник",
        "Tuesday": "вторник",
        "Wednesday": "среда",
        "Thursday": "четверг",
        "Friday": "пятница",
        "Saturday": "суббота",
        "Sunday": "воскресенье"
    }

    for k, v in days_of_the_week.items():
        if date.find(k) != -1:
            date = date.replace(k, v, 1)
            break

    return date
