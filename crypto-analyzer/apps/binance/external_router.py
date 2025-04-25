import httpx
from httpx import HTTPStatusError, TimeoutException, ConnectError, RequestError

from core.controller import BinanceAPIController
from core.endpoints import endpoints
from core.logger import custom_logger


class BinanceAPI:
    """апи для получения триггеров и их запуск"""
    BASE_URL = "https://api.binance.com/api"

    def __init__(self, controller: BinanceAPIController):
        self.client = httpx.AsyncClient(timeout=10)
        self.controller = controller

    @staticmethod
    def check_and_update_endpoint_weights(response, endpoint, current_weight):
        """
        Method for adjusting the weights of endpoints. It isn't periodic task since the Binance weights are updated too
        often, it makes no sense to make a periodic renewal once every few minutes, given the low cost of the operation
        """
        used_weight = response.headers.get("x-mbx-used-weight-1m", "0") # x-mbx-used-weight

        try:
            used_weight = int(used_weight)
        except ValueError:
            used_weight = None

        api_weight = used_weight - current_weight
        expected_weight = endpoints[endpoint]

        if api_weight != expected_weight:
            expected_weight = api_weight
            custom_logger.log_with_path(
                level=3,
                msg=f"Endpoint updated {endpoint}: {expected_weight} -> {api_weight}",
                filename="endpoints.log"
            )
        return used_weight

    async def get(self, endpoint: str, endpoint_weight: int, response_model=None, params: dict = None):
        async def request():
            url = f"{self.BASE_URL}{endpoint}"
            try:
                response = await self.client.get(url, params=params)
                response.raise_for_status()

                used_weight = self.check_and_update_endpoint_weights(
                    response=response, endpoint=endpoint, current_weight=self.controller.current_weight
                )
                self.controller.current_weight = used_weight

                data = response.json()

                print(f'Current loading: {self.controller.current_weight}')
                if response_model:
                    return response_model(**data)
                return data
            except HTTPStatusError as e: # обработать, как правда пока не ясно, по идее просто стоит увеличить задержку
                print(f"BINANCE ERROR: херовый код ответа - {e.response.status_code} on {url}")
                raise
            except TimeoutException: # сразу оффать систему
                custom_logger.log_with_path(
                    level=2,
                    msg=f"TimeoutException: Request to {url}",
                    filename="BinanceErrors.log"
                )
                return True
            except ConnectError: # сразу оффать систему - возникает из-за сбоя сервера
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

        if isinstance(request, bool): # если ошибка то будет булево значение
            return request
        return await self.controller.request_with_limit(endpoint_weight, request)

    async def get_apiv3_accessibility(self):
        """проверяет доступен ли api/v3"""
        return await self.get(
            endpoint='/v3/ping',
            endpoint_weight=endpoints.get("/v3/ping")
        )

    # ситуативный апи, думаю лучше всего его не использовать(для такого лучше будет вебсокет, даже в доке так пишут)
    async def get_ticker_current_price(self, symbol=None):
        """цена конкретной валюты"""
        return await self.get(
            endpoint="/v3/ticker/price",
            endpoint_weight=endpoints.get("/v3/ticker/price")
        )

    async def get_price_change_24h(self, symbol=None):
        # если без атрибутов то стоимость будет 80, но данных будет намного больше, проверить как часто обновляются данные
        """изменение процентного значения за 24 часа"""
        return await self.get(
            endpoint="/v3/ticker/24hr",
            endpoint_weight=endpoints.get("/v3/ticker/24hr")
        )

# апи для торговых пар(триггеры для обмена конкретных криптовалют - на этих скачках можно также заработать)

#https://api.binance.com/api/v3/exchangeInfo?symbol=BTCUSDT
#https://api.binance.com/api/v3/klines?symbol=BTCUSDT&interval=1m
