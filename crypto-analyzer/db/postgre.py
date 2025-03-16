from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession, async_sessionmaker

from core.config import POSTGRE_URL
from db.base import AbstractRepository


class PostgresDatabase(AbstractRepository):
    def __init__(self):
        self.engine = create_async_engine(POSTGRE_URL, echo=True)
        self.session_factory = async_sessionmaker(
            bind=self.engine,
            autoflush=False,
            expire_on_commit=False
        )

    def get_scoped_session(self) -> AsyncSession:
        return async_sessionmaker(bind=self.engine)()

    async def session_dependency(self):
        session = self.get_scoped_session()
        try:
            yield session
        finally:
            await session.close()

postgres_db = PostgresDatabase()
