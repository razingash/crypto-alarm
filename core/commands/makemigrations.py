import asyncpg

from asyncio import run as asyncio_run
from colorama import Style, Fore
from sqlalchemy.ext.asyncio import create_async_engine

from core.config import PG_USER, PG_HOST, PG_PORT, PG_PASSWORD, PG_NAME, POSTGRE_URL
from core.models import Base


def command_makemigrations():
    asyncio_run(makemigrations())


async def makemigrations() -> None:
    try:
        conn = await asyncpg.connect(f"postgresql://{PG_USER}:{PG_PASSWORD}@{PG_HOST}:{PG_PORT}/postgres")
        databases = await conn.fetch("SELECT datname FROM pg_database")
        db_names = [db['datname'] for db in databases]

        if PG_NAME not in db_names:
            print(Fore.LIGHTYELLOW_EX + f"Database '{PG_NAME}' isn't detected and will be created")
            await conn.execute(f'CREATE DATABASE "{PG_NAME}" OWNER "{PG_USER}"')
            print(Fore.LIGHTBLACK_EX + "Database created successfully")
            await init_db()
        else:
            print(Fore.LIGHTWHITE_EX + "Database already exists, if you want to change models you need to recreate db,"
                                       " because alembic can't work with asynchronous postgress")
        await conn.close()
    except Exception as e:
        print(e)


async def init_db() -> None:
    engine_new = create_async_engine(POSTGRE_URL, echo=True)
    async with engine_new.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    await engine_new.dispose()
    print(Style.BRIGHT + Fore.GREEN + 'success')
