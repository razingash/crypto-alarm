from os import getenv
from pathlib import Path

from dotenv import load_dotenv, find_dotenv


load_dotenv(find_dotenv())

BASE_DIR = Path(__file__).resolve().parent.parent

ALEMBIC_INI_PATH = f'{BASE_DIR}/alembic.ini'

# postgres
POSTGRE_URL = "postgresql+asyncpg://user:password@localhost/db_name"
POSTGRE_NAME = getenv('DB_NAME')
POSTGRE_USER = getenv('DB_USER')
POSTGRE_PASSWORD = getenv('DB_PASSWORD')
POSTGRE_HOST = getenv('DB_HOST')
POSTGRE_PORT = getenv('DB_PORT')

# mongoDB
