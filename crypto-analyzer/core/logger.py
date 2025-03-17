import logging
import os

from core.config import BASE_DIR


class CustomLogger(logging.Logger):
    LOG_DIR = os.path.join(BASE_DIR, 'logs')
    log_file_path = os.path.join(LOG_DIR, 'logs.log')

    def __init__(self, name):
        super().__init__(name)

    def log_with_path(self, level: int, msg: str, path=None, filename=None, *args, **kwargs) -> None:
        """level: 1 - ERROR, 2 - WARNING, 3 - INFO, 4(else) - DEBUG"""
        if path:
            log_directory = os.path.relpath(os.path.join(BASE_DIR, path), BASE_DIR)
        else:
            log_directory = self.log_file_path

        log_file = os.path.join(log_directory, filename) if filename else self.log_file_path

        log_file = os.path.normpath(log_file)

        if level == 1:
            level = logging.ERROR
        elif level == 2:
            level = logging.WARNING
        elif level == 3:
            level = logging.INFO
        else:
            level = logging.DEBUG

        extra = kwargs.get("extra", {})
        extra["custom_path"] = log_file
        kwargs["extra"] = extra

        handler = logging.FileHandler(log_file, encoding="utf-8")
        handler.setLevel(logging.DEBUG)
        formatter = logging.Formatter('%(levelname)s (%(asctime)s): %(message)s [%(filename)s]',
                                      datefmt='%d/%m/%Y %H:%M:%S')
        handler.setFormatter(formatter)

        self.addHandler(handler)
        self.log(level, msg, *args, **kwargs)
        self.removeHandler(handler)

logging.setLoggerClass(CustomLogger)
custom_logger = CustomLogger("custom_logger")
