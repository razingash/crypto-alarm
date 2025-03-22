from datetime import datetime

from sqlalchemy import String, DateTime, ForeignKey, Boolean
from sqlalchemy.orm import relationship, Mapped, mapped_column

from core.models.base import Base

"""
These models are used in the GO 'JWT' service. here they are only because in this service
there is a setting and initialization of data
"""

class AccessToken(Base):
    user_uuid: Mapped[str] = mapped_column(String(36), ForeignKey('user_user.uuid'), nullable=False)
    token: Mapped[str] = mapped_column(String, nullable=False, unique=True)
    expires_at: Mapped[datetime] = mapped_column(DateTime, nullable=False)
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow)

    user = relationship("User", back_populates="access_tokens") # сейчас тут один ко многим, возможно стоит изменить на один к одному

    __tablename__ = "access_tokens"


class RefreshToken(Base):
    user_uuid: Mapped[str] = mapped_column(String(36), ForeignKey('user_user.uuid'), nullable=False)
    token: Mapped[str] = mapped_column(String, nullable=False, unique=True)
    revoked: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)
    expires_at: Mapped[datetime] = mapped_column(DateTime, nullable=False)
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow)

    user = relationship("User", back_populates="refresh_token", uselist=False)

    __tablename__ = "refresh_tokens"

