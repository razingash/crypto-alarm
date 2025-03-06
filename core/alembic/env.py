import asyncio
from logging.config import fileConfig

from sqlalchemy import pool

from alembic import context
from sqlalchemy.ext.asyncio import create_async_engine

from core.config import POSTGRE_URL
from core.models.base import Base

config = context.config

if config.config_file_name is not None:
    fileConfig(config.config_file_name)

target_metadata = Base.metadata
config.set_main_option("sqlalchemy.url", POSTGRE_URL)

def run_migrations_offline() -> None:
    context.configure(
        url=POSTGRE_URL,
        target_metadata=target_metadata,
        literal_binds=True,
        dialect_opts={"paramstyle": "named"},
    )

    with context.begin_transaction():
        context.run_migrations()


async def run_migrations_online() -> None:
    connectable = create_async_engine(POSTGRE_URL, poolclass=pool.NullPool)

    async with connectable.connect() as connection:
        await connection.run_sync(
            lambda conn: context.configure(
                connection=conn, target_metadata=target_metadata, compare_type=True
            )
        )
        await connection.run_sync(
            lambda sync_connection: context.run_migrations()
        )


if context.is_offline_mode():
    run_migrations_offline()
else:
    asyncio.run(run_migrations_online())
