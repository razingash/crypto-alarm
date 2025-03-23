from datetime import datetime

from sqlalchemy import String, DateTime, Boolean, ForeignKey, Integer
from sqlalchemy.orm import mapped_column, Mapped, relationship

from core.models.base import Base

__all__ = ["CryptoApi", "CryptoParams"]

"""
crypto-analizer:
    1) система актуализации (в FormulaLoader)
        - работает путем периодической проверки в FormulaLoader(но скорее всего это можно сделать лучше)
    2) контроль над ними с помощью хуков
crypto-gateway:
    1) получение доступных переменных подвязанных под апи, для заполнения данных о актуальных переменных для клавиатуры
"""

class CryptoApi(Base):
    """список доступных апи"""
    api: Mapped[str] = mapped_column(String(500), nullable=False, unique=True, index=True)
    is_actual: Mapped[bool] = mapped_column(Boolean, nullable=False, default=True)
    last_updated: Mapped[bool] = mapped_column(DateTime, nullable=False, default=datetime.utcnow)

    params = relationship("CryptoParams", back_populates="crypto_api", cascade="all, delete-orphan")

    __tablename__ = "crypto_api"


class CryptoParams(Base):
    """
    доступные на данный момент параметры апи, сделать чтобы они обновлялись по триггерам
    """
    parameter: Mapped[str] = mapped_column(String(500), nullable=False)
    is_active: Mapped[bool] = mapped_column(Boolean, nullable=False, default=True)
    last_updated: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow)
    excluded_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow)

    crypto_api_id: Mapped[int] = mapped_column(Integer, ForeignKey("crypto_api.id", ondelete="CASCADE"), nullable=False)
    crypto_api = relationship("CryptoApi", back_populates="params")

    __tablename__ = "crypto_params"


"""
MongoDB
class CryptoHistory(Base):
    symbol = mapped_column(String(10), nullable=False, index=True)
    price = mapped_column(Float, nullable=False)
    timestamp = mapped_column(DateTime, default=datetime.utcnow)

    __tablename__ = "crypto_history"

    def __repr__(self):
        return f"<CryptoHistory {self.symbol} - {self.price} at {self.timestamp}>"
"""
