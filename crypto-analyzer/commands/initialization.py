import asyncio

from apps.binance.external_router import BinanceAPI
from core.controller import controller
from core.endpoints import endpoints
from core.models import CryptoParams, CryptoApi, CryptoCurrencies
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
    binance_api = BinanceAPI(controller=controller, middleware=None)
    loop.run_until_complete(get_initial_data_params(binance_api))
    loop.run_until_complete(get_valid_currencies(binance_api))

async def get_initial_data_params(binance_api: BinanceAPI) -> dict:
    """
    receives the keys to the datasets regarding the API Binance to initialize them in the database
    and use in the keyboard on the client side
    """
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

async def get_valid_currencies(binance_api: BinanceAPI) -> None:
    """получает нормальные значения для криптовалют"""
    dataset = {}
    api = [
        '/v3/ticker/price',
        '/v3/ticker/24hr',
        '/v3/exchangeInfo'
    ]

    for i in range(len(api)):
        data = await binance_api.get(endpoint=api[i], weight=80)
        if data:
            if i == 0:
                dataset["ticker_price_symbols"] = [item["symbol"] for item in data]
            if i == 1:
                dataset["ticker_24hr_symbols"] = [item["symbol"] for item in data]
            if i == 2:
                dataset["exchange_info_symbols"] = [item["symbol"] for item in data["symbols"]]

    avg_currencies = set(dataset["ticker_price_symbols"]) & set(dataset["ticker_24hr_symbols"]) & set(dataset["exchange_info_symbols"])

    async with postgres_db.session_factory() as session:
        for currency in avg_currencies:
            crypto_currency = CryptoCurrencies(currency=currency)
            session.add(crypto_currency)
        await session.commit()

