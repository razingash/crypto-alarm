from pydantic import BaseModel


class TickerCurrentPriceResponse(BaseModel):
    symbol: str
    price: str


class Ticker24hrResponse(BaseModel):
    """
    {"symbol": "ETHBTC", "priceChange": "-0.00004000", "priceChangePercent": "-0.158", "weightedAvgPrice": "0.02551251",
     "prevClosePrice": "0.02527000", "lastPrice": "0.02522000", "lastQty": "0.25940000", "bidPrice": "0.02521000",
     "bidQty": "39.79540000", "askPrice": "0.02522000", "askQty": "42.10570000", "openPrice": "0.02526000",
     "highPrice": "0.02601000", "lowPrice": "0.02511000", "volume": "24013.75370000", "quoteVolume": "612.65117525",
     "openTime": 1741448886710, "closeTime": 1741535286710, "firstId": 490112110, "lastId": 490168843, "count": 56734}
    """
    symbol: str
    priceChange: str
    priceChangePercent: str
    weightedAvgPrice: str
    prevClosePrice: str
    lastPrice: str
    lastQty: str
    bidPrice: str
    bidQty: str
    askPrice: str
    askQty: str
    openPrice: str
    highPrice: str
    lowPrice: str
    volume: str
    quoteVolume: str
    openTime: int
    closeTime: int
    firstId: int
    lastId: int
    count: int

