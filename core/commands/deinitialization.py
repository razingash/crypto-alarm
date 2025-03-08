import asyncpg
from asyncio import run as asyncio_run

from colorama import Fore

from core.config import PG_USER, PG_PASSWORD, PG_HOST, PG_PORT, PG_NAME


def command_deinitialization():
    """delete database and migrations"""
    asyncio_run(deinitialization())


async def deinitialization() -> None:
    try:
        conn = await asyncpg.connect(f"postgresql://{PG_USER}:{PG_PASSWORD}@{PG_HOST}:{PG_PORT}/postgres")

        await conn.execute(f"""
                SELECT pg_terminate_backend(pid)
                FROM pg_stat_activity
                WHERE datname = '{PG_NAME}' AND pid <> pg_backend_pid();
            """)

        await conn.execute(f'DROP DATABASE IF EXISTS "{PG_NAME}";')
        print(Fore.GREEN + f"Database '{PG_NAME}' sucessfully deleteted.")

        await conn.close()
    except Exception as e:
        print(Fore.LIGHTRED_EX + f"Error during deleting a database: {e}")
