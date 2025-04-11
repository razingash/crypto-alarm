import asyncio

from apps.binance.external_router import BinanceAPI
from core.logger import custom_logger

"""
добавить в систему которая будедт подтягивать необходимые данные относительно формул (начать с инициализации)
"""

class BinanceAPIOrchestrator:
    """
    запускает периодические задачи
    на первый взгляд выглядит не плохо, вот только большое количество тасок он сейчас не вывезет
    """
    def __init__(self, binance_api: BinanceAPI):
        self.binance_api = binance_api
        self.is_binance_online = True
        self.tasks = []

    async def start(self):
        """Запуск фоновых задач. первая задача должна быть проверка апи"""
        response = await self.binance_api.get_apiv3_accessibility()
        await self.check_binance_response(response)

        if self.is_binance_online:
            self.tasks.extend([
                asyncio.create_task(self.update_weights_loop()),
            ])

    async def check_binance_response(self, response):
        """checks Binacne availability"""
        if not isinstance(response, dict) or response:  # бинанс лег
            if self.is_binance_online is True:  # первое срабатывание - документация
                custom_logger.log_with_path(
                    level=3,
                    msg=f"Binance DOWN",
                    filename="BinanceAvailablity.log"
                )
                self.is_binance_online = False
            if len(self.tasks) > 1:
                for task in self.tasks:
                    task.cancel()
                self.tasks = [
                    asyncio.create_task(self.check_api_status_loop(60))
                ]
        elif self.is_binance_online is False:  # бинанс ожил
            custom_logger.log_with_path(
                level=3,
                msg=f"Binance UP",
                filename="BinanceAvailablity.log"
            )
            self.is_binance_online = True
            for task in self.tasks:
                task.cancel()

            #  нужно еще запустить нужные таски

    async def update_weights_loop(self): # улучшить чтобы были не все эндпоинты а используемые
        """Периодическая актуализация весов. не добавляет сложности, просто постепенно актуализирует весы"""
        while True:
            await asyncio.sleep(3600)
            response = await self.binance_api.check_and_update_weights()
            await self.check_binance_response(response)

    async def check_api_status_loop(self, cooldown: int): # возможно лучше сделать чтобы запрос срабатывал один раз при инициализации
        """Проверка доступности API."""
        while True:
            await asyncio.sleep(cooldown)
            response = await self.binance_api.get_apiv3_accessibility()
            await self.check_binance_response(response)
            if self.is_binance_online is True:
                break
