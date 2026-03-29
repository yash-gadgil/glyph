from google.protobuf import empty_pb2 as _empty_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Timeframe(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    DAY: _ClassVar[Timeframe]
    HOUR: _ClassVar[Timeframe]
    MIN: _ClassVar[Timeframe]
DAY: Timeframe
HOUR: Timeframe
MIN: Timeframe

class NewsRequest(_message.Message):
    __slots__ = ("symbols", "limit")
    SYMBOLS_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    symbols: _containers.RepeatedScalarFieldContainer[str]
    limit: int
    def __init__(self, symbols: _Optional[_Iterable[str]] = ..., limit: _Optional[int] = ...) -> None: ...

class NewsArticle(_message.Message):
    __slots__ = ("id", "headline", "summary", "source", "url", "symbols", "image_url", "created_at")
    ID_FIELD_NUMBER: _ClassVar[int]
    HEADLINE_FIELD_NUMBER: _ClassVar[int]
    SUMMARY_FIELD_NUMBER: _ClassVar[int]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    URL_FIELD_NUMBER: _ClassVar[int]
    SYMBOLS_FIELD_NUMBER: _ClassVar[int]
    IMAGE_URL_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    id: str
    headline: str
    summary: str
    source: str
    url: str
    symbols: _containers.RepeatedScalarFieldContainer[str]
    image_url: str
    created_at: str
    def __init__(self, id: _Optional[str] = ..., headline: _Optional[str] = ..., summary: _Optional[str] = ..., source: _Optional[str] = ..., url: _Optional[str] = ..., symbols: _Optional[_Iterable[str]] = ..., image_url: _Optional[str] = ..., created_at: _Optional[str] = ...) -> None: ...

class NewsResponse(_message.Message):
    __slots__ = ("articles",)
    ARTICLES_FIELD_NUMBER: _ClassVar[int]
    articles: _containers.RepeatedCompositeFieldContainer[NewsArticle]
    def __init__(self, articles: _Optional[_Iterable[_Union[NewsArticle, _Mapping]]] = ...) -> None: ...

class MoversRequest(_message.Message):
    __slots__ = ("limit",)
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    limit: int
    def __init__(self, limit: _Optional[int] = ...) -> None: ...

class Mover(_message.Message):
    __slots__ = ("symbol", "company_name", "price_cents", "change_percent", "volume")
    SYMBOL_FIELD_NUMBER: _ClassVar[int]
    COMPANY_NAME_FIELD_NUMBER: _ClassVar[int]
    PRICE_CENTS_FIELD_NUMBER: _ClassVar[int]
    CHANGE_PERCENT_FIELD_NUMBER: _ClassVar[int]
    VOLUME_FIELD_NUMBER: _ClassVar[int]
    symbol: str
    company_name: str
    price_cents: int
    change_percent: float
    volume: int
    def __init__(self, symbol: _Optional[str] = ..., company_name: _Optional[str] = ..., price_cents: _Optional[int] = ..., change_percent: _Optional[float] = ..., volume: _Optional[int] = ...) -> None: ...

class MoversResponse(_message.Message):
    __slots__ = ("gainers", "losers")
    GAINERS_FIELD_NUMBER: _ClassVar[int]
    LOSERS_FIELD_NUMBER: _ClassVar[int]
    gainers: _containers.RepeatedCompositeFieldContainer[Mover]
    losers: _containers.RepeatedCompositeFieldContainer[Mover]
    def __init__(self, gainers: _Optional[_Iterable[_Union[Mover, _Mapping]]] = ..., losers: _Optional[_Iterable[_Union[Mover, _Mapping]]] = ...) -> None: ...

class LatestPricesRequest(_message.Message):
    __slots__ = ("symbols",)
    SYMBOLS_FIELD_NUMBER: _ClassVar[int]
    symbols: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, symbols: _Optional[_Iterable[str]] = ...) -> None: ...

class SymbolPrice(_message.Message):
    __slots__ = ("symbol", "price_cents")
    SYMBOL_FIELD_NUMBER: _ClassVar[int]
    PRICE_CENTS_FIELD_NUMBER: _ClassVar[int]
    symbol: str
    price_cents: int
    def __init__(self, symbol: _Optional[str] = ..., price_cents: _Optional[int] = ...) -> None: ...

class LatestPricesResponse(_message.Message):
    __slots__ = ("prices",)
    PRICES_FIELD_NUMBER: _ClassVar[int]
    prices: _containers.RepeatedCompositeFieldContainer[SymbolPrice]
    def __init__(self, prices: _Optional[_Iterable[_Union[SymbolPrice, _Mapping]]] = ...) -> None: ...

