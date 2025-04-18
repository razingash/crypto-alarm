from datetime import datetime
from uuid import uuid4

from sqlalchemy import Boolean, String, DateTime
from sqlalchemy.orm import mapped_column, Mapped, relationship

from core.models.base import Base

__all__ = ["User", ]

class User(Base):
    uuid: Mapped[str] = mapped_column(String(36), default=lambda: str(uuid4()), unique=True, index=True)
    username: Mapped[str] = mapped_column(String(50), unique=True, nullable=False, index=True)
    password: Mapped[str] = mapped_column(String, nullable=False, index=True)
    registered_date: Mapped[datetime] = mapped_column(DateTime, nullable=False)
    ispremium: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False, index=True)
    since: Mapped[datetime] = mapped_column(DateTime, nullable=True)
    to: Mapped[datetime] = mapped_column(DateTime, nullable=True)

    access_tokens: Mapped[list["AccessToken"]] = relationship("AccessToken", back_populates="user")
    refresh_token: Mapped["RefreshToken"] = relationship("RefreshToken", back_populates="user", uselist=False)
    triggers: Mapped[list["TriggerFormula"]] = relationship("TriggerFormula", back_populates="owner")
    devices: Mapped[list["PushSubscription"]] = relationship("PushSubscription", back_populates="user")

    __tablename__ = "user_user"
