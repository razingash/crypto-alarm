from sqlalchemy import Float, Boolean
from sqlalchemy.orm import Mapped, mapped_column

from core.models.base import BaseTrigger

__all__ = ["TriggerExitingPrice"]

"""
MongoDB
class TriggersHistory(Base):
    pass
"""

class TriggerExitingPrice(BaseTrigger):
    """триггер пробития порога, должен срабатывать когда будет пробитие порога цены, is_percentage влияет на то будет
    ли это работать для процентного значения или фиксирования цены"""
    min_price: Mapped[float] = mapped_column(Float, nullable=False)
    max_price: Mapped[float] = mapped_column(Float, nullable=False)
    fixed_price: Mapped[float] = mapped_column(Float, nullable=False)  # сделать дефолтным, чтобы оно само устанавливалось в самом начале
    is_percentage: Mapped[bool] = mapped_column(Boolean, nullable=False)

    __tablename__ = "trigger_exiting_price"


# также нужна модель для "скачка" но это уже будет алгоритмическая торговля так что лучше не делать
