from urllib.request import Request

from starlette.middleware.base import BaseHTTPMiddleware

from core.endpoints import endpoints
from core.logger import custom_logger


class WeightTrackingMiddleware(BaseHTTPMiddleware):
    """Мидлвейр для динамической актуализации весов"""
    def __init__(self, app):
        super().__init__(app)
        self.update_mode = False
        self.pending_endpoints = set()

    async def dispatch(self, request: Request, call_next):
        response = await call_next(request)

        if self.update_mode:
            endpoint = request.url.path # full_url?
            print(f'update mode: {endpoint}', request.full_url, request.full_url.path, request.url)

            if endpoint not in self.pending_endpoints:
                used_weight = response.headers.get("X-MBX-API-WEIGHT", "0")

                try:
                    used_weight = int(used_weight)
                except ValueError:
                    used_weight = 0 # скорее всего не 0

                current_weight = endpoints[endpoint]

                if used_weight != current_weight:
                    endpoints[endpoint] = used_weight
                    custom_logger.log_with_path(
                        level=3,
                        msg=f"Endpoint updated {endpoint}: {current_weight} -> {used_weight}",
                        path="endpoints.log"
                    )

                self.pending_endpoints.discard(endpoint)

                if not self.pending_endpoints:
                    self.disable_update_mode()

        return response

    def enable_update_mode(self):
        self.update_mode = True
        self.pending_endpoints = set(endpoints.keys())

    def disable_update_mode(self):
        self.update_mode = False
