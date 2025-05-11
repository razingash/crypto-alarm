import asyncio
import time

from core.logger import custom_logger


# проверить логи, возможно предел будет пробиватся, если нет то повысить до 6000

class BinanceAPIController:
    """
    данный контроллер следит за тем чтобы не пробить весовой лимит Binance.com и не получить автобан
    Также он отвечает за выполнение запросов из класса BinanceApi
    # позже добавить учет вебсокет соединений | лимит - 300 подключений за попытку каждые 5 минут
    """
    def __init__(self, max_weight: int = 5700):
        """возможно слишком трудно будет это развивать дальше"""
        self.max_weight = max_weight # на самом деле лимит 6000, но лучше сделать поправку на асинхронность
        self.current_weight = 0
        self.last_reset_time = time.time()
        self.pending_requests = set()

        self.lock = asyncio.Lock()
        self.queue = asyncio.Queue()
        self.queue_event = None

    async def start(self, queue_event):
        self.queue_event = queue_event
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
            await self.queue_event.wait()

            async with self.lock:
                if self.queue.empty():
                    self.queue_event.clear()
                    continue

                request_func, weight = await self.queue.get()
                self.pending_requests.discard(request_func)

            await request_func()

            async with self.lock:
                self.current_weight += weight
                self.queue.task_done()

    async def request_with_limit(self, endpoint_weight: int, request_func):
        """Управляет лимитом и выполняет запрос. Если лимит превышен, запрос ставится в очередь."""
        async with self.lock:
            if self.current_weight + endpoint_weight > self.max_weight:
                custom_logger.log_with_path(
                    level=1,
                    msg=f"reached the limit for Binance API. Current weight:  {self.current_weight}",
                    filename="ApiLimits.log"
                )
                await self.queue.put((request_func, endpoint_weight))
                self.queue_event.set()
                return None

        return await request_func()

controller = BinanceAPIController()
