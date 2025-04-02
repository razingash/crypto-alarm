from datetime import datetime

from sqlalchemy import String, SmallInteger, ForeignKey, Integer, Boolean, DateTime
from sqlalchemy.orm import Mapped, mapped_column, relationship

from core.models.base import Base

__all__ = ["TriggerFormula", "TriggerFormulaComponent"]

"""
MongoDB
class TriggersHistory(Base):
    pass
"""
"""
Нужно также добавить таблицу истории стратегии, именно в постгрессе, чтобы была возможность строить график конкретной
стратегии.
Возможно также стоит добавить в MongoDB 
"""

class TriggerFormula(Base):
    formula: Mapped[str] = mapped_column(String, nullable=False)
    name: Mapped[str] = mapped_column(String(150), nullable=True)
    description: Mapped[str] = mapped_column(String(1500), nullable=True)
    is_notified: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False) # тут могут быть баги при большой нагрузке
    is_active: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False, index=True)
    is_history_on: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False) # не работает сейчас!
    is_shutted_off: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False) # оффается из-за изменений в апи
    last_triggered: Mapped[datetime] = mapped_column(DateTime, nullable=True)

    owner_id: Mapped[int] = mapped_column(Integer, ForeignKey("user_user.id"), nullable=False)
    owner = relationship("User", back_populates="triggers")
    components: Mapped[list["TriggerFormulaComponent"]] = relationship("TriggerFormulaComponent", back_populates="formula")

    __tablename__ = "trigger_formula"


class TriggerFormulaComponent(Base):
    """
    брать во внимание только компоненты с amount > 1\n
    amount должен повышатся и понижатся в зависимости от количества указывающих на него is_active
    """
    component: Mapped[str] = mapped_column(String, nullable=False)
    amount: Mapped[int] = mapped_column(SmallInteger, nullable=False, default=1, index=True)

    formula_id: Mapped[int] = mapped_column(Integer, ForeignKey("trigger_formula.id"), nullable=False)
    formula = relationship("TriggerFormula", back_populates="components")

    __tablename__ = "trigger_formula_component"