class Bar(_message.Message):
    __slots__ = ("symbol", "open", "high", "low", "close", "time", "volume", "vwap")
    SYMBOL_FIELD_NUMBER: _ClassVar[int]
    OPEN_FIELD_NUMBER: _ClassVar[int]
    HIGH_FIELD_NUMBER: _ClassVar[int]
    LOW_FIELD_NUMBER: _ClassVar[int]
    CLOSE_FIELD_NUMBER: _ClassVar[int]
    TIME_FIELD_NUMBER: _ClassVar[int]
    VOLUME_FIELD_NUMBER: _ClassVar[int]
    VWAP_FIELD_NUMBER: _ClassVar[int]
    symbol: str
    open: float
    high: float
    low: float
    close: float
    time: str
    volume: int
    vwap: float
    def __init__(self, symbol: _Optional[str] = ..., open: _Optional[float] = ..., high: _Optional[float] = ..., low: _Optional[float] = ..., close: _Optional[float] = ..., time: _Optional[str] = ..., volume: _Optional[int] = ..., vwap: _Optional[float] = ...) -> None: ...

class Date(_message.Message):
    __slots__ = ("year", "month", "day", "hour", "min")
    YEAR_FIELD_NUMBER: _ClassVar[int]
    MONTH_FIELD_NUMBER: _ClassVar[int]
    DAY_FIELD_NUMBER: _ClassVar[int]
    HOUR_FIELD_NUMBER: _ClassVar[int]
    MIN_FIELD_NUMBER: _ClassVar[int]
    year: int
    month: int
    day: int
    hour: int
    min: int
    def __init__(self, year: _Optional[int] = ..., month: _Optional[int] = ..., day: _Optional[int] = ..., hour: _Optional[int] = ..., min: _Optional[int] = ...) -> None: ...

class SymbolBars(_message.Message):
    __slots__ = ("symbol", "bars")
    SYMBOL_FIELD_NUMBER: _ClassVar[int]
    BARS_FIELD_NUMBER: _ClassVar[int]
    symbol: str
    bars: _containers.RepeatedCompositeFieldContainer[Bar]
    def __init__(self, symbol: _Optional[str] = ..., bars: _Optional[_Iterable[_Union[Bar, _Mapping]]] = ...) -> None: ...

class HistoricalStockDataRequest(_message.Message):
    __slots__ = ("symbols", "timeframe", "start", "end")
    SYMBOLS_FIELD_NUMBER: _ClassVar[int]
    TIMEFRAME_FIELD_NUMBER: _ClassVar[int]
    START_FIELD_NUMBER: _ClassVar[int]
    END_FIELD_NUMBER: _ClassVar[int]
    symbols: _containers.RepeatedScalarFieldContainer[str]
    timeframe: Timeframe
    start: Date
    end: Date
    def __init__(self, symbols: _Optional[_Iterable[str]] = ..., timeframe: _Optional[_Union[Timeframe, str]] = ..., start: _Optional[_Union[Date, _Mapping]] = ..., end: _Optional[_Union[Date, _Mapping]] = ...) -> None: ...

class HistoricalStockDataResponse(_message.Message):
    __slots__ = ("symbol_bars",)
    SYMBOL_BARS_FIELD_NUMBER: _ClassVar[int]
    symbol_bars: _containers.RepeatedCompositeFieldContainer[SymbolBars]
    def __init__(self, symbol_bars: _Optional[_Iterable[_Union[SymbolBars, _Mapping]]] = ...) -> None: ...

class WatchlistStreamRequest(_message.Message):
    __slots__ = ("action", "symbols")
    class Action(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        START: _ClassVar[WatchlistStreamRequest.Action]
        SUBSCRIBE: _ClassVar[WatchlistStreamRequest.Action]
        UNSUBSCRIBE: _ClassVar[WatchlistStreamRequest.Action]
    START: WatchlistStreamRequest.Action
    SUBSCRIBE: WatchlistStreamRequest.Action
    UNSUBSCRIBE: WatchlistStreamRequest.Action
    ACTION_FIELD_NUMBER: _ClassVar[int]
    SYMBOLS_FIELD_NUMBER: _ClassVar[int]
    action: WatchlistStreamRequest.Action
    symbols: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, action: _Optional[_Union[WatchlistStreamRequest.Action, str]] = ..., symbols: _Optional[_Iterable[str]] = ...) -> None: ...

class MarketUpdate(_message.Message):
    __slots__ = ("symbol_bar",)
    SYMBOL_BAR_FIELD_NUMBER: _ClassVar[int]
    symbol_bar: _containers.RepeatedCompositeFieldContainer[Bar]
    def __init__(self, symbol_bar: _Optional[_Iterable[_Union[Bar, _Mapping]]] = ...) -> None: ...

class Symbol(_message.Message):
    __slots__ = ("name", "company_name")
    NAME_FIELD_NUMBER: _ClassVar[int]
    COMPANY_NAME_FIELD_NUMBER: _ClassVar[int]
    name: str
    company_name: str
    def __init__(self, name: _Optional[str] = ..., company_name: _Optional[str] = ...) -> None: ...

class AvailableSymbolsResponse(_message.Message):
    __slots__ = ("symbols",)
    SYMBOLS_FIELD_NUMBER: _ClassVar[int]
    symbols: _containers.RepeatedCompositeFieldContainer[Symbol]
    def __init__(self, symbols: _Optional[_Iterable[_Union[Symbol, _Mapping]]] = ...) -> None: ...
