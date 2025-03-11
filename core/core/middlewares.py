from urllib.request import Request

from starlette.middleware.base import BaseHTTPMiddleware

from core.controller import BinanceAPIController
from core.endpoints import endpoints
from core.logger import custom_logger


class WeightTrackingMiddleware(BaseHTTPMiddleware):
    """Мидлвейр для динамической актуализации весов"""
    def __init__(self, app, rate_limiter: BinanceAPIController):
        super().__init__(app)
        self.rate_limiter = rate_limiter
        self.update_mode = False
        self.updated_endpoints = set()

    async def dispatch(self, request: Request, call_next):
        response = await call_next(request)

        if self.update_mode:
            endpoint = request.url.path
            print(f'update mode: {endpoint}')

            if endpoint not in self.updated_endpoints:
                used_weight = response.headers.get("X-MBX-API-WEIGHT", "0")

                try:
                    used_weight = int(used_weight)
                except ValueError:
                    used_weight = 0

                current_weight = endpoints[endpoint]

                if used_weight != current_weight:
                    endpoints[endpoint] = used_weight
                    custom_logger.log_with_path(
                        level=3,
                        msg=f"Endpoint updated {endpoint}: {current_weight} -> {used_weight}",
                        path="endpoints.log"
                    )

                self.updated_endpoints.add(endpoint)

        return response

    def enable_update_mode(self):
        self.update_mode = True
        self.updated_endpoints.clear()

    def disable_update_mode(self):
        self.update_mode = False
