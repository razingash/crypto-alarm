from sqlalchemy import String, SmallInteger, ForeignKey, Integer, Boolean
from sqlalchemy.orm import Mapped, mapped_column, relationship

from core.models.base import Base

__all__ = ["TriggerFormula", "TriggerFormulaComponent"]

"""
MongoDB
class TriggersHistory(Base):
    pass
"""


class TriggerFormula(Base):
    owner_id: Mapped[int] = mapped_column(Integer, ForeignKey("user_user.id"), nullable=False)
    is_notified: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False) # тут могут быть баги при большой нагрузке
    is_active: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False, index=True)
    formula: Mapped[str] = mapped_column(String, nullable=False)
    components: Mapped[list["TriggerFormulaComponent"]] = relationship("TriggerFormulaComponent", back_populates="owner")

    __tablename__ = "trigger_formula"


class TriggerFormulaComponent(Base):
    """брать во внимание только компоненты с amount > 1"""
    component: Mapped[str] = mapped_column(String, nullable=False)
    amount: Mapped[int] = mapped_column(SmallInteger, nullable=False, default=1, index=True)
    # amount должен повышатся и понижатся в зависимости от количества указывающих на него is_active

    __tablename__ = "trigger_formula_component"
