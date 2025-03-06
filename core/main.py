from argparse import ArgumentParser
from uvicorn import run as uvicorn_run

from fastapi import FastAPI
from contextlib import asynccontextmanager

@asynccontextmanager
async def lifespan(app: FastAPI):
    print("START")
    yield
    print("END")

app = FastAPI(lifespan=lifespan)

if __name__ == "__main__":
    parser = ArgumentParser(description="Run FastAPI server")
    parser.add_argument("--addr", type=str, default="127.0.0.1:8000", help="Host and port to bind, e.g. 0.0.0.0:8000")
    args = parser.parse_args()

    host, port = args.addr.split(":")
    port = int(port)

    uvicorn_run("main:app", host=host, port=port, reload=True)
