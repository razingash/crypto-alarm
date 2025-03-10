from urllib.request import Request

from fastapi import HTTPException
from starlette.middleware.base import BaseHTTPMiddleware

from core.controller import BinanceAPIController


class WeightTrackingMiddleware(BaseHTTPMiddleware):
    def __init__(self, app, rate_limiter: BinanceAPIController):
        super().__init__(app)
        self.rate_limiter = rate_limiter

    async def dispatch(self, request: Request, call_next):
        response = await call_next(request)

        used_weight = response.headers.get("X-MBX-API-WEIGHT", "0")
        try:
            used_weight = int(used_weight)
        except ValueError:
            used_weight = 0

        if not await self.rate_limiter.add_weight(used_weight):
            raise HTTPException(status_code=429, detail="Rate limit exceeded. Try again later.")

        return response
