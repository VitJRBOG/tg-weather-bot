# coding: utf-8
import tools.tools as tools
import bot.core as bot


pogoda_api_url, ok = tools.get_pogoda_api_url()
if ok:
    bot_conf, ok = tools.get_bot_conf()
    if ok:
        bot.start(bot_conf, pogoda_api_url)
