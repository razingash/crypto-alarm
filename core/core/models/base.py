from sqlalchemy import Integer
from sqlalchemy.orm import DeclarativeBase, mapped_column, Mapped


class Base(DeclarativeBase):
    id: Mapped[str] = mapped_column(Integer, primary_key=True, autoincrement=True, index=True)

    __abstract__ = True
