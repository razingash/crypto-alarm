import asyncio
from argparse import ArgumentParser
from uvicorn import run as uvicorn_run

from fastapi import FastAPI
from contextlib import asynccontextmanager

from apps.binance.external_router import BinanceAPI
from apps.analytics.router import router as router_analytics
from core.analysis.manager import formula_manager
from core.controller import controller
from core.middlewares import WeightTrackingMiddleware
from core.orchestrator import BinanceAPIOrchestrator


@asynccontextmanager
async def lifespan(app: FastAPI):
    print("START")

    app.state.queue_event = asyncio.Event()
    await controller.start(app.state.queue_event)
    await binance_api.check_and_update_weights()
    await formula_manager.load()
    await orchestrator.start()

    yield
    print("END")


app = FastAPI(lifespan=lifespan)
middleware_binance_api_weight = WeightTrackingMiddleware(app=app)
binance_api = BinanceAPI(controller=controller, middleware=middleware_binance_api_weight)
orchestrator = BinanceAPIOrchestrator(binance_api)
formula_manager.set_orchestrator(orchestrator)

app.include_router(router_analytics, prefix="/api/v1/analytics")

if __name__ == "__main__":
    parser = ArgumentParser(description="Run FastAPI server")
    parser.add_argument("--addr", type=str, default="127.0.0.1:8000", help="Host and port to bind, e.g. 0.0.0.0:8000")
    args = parser.parse_args()

    host, port = args.addr.split(":")
    port = int(port)

    uvicorn_run("main:app", host=host, port=port, reload=True)
