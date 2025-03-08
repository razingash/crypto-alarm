from datetime import datetime

from sqlalchemy import String, Float, DateTime
from sqlalchemy.orm import mapped_column

from core.models.base import Base

__all__ = ["CryptoPrice", "CryptoHistory"]

class CryptoPrice(Base):
    symbol = mapped_column(String(10), nullable=False, index=True, unique=True)
    name = mapped_column(String(50), nullable=False, unique=True)
    price = mapped_column(Float, nullable=False)
    market_cap = mapped_column(Float)  # Рыночная капитализация
    volume_24h = mapped_column(Float)  # Объём торгов | может быть спекулятивным параметром
    percent_change_1h = mapped_column(Float)
    percent_change_24h = mapped_column(Float)
    percent_change_7d = mapped_column(Float)
    last_updated = mapped_column(DateTime, default=datetime.utcnow)

    __tablename__ = "crypto_prices"

    def __repr__(self):
        return f"<CryptoPrice {self.symbol} - {self.price}>"


class CryptoHistory(Base):
    symbol = mapped_column(String(10), nullable=False, index=True)
    price = mapped_column(Float, nullable=False)
    timestamp = mapped_column(DateTime, default=datetime.utcnow)

    __tablename__ = "crypto_history"

    def __repr__(self):
        return f"<CryptoHistory {self.symbol} - {self.price} at {self.timestamp}>"

