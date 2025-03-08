import alembic.util.exc
from asyncio import run as asyncio_run
from alembic import command
from alembic.config import Config
from colorama import Style, Fore
from sqlalchemy import inspect
from sqlalchemy.ext.asyncio import create_async_engine

from commands.utils import is_database_exists
from core.config import POSTGRE_URL
from core.models import Base


def command_makemigrations():
    asyncio_run(async_makemigrations())

async def async_makemigrations():
    await is_database_exists()

    try:
        alembic_cfg = Config("alembic.ini")
        command.upgrade(alembic_cfg, "head")
        command.revision(alembic_cfg, message=None, autogenerate=True)
        await create_tables()
    except alembic.util.exc.CommandError:
        print('old migrations were found')
    else:
        print(Style.BRIGHT + Fore.GREEN + 'success')


async def create_tables():
    engine = create_async_engine(POSTGRE_URL, echo=True)

    async with engine.connect() as conn:
        existing_tables = await conn.run_sync(lambda sync_conn: inspect(sync_conn).get_table_names())

        if not existing_tables:
            await conn.run_sync(Base.metadata.create_all)
