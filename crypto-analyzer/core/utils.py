import asyncio
from typing import Callable

def run_sync_to_async(func: Callable, *args, **kwargs):
    """нужна чтобы запускать асинхронные функции конкретно в данном классе"""
    try:
        loop = asyncio.get_running_loop()
    except RuntimeError:
        return asyncio.run(func(*args, **kwargs))
    else:
        future = asyncio.run_coroutine_threadsafe(func(*args, **kwargs), loop)
        return future.result()
