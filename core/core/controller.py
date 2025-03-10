import asyncio
import time


class BinanceAPIController:
    """
    данный контроллер нужен чтобы не пробить весовой лимит BinanceApi и не получить автобан
    # позже добавить учет вебсокет соединений
    """
    def __init__(self, max_weight: int = 5900):
        self.max_weight = max_weight
        self.current_weight = 0
        self.last_reset_time = time.time()
        self.lock = asyncio.Lock()

    async def add_weight(self, weight: int):
        async with self.lock:
            self._reset_if_needed()

            if self.current_weight + weight > self.max_weight:  # Превышение лимита
                return False

            self.current_weight += weight
            return True

    def _reset_if_needed(self):
        if time.time() - self.last_reset_time > 60:
            self.current_weight = 0
            self.last_reset_time = time.time()
