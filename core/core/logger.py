import logging
import os

from core.config import BASE_DIR


LOG_DIR = os.path.join(BASE_DIR, 'logs')
log_file_path = os.path.join(LOG_DIR, 'logs.log')


class CustomLogger(logging.Logger):
    def __init__(self, name):
        super().__init__(name)

    def log_with_path(self, level: int, msg: str, path=None, *args, **kwargs) -> None:
        """level: 1 - ERROR, 2 - WARNING, 3 - INFO, 4(else) - DEBUG"""
        if path:
            relative_path = os.path.relpath(os.path.join(BASE_DIR, path), BASE_DIR)
        else:
            relative_path = log_file_path

        if level == 1:
            level = logging.ERROR
        elif level == 2:
            level = logging.WARNING
        elif level == 3:
            level = logging.INFO
        else:
            level = logging.DEBUG

        extra = kwargs.get("extra", {})
        extra["custom_path"] = relative_path
        kwargs["extra"] = extra

        self.log(level, msg, *args, **kwargs)


logging.setLoggerClass(CustomLogger)
custom_logger = CustomLogger("custom_logger")

formatter = logging.Formatter('%(levelname)s (%(asctime)s): %(message)s [%(filename)s]', datefmt='%d/%m/%Y %H:%M:%S')

handler = logging.FileHandler(log_file_path, encoding="utf-8")
handler.setLevel(logging.DEBUG)
handler.setFormatter(formatter)

custom_logger.addHandler(handler)
