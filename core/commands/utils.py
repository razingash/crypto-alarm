import asyncpg
from core.config import PG_NAME, PG_USER, PG_PASSWORD, PG_HOST, PG_PORT


async def is_database_exists():
    try:
        conn = await asyncpg.connect(f"postgresql://{PG_USER}:{PG_PASSWORD}@{PG_HOST}:{PG_PORT}/postgres")
        databases = await conn.fetch("SELECT datname FROM pg_database")
        db_names = [db['datname'] for db in databases]

        if PG_NAME not in db_names:
            print("Database is not detected and will be created")
            await conn.execute(f'CREATE DATABASE "{PG_NAME}" OWNER "{PG_USER}"')
            print("Database created")

        await conn.close()
    except Exception as e:
        print(e)
        return False
    else:
        return True
