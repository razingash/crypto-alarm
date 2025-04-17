from datetime import datetime

from sqlalchemy import String, DateTime, Boolean, ForeignKey, Integer
from sqlalchemy.orm import mapped_column, Mapped, relationship

from core.models.base import Base

__all__ = ["CryptoApi", "CryptoParams", "CryptoCurrencies", "TriggerFormula", "TriggerComponent", "TriggerFormulaComponent"]


class CryptoApi(Base):
    """список доступных апи"""
    api: Mapped[str] = mapped_column(String(500), nullable=False, unique=True, index=True)
    is_actual: Mapped[bool] = mapped_column(Boolean, nullable=False, default=True)
    last_updated: Mapped[bool] = mapped_column(DateTime, nullable=False, default=datetime.utcnow)

    params = relationship("CryptoParams", back_populates="crypto_api", cascade="all, delete-orphan")
    trigger_components = relationship("TriggerComponent", back_populates="api", cascade="all, delete-orphan")

    __tablename__ = "crypto_api"


class CryptoParams(Base):
    """доступные на данный момент параметры апи"""
    parameter: Mapped[str] = mapped_column(String(500), nullable=False)
    is_active: Mapped[bool] = mapped_column(Boolean, nullable=False, default=True)
    last_updated: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow)
    excluded_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow)

    crypto_api_id: Mapped[int] = mapped_column(Integer, ForeignKey("crypto_api.id", ondelete="CASCADE"), nullable=False)
    crypto_api = relationship("CryptoApi", back_populates="params")
    trigger_components = relationship("TriggerComponent", back_populates="parameter", cascade="all, delete-orphan")

    __tablename__ = "crypto_params"


class CryptoCurrencies(Base):
    """
    универсальные доступные криптовалюты
    Note:
        доступные валюты получаются путем нахождения общего множества среди доступных валют из апи:
        1) https://api.binance.com/api/v3/ticker/price
        2) https://api.binance.com/api/v3/ticker/24hr
        3) https://api.binance.com/api/v3/exchangeInfo
        нужно сравнивать данные из нескольких апи поскольку новые валюты добавляются неравномерно
    """
    currency: Mapped[str] = mapped_column(String(500), nullable=False)
    is_available: Mapped[bool] = mapped_column(Boolean, nullable=False, default=True)
    last_updated: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow)
    trigger_components = relationship("TriggerComponent", back_populates="currency", cascade="all, delete-orphan")

    __tablename__ = "crypto_currencies"


class TriggerComponent(Base):
    api_id: Mapped[int] = mapped_column(ForeignKey("crypto_api.id", ondelete="CASCADE"), nullable=False)
    currency_id: Mapped[int] = mapped_column(ForeignKey("crypto_currencies.id", ondelete="CASCADE"), nullable=False)
    parameter_id: Mapped[int] = mapped_column(ForeignKey("crypto_params.id", ondelete="CASCADE"), nullable=False)

    api = relationship("CryptoApi", back_populates="trigger_components")
    currency = relationship("CryptoCurrencies", back_populates="trigger_components")
    parameter = relationship("CryptoParams", back_populates="trigger_components")
    trigger_formula_components = relationship("TriggerFormulaComponent", back_populates="component", cascade="all, delete-orphan")

    __tablename__ = "trigger_component"


class TriggerFormulaComponent(Base):
    """
    брать во внимание только компоненты с amount > 1\n
    amount должен повышатся и понижатся в зависимости от количества указывающих на него is_active
    """
    component_id: Mapped[int] = mapped_column(Integer, ForeignKey("trigger_component.id", ondelete="CASCADE"), nullable=False)
    formula_id: Mapped[int] = mapped_column(Integer, ForeignKey("trigger_formula.id", ondelete="CASCADE"), nullable=False)

    component = relationship("TriggerComponent", back_populates="trigger_formula_components")
    formula = relationship("TriggerFormula", back_populates="components")

    __tablename__ = "trigger_formula_component"


class TriggerFormula(Base):
    """формула может быть активной, но если она не будет фиксировать историю, или отправлять уведомления, то она будет бесполезной"""
    formula: Mapped[str] = mapped_column(String, nullable=False)
    name: Mapped[str] = mapped_column(String(150), nullable=False)
    description: Mapped[str] = mapped_column(String(1500), nullable=True)
    is_notified: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False) # тут могут быть баги при большой нагрузке
    is_active: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False, index=True)
    is_history_on: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False) # не работает сейчас!
    is_shutted_off: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False) # оффается из-за изменений в апи
    last_triggered: Mapped[datetime] = mapped_column(DateTime, nullable=True)
    cooldown: Mapped[int] = mapped_column(Integer, nullable=False, default=3600) # час по дефолту, может сделать больше

    owner_id: Mapped[int] = mapped_column(Integer, ForeignKey("user_user.id"), nullable=False)
    owner = relationship("User", back_populates="triggers")
    components: Mapped[list["TriggerFormulaComponent"]] = relationship("TriggerFormulaComponent", back_populates="formula")

    __tablename__ = "trigger_formula"


class PushSubscription(Base):
    endpoint: Mapped[str] = mapped_column(String, nullable=False)
    p256dh: Mapped[str] = mapped_column(String, nullable=False)
    auth: Mapped[str] = mapped_column(String, nullable=False)
    created_at: Mapped[datetime] = mapped_column(DateTime, nullable=False)

    __tablename__ = "trigger_push_subscription"


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
