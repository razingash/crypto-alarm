import logging
import os

from core.config import BASE_DIR


LOG_DIR = os.path.join(BASE_DIR, 'logs')
log_file_path = os.path.join(LOG_DIR, 'logs.log')

custom_logger = logging.getLogger('custom_logger')
custom_logger.setLevel(logging.DEBUG)
handler = logging.FileHandler(log_file_path, encoding='utf-8')
handler.setLevel(logging.DEBUG)
formatter = logging.Formatter('%(levelname)s (%(asctime)s): %(message)s [%(filename)s]', datefmt='%d/%m/%Y %H:%M:%S')
handler.setFormatter(formatter)
custom_logger.addHandler(handler)
