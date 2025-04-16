import asyncpg

from asyncio import run as asyncio_run
from colorama import Style, Fore
from sqlalchemy.ext.asyncio import create_async_engine

from core.config import PG_USER, PG_HOST, PG_PORT, PG_PASSWORD, PG_NAME, POSTGRE_URL
from core.models import Base


def command_makemigrations():
    return asyncio_run(makemigrations())


async def makemigrations():
    try:
        conn = await asyncpg.connect(f"postgresql://{PG_USER}:{PG_PASSWORD}@{PG_HOST}:{PG_PORT}/postgres")

        db_exists = await conn.fetchval("""
            SELECT 1 FROM pg_database WHERE datname = $1
        """, PG_NAME)

        if not db_exists:
            print(Fore.LIGHTYELLOW_EX + f"Database '{PG_NAME}' isn't detected and will be created")
            await conn.execute(f'CREATE DATABASE "{PG_NAME}" OWNER "{PG_USER}"')
            print(Fore.LIGHTBLACK_EX + "Database created successfully")
            await conn.close()

            await init_db()
        else:
            await conn.close()

            conn_db = await asyncpg.connect(f"postgresql://{PG_USER}:{PG_PASSWORD}@{PG_HOST}:{PG_PORT}/{PG_NAME}")
            table_exists = await conn_db.fetchval("""
                SELECT 1 FROM information_schema.tables 
                WHERE table_schema = 'public' AND table_name = 'user_user'
            """)

            if not table_exists:
                print(Fore.LIGHTYELLOW_EX + "Database exists but appears empty. Running initialization.")
                await conn_db.close()
                await init_db()
            else:
                print(Fore.LIGHTWHITE_EX + "Database already exists and is already initialized.")
                await conn_db.close()
                return True
    except Exception as e:
        print(Style.BRIGHT + Fore.RED + f"Error: {str(e)}")
        return True


async def init_db() -> None:
    engine_new = create_async_engine(POSTGRE_URL, echo=True)
    async with engine_new.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    await engine_new.dispose()
    print(Style.BRIGHT + Fore.GREEN + 'Database created successfully')
