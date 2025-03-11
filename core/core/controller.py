import asyncio
import time

from core.logger import custom_logger


# проверить логи, возможно предел будет пробиватся, если нет то повысить до 6000

class BinanceAPIController:
    """
    данный контроллер нужен чтобы не пробить весовой лимит BinanceApi и не получить автобан
    # позже добавить учет вебсокет соединений | лимит - 300 подключений за попытку каждые 5 минут
    """
    def __init__(self, max_weight: int = 5700):
        """возможно слишком трудно будет это развивать дальше"""
        self.max_weight = max_weight # на самом деле лимит 6000, но лучше сделать поправку на асинхронность
        self.current_weight = 0
        self.last_reset_time = time.time()
        self.lock = asyncio.Lock()
        self.queue = asyncio.Queue()
        self.lock = asyncio.Lock()
        self.queue = asyncio.Queue()

    async def start(self):
        asyncio.create_task(self.reset_loop())
        asyncio.create_task(self.process_queue())

    async def reset_loop(self): # автосброс(скорее всего слишком затратный)
        while True:
            await asyncio.sleep(60)
            async with self.lock:
                self.current_weight = 0
                self.last_reset_time = time.time()

    async def process_queue(self):
        """Обрабатывает очередь запросов (если лимит был превышен)"""
        while True:
            request = await self.queue.get()
            await request()
            self.queue.task_done()

    async def request_with_limit(self, weight: int, request_func):
        """Управляет лимитом и выполняет запрос. Если лимит превышен, запрос ставится в очередь."""
        async with self.lock:
            if self.current_weight + weight > self.max_weight:
                custom_logger.log_with_path(
                    level=1,
                    msg=f"reached the limit for Binance API. Current weight:  {self.current_weight}",
                    path="ApiLimits.log"
                )
                await self.queue.put(lambda: request_func())
                return None

            self.current_weight += weight

        return await request_func()

controller = BinanceAPIController()
