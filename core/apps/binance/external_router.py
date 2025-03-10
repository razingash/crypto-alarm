import httpx
from fastapi import APIRouter

from apps.binance.schemas import Ticker24hrResponse, TickerCurrentPriceResponse
from core.controller import BinanceAPIController

router = APIRouter()


# дефолтные апи для получения первичных триггеров

class BinanceAPI: # сильный минус - трудно реализовать автокоррекцию весов, чтобы не было сильной нагрузки
    BASE_URL = "https://api.binance.com/api"

    def __init__(self, controller: BinanceAPIController):
        self.client = httpx.AsyncClient()
        self.controller = controller

    async def get(self, endpoint: str, weight: int, response_model=None, params: dict = None):
        async def request():
            url = f"{self.BASE_URL}{endpoint}"
            response = await self.client.get(url, params=params)
            response.raise_for_status()
            data = response.json()

            if response_model:
                return response_model(**data)

            return data

        return await self.controller.request_with_limit(weight, request)

    async def get_apiv3_accessibility(self):
        """x-mbx-used-weight: weigth=4 | проверяет доступен ли api/v3"""
        await self.get('/v3/ping', weight=4)

    # ситуативный апи, думаю лучше всего его не использовать(для такого лучше будет вебсокет, даже в доке так пишут)
    async def get_ticker_current_price(self, symbol):
        """x-mbx-used-weight: weigth=4 | цена конкретной валюты"""
        await self.get(endpoint="/v3/ticker/price", weight=4, response_model=TickerCurrentPriceResponse, params={"symbol": symbol})

    async def get_price_change_24h(self, symbol=None):
        # если без атрибутов то стоимость будет 80, но данных будет намного больше, проверить как часто обновляются данные
        """x-mbx-used-weight: weigth=8 |  изменение процентного значения за 24 часа"""
        await self.get(endpoint="/v3/ticker/24hr", weight=8, response_model=Ticker24hrResponse, params={"symbol": symbol})


# апи для торговых пар(триггеры для обмена конкретных криптовалют - на этих скачках можно также заработать)

#https://api.binance.com/api/v3/exchangeInfo?symbol=BTCUSDT
#https://api.binance.com/api/v3/klines?symbol=BTCUSDT&interval=1m
