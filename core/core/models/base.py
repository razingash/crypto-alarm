from sqlalchemy import Integer, String, Boolean, ForeignKey
from sqlalchemy.orm import DeclarativeBase, mapped_column, Mapped


class Base(DeclarativeBase):
    id: Mapped[str] = mapped_column(Integer, primary_key=True, autoincrement=True, index=True)

    __abstract__ = True


class BaseTrigger(Base):
    """базовый тип триггера который зависит от символа"""
    owner_id: Mapped[int] = mapped_column(Integer, ForeignKey("user_user.id"), nullable=False)
    symbol: Mapped[str] = mapped_column(String(10), nullable=False)
    is_notified: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False) # тут могут быть баги при большой нагрузке

    __abstract__ = True
