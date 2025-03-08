from asyncio import run as asyncio_run

from sqlalchemy.exc import ProgrammingError, OperationalError
from colorama import Fore, Style
from alembic.config import Config
from alembic.script import ScriptDirectory
from sqlalchemy import text

from db.postgre import postgres_db
from .makemigrations import command_makemigrations
from core.config import ALEMBIC_INI_PATH
from .utils import is_database_exists


def command_initialization():
    try:
        if not asyncio_run(is_database_exists()):
            initialize()
            return
        is_migrations_applied = asyncio_run(are_migrations_applied())
    except OperationalError: # database doesn't exist
        initialize()
    else:
        if not is_migrations_applied:
            initialize()


def initialize():
    command_makemigrations()


async def are_migrations_applied() -> bool:
    """checks the relevance of migrations"""
    try:
        async with postgres_db.session_factory() as session:
            result = await session.execute(text("SELECT version_num FROM alembic_version LIMIT 1"))
            applied_version = result.scalar()
    except ProgrammingError:
        return False

    alembic_cfg = Config(ALEMBIC_INI_PATH)
    script_dir = ScriptDirectory.from_config(alembic_cfg)
    try:
        latest_revision = script_dir.get_heads()[0]
    except IndexError:
        return False

    return applied_version == latest_revision
