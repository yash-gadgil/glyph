from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class OrderType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    MARKET: _ClassVar[OrderType]
    LIMIT: _ClassVar[OrderType]
    STOP: _ClassVar[OrderType]
    STOP_LIMIT: _ClassVar[OrderType]

class Side(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    BUY: _ClassVar[Side]
    SELL: _ClassVar[Side]

class TimeInForce(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    DAY: _ClassVar[TimeInForce]
    GTC: _ClassVar[TimeInForce]
    IOC: _ClassVar[TimeInForce]
    FOK: _ClassVar[TimeInForce]
MARKET: OrderType
LIMIT: OrderType
STOP: OrderType
STOP_LIMIT: OrderType
BUY: Side
SELL: Side
DAY: TimeInForce
GTC: TimeInForce
IOC: TimeInForce
FOK: TimeInForce

class AddOrderRequest(_message.Message):
    __slots__ = ("id", "user_id", "symbol", "side", "order_type", "time_in_force", "qty", "price", "stop_price")
    ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    SYMBOL_FIELD_NUMBER: _ClassVar[int]
    SIDE_FIELD_NUMBER: _ClassVar[int]
    ORDER_TYPE_FIELD_NUMBER: _ClassVar[int]
    TIME_IN_FORCE_FIELD_NUMBER: _ClassVar[int]
    QTY_FIELD_NUMBER: _ClassVar[int]
    PRICE_FIELD_NUMBER: _ClassVar[int]
    STOP_PRICE_FIELD_NUMBER: _ClassVar[int]
    id: str
    user_id: str
    symbol: str
    side: Side
    order_type: OrderType
    time_in_force: TimeInForce
    qty: int
    price: int
    stop_price: int
    def __init__(self, id: _Optional[str] = ..., user_id: _Optional[str] = ..., symbol: _Optional[str] = ..., side: _Optional[_Union[Side, str]] = ..., order_type: _Optional[_Union[OrderType, str]] = ..., time_in_force: _Optional[_Union[TimeInForce, str]] = ..., qty: _Optional[int] = ..., price: _Optional[int] = ..., stop_price: _Optional[int] = ...) -> None: ...

class TradeInfo(_message.Message):
    __slots__ = ("order_id", "price", "quantity")
    ORDER_ID_FIELD_NUMBER: _ClassVar[int]
    PRICE_FIELD_NUMBER: _ClassVar[int]
    QUANTITY_FIELD_NUMBER: _ClassVar[int]
    order_id: str
    price: int
    quantity: int
    def __init__(self, order_id: _Optional[str] = ..., price: _Optional[int] = ..., quantity: _Optional[int] = ...) -> None: ...

class Trade(_message.Message):
    __slots__ = ("bid_trade", "ask_trade")
    BID_TRADE_FIELD_NUMBER: _ClassVar[int]
    ASK_TRADE_FIELD_NUMBER: _ClassVar[int]
    bid_trade: TradeInfo
    ask_trade: TradeInfo
    def __init__(self, bid_trade: _Optional[_Union[TradeInfo, _Mapping]] = ..., ask_trade: _Optional[_Union[TradeInfo, _Mapping]] = ...) -> None: ...

class AddOrderResponse(_message.Message):
    __slots__ = ("accepted", "trades")
    ACCEPTED_FIELD_NUMBER: _ClassVar[int]
    TRADES_FIELD_NUMBER: _ClassVar[int]
    accepted: bool
    trades: _containers.RepeatedCompositeFieldContainer[Trade]
    def __init__(self, accepted: bool = ..., trades: _Optional[_Iterable[_Union[Trade, _Mapping]]] = ...) -> None: ...

class CancelOrderRequest(_message.Message):
    __slots__ = ("order_id", "symbol")
    ORDER_ID_FIELD_NUMBER: _ClassVar[int]
    SYMBOL_FIELD_NUMBER: _ClassVar[int]
    order_id: str
    symbol: str
    def __init__(self, order_id: _Optional[str] = ..., symbol: _Optional[str] = ...) -> None: ...

class CancelOrderResponse(_message.Message):
    __slots__ = ("cancelled",)
    CANCELLED_FIELD_NUMBER: _ClassVar[int]
    cancelled: bool
    def __init__(self, cancelled: bool = ...) -> None: ...

class InjectPriceRequest(_message.Message):
    __slots__ = ("symbol", "price_cents")
    SYMBOL_FIELD_NUMBER: _ClassVar[int]
    PRICE_CENTS_FIELD_NUMBER: _ClassVar[int]
    symbol: str
    price_cents: int
    def __init__(self, symbol: _Optional[str] = ..., price_cents: _Optional[int] = ...) -> None: ...

class InjectPriceResponse(_message.Message):
    __slots__ = ("fills",)
    FILLS_FIELD_NUMBER: _ClassVar[int]
    fills: int
    def __init__(self, fills: _Optional[int] = ...) -> None: ...
