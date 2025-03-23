import asyncio

from apps.binance.external_router import BinanceAPI
from core.controller import controller
from core.endpoints import endpoints
from core.models import CryptoParams, CryptoApi
from db.postgre import postgres_db
from .makemigrations import command_makemigrations

def command_initialization():
    command_makemigrations()
    fill_crypto_models()


def fill_crypto_models() -> None:
    """заполняет CryptoApi и CryptoParams модели полученными данными из списка апи в endpoints.py"""
    try:
        loop = asyncio.get_running_loop()
    except RuntimeError:
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
    loop.run_until_complete(get_initial_data_params())

async def get_initial_data_params() -> dict:
    """
    receives the keys to the datasets regarding the API Binance to initialize them in the database
    and use in the keyboard on the client side
    """
    binance_api = BinanceAPI(controller=controller, middleware=None)
    dataset = {}

    excluded_endpoints = ["/v3/ping"]
    for endpoint, weigth in endpoints.items():
        if endpoint not in excluded_endpoints:
            # weigth не точный но это не критично поскольку апи не так много и весы будут меньше из-за ?symbol
            data = await binance_api.get(
                endpoint=f"{endpoint}?symbol=ETHBTC",
                weight=weigth
            )
            dataset[endpoint] = data

    await initialize_crypto_models(dataset)

    return dataset

async def initialize_crypto_models(dataset: dict) -> None:
    """Initialization of variables for user strategies"""
    async with postgres_db.session_factory() as session:
        for api, data in dataset.items():
            crypto_api = CryptoApi(api=api)
            session.add(crypto_api)

            await session.flush()

            for parameter in data.keys():
                crypto_param = CryptoParams(parameter=parameter, crypto_api_id=crypto_api.id)
                session.add(crypto_param)

        await session.commit()

