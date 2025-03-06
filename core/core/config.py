from os import getenv
from pathlib import Path

from dotenv import load_dotenv, find_dotenv


load_dotenv(find_dotenv())

BASE_DIR = Path(__file__).resolve().parent.parent

ALEMBIC_INI_PATH = f'{BASE_DIR}/alembic.ini'

# SECRET_KEY = getenv('SECRET_KEY')

# postgres
PG_NAME = getenv('DB_NAME')
PG_USER = getenv('DB_USER')
PG_PASSWORD = getenv('DB_PASSWORD')
PG_HOST = getenv('DB_HOST')
PG_PORT = getenv('DB_PORT')
POSTGRE_URL = f"postgresql+asyncpg://{PG_USER}:{PG_PASSWORD}@{PG_HOST}:{PG_PORT}/{PG_NAME}"

# mongoDB
