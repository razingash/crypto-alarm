import httpx
from httpx import HTTPStatusError, TimeoutException, ConnectError, RequestError

from core.controller import BinanceAPIController
from core.endpoints import endpoints
from core.logger import custom_logger
from core.middlewares import WeightTrackingMiddleware


class BinanceAPI:
    """апи для получения триггеров и их запуск"""
    BASE_URL = "https://api.binance.com/api"

    def __init__(self, controller: BinanceAPIController, middleware: WeightTrackingMiddleware):
        self.client = httpx.AsyncClient(timeout=10)
        self.controller = controller
        self.middleware = middleware

    async def get(self, endpoint: str, weight: int, response_model=None, params: dict = None):
        async def request():
            url = f"{self.BASE_URL}{endpoint}"
            try:
                response = await self.client.get(url, params=params)
                response.raise_for_status()
                data = response.json()

                if response_model:
                    return response_model(**data)
                return data
            except HTTPStatusError as e: # обработать
                print(f"BINANCE ERROR: херовый код ответа - {e.response.status_code} on {url}")
                raise
            except TimeoutException: # сразу оффать систему
                print(f"BINANCE ERROR Request to {url} timed out.")
                custom_logger.log_with_path(
                    level=2,
                    msg=f"TimeoutException: Request to {url}",
                    filename="BinanceErrors.log"
                )
                return True
            except ConnectError: # сразу оффать систему - возникает из-за сбоя сервера
                print(f"BINANCE ERROR Could not connect to {url}.")
                custom_logger.log_with_path(
                    level=1,
                    msg=f"ConnectError: Could not connect to {url}.",
                    filename="BinanceErrors.log"
                )
                return True
            except RequestError as e: # api перестал быть действительным - нужно исключить его из эндпоинтов и тд
                print(f"An error occurred while requesting {url}: {str(e)}")
                custom_logger.log_with_path(
                    level=1,
                    msg=f"RequestError: An error occurred while requesting {url}: {str(e)}",
                    filename="BinanceErrors.log"
                )
                raise
            except Exception as e: # пока так оставить чтобы валилась вся система
                print(f"[UNKNOWN ERROR] {str(e)} while calling {url}")
                raise

        print(f'Current loading: {self.controller.current_weight}')
        if isinstance(request, bool): # если ошибка то будет булево значение
            return request
        return await self.controller.request_with_limit(weight, request)

    async def check_and_update_weights(self): # можно сделать из этой функции команду, и использовать после инициализации
        """Функция для постепенной корректировки весов"""
        print("Запуск актуализации весов через мидлвейр...")
        self.middleware.enable_update_mode()

    async def get_apiv3_accessibility(self):
        """проверяет доступен ли api/v3"""
        return await self.get(
            endpoint='/v3/ping',
            weight=endpoints.get("/v3/ping")
        )

    # ситуативный апи, думаю лучше всего его не использовать(для такого лучше будет вебсокет, даже в доке так пишут)
    async def get_ticker_current_price(self, symbol=None):
        """цена конкретной валюты"""
        await self.get(
            endpoint="/v3/ticker/price",
            weight=endpoints.get("/v3/ticker/price")
        )

    async def get_price_change_24h(self, symbol=None):
        # если без атрибутов то стоимость будет 80, но данных будет намного больше, проверить как часто обновляются данные
        """изменение процентного значения за 24 часа"""
        await self.get(
            endpoint="/v3/ticker/24hr",
            weight=endpoints.get("/v3/ticker/24hr")
        )

# апи для торговых пар(триггеры для обмена конкретных криптовалют - на этих скачках можно также заработать)

#https://api.binance.com/api/v3/exchangeInfo?symbol=BTCUSDT
#https://api.binance.com/api/v3/klines?symbol=BTCUSDT&interval=1m
