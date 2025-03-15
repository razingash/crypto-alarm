import asyncio

from apps.binance.external_router import BinanceAPI


class BinanceAPIOrchestrator:
    """
    запускает периодические задачи
    на первый взгляд выглядит не плохо, вот только большое количество тасок он сейчас не вывезет
    """
    def __init__(self, binance_api: BinanceAPI):
        self.binance_api = binance_api
        self.tasks = []

    async def start(self):
        """Запуск фоновых задач."""
        self.tasks.extend([
            asyncio.create_task(self.update_weights_loop()),
            asyncio.create_task(self.check_api_status_loop())
        ])

    async def update_weights_loop(self):
        """Периодическая актуализация весов."""
        while True:
            await asyncio.sleep(3600)  # Раз в час
            await self.binance_api.check_and_update_weights()

    async def check_api_status_loop(self):
        """Проверка доступности API."""
        while True:
            await asyncio.sleep(60)
            await self.binance_api.get_apiv3_accessibility()
