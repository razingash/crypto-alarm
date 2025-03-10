from fastapi import APIRouter

from apps.binance.schemas import Ticker24hrResponse, TickerCurrentPriceResponse

router = APIRouter()


# дефолтные апи для получения первичных триггеров

@router.get(path='https://api.binance.com/api/v3/ping')
async def get_apiv3_accessibility():
    """x-mbx-used-weight: weigth=4 | проверяет доступен ли api/v3"""
    pass

@router.get(path='https://api.binance.com/api/v3/ticker/price?symbol={symbol}', response_model=TickerCurrentPriceResponse)
async def get_ticker_current_price(): # ситуативный апи, думаю лучше всего его не использовать(для такого лучше будет вебсокет)
    """x-mbx-used-weight: weigth=4 | цена конкретной валюты"""
    pass

@router.get(path="https://api.binance.com/api/v3/ticker/24hr?symbol={symbol}", response_model=Ticker24hrResponse)
async def get_price_change_24h():
    """x-mbx-used-weight: weigth=8 |  изменение процентного значения за 24 часа"""
    pass


# апи для торговых пар(триггеры для обмена конкретных криптовалют - на этих скачках можно также заработать)

#https://api.binance.com/api/v3/exchangeInfo?symbol=BTCUSDT
#https://api.binance.com/api/v3/klines?symbol=BTCUSDT&interval=1m
