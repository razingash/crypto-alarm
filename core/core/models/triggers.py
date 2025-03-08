from sqlalchemy import Float
from sqlalchemy.orm import Mapped, mapped_column

from core.models.base import Base


__all__ = ["TriggerSimple"]
#class TriggersHistory(Base):
#    pass

class TriggerSimple(Base):
    """триггер пробития порога"""
    fixed_price: Mapped[float] = mapped_column(Float, nullable=False)
    min_leap: Mapped[float] = mapped_column(Float, nullable=False)
    max_leap: Mapped[float] = mapped_column(Float, nullable=False)

    __tablename__ = "trigger_simple"

