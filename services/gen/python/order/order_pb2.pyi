from google.protobuf import timestamp_pb2 as _timestamp_pb2
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

class OrderStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    PENDING: _ClassVar[OrderStatus]
    OPEN: _ClassVar[OrderStatus]
    PARTIAL_FILL: _ClassVar[OrderStatus]
    FILLED: _ClassVar[OrderStatus]
    CANCELLED: _ClassVar[OrderStatus]
    REJECTED: _ClassVar[OrderStatus]
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
PENDING: OrderStatus
OPEN: OrderStatus
PARTIAL_FILL: OrderStatus
FILLED: OrderStatus
CANCELLED: OrderStatus
REJECTED: OrderStatus

class Order(_message.Message):
    __slots__ = ("id", "user_id", "symbol", "side", "order_type", "time_in_force", "qty", "filled_qty", "price", "stop_price", "status", "created_at", "updated_at")
    ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    SYMBOL_FIELD_NUMBER: _ClassVar[int]
    SIDE_FIELD_NUMBER: _ClassVar[int]
    ORDER_TYPE_FIELD_NUMBER: _ClassVar[int]
    TIME_IN_FORCE_FIELD_NUMBER: _ClassVar[int]
    QTY_FIELD_NUMBER: _ClassVar[int]
    FILLED_QTY_FIELD_NUMBER: _ClassVar[int]
    PRICE_FIELD_NUMBER: _ClassVar[int]
    STOP_PRICE_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    id: str
    user_id: str
    symbol: str
    side: Side
    order_type: OrderType
    time_in_force: TimeInForce
    qty: int
    filled_qty: int
    price: int
    stop_price: int
    status: OrderStatus
    created_at: _timestamp_pb2.Timestamp
    updated_at: _timestamp_pb2.Timestamp
    def __init__(self, id: _Optional[str] = ..., user_id: _Optional[str] = ..., symbol: _Optional[str] = ..., side: _Optional[_Union[Side, str]] = ..., order_type: _Optional[_Union[OrderType, str]] = ..., time_in_force: _Optional[_Union[TimeInForce, str]] = ..., qty: _Optional[int] = ..., filled_qty: _Optional[int] = ..., price: _Optional[int] = ..., stop_price: _Optional[int] = ..., status: _Optional[_Union[OrderStatus, str]] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., updated_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class PlaceOrderRequest(_message.Message):
    __slots__ = ("user_id", "symbol", "side", "order_type", "time_in_force", "qty", "price", "stop_price", "reference_price_cents", "strategy_id")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    SYMBOL_FIELD_NUMBER: _ClassVar[int]
    SIDE_FIELD_NUMBER: _ClassVar[int]
    ORDER_TYPE_FIELD_NUMBER: _ClassVar[int]
    TIME_IN_FORCE_FIELD_NUMBER: _ClassVar[int]
    QTY_FIELD_NUMBER: _ClassVar[int]
    PRICE_FIELD_NUMBER: _ClassVar[int]
    STOP_PRICE_FIELD_NUMBER: _ClassVar[int]
    REFERENCE_PRICE_CENTS_FIELD_NUMBER: _ClassVar[int]
    STRATEGY_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    symbol: str
    side: Side
    order_type: OrderType
    time_in_force: TimeInForce
    qty: int
    price: int
    stop_price: int
    reference_price_cents: int
    strategy_id: str
    def __init__(self, user_id: _Optional[str] = ..., symbol: _Optional[str] = ..., side: _Optional[_Union[Side, str]] = ..., order_type: _Optional[_Union[OrderType, str]] = ..., time_in_force: _Optional[_Union[TimeInForce, str]] = ..., qty: _Optional[int] = ..., price: _Optional[int] = ..., stop_price: _Optional[int] = ..., reference_price_cents: _Optional[int] = ..., strategy_id: _Optional[str] = ...) -> None: ...

class PlaceOrderResponse(_message.Message):
    __slots__ = ("order",)
    ORDER_FIELD_NUMBER: _ClassVar[int]
    order: Order
    def __init__(self, order: _Optional[_Union[Order, _Mapping]] = ...) -> None: ...

class CancelOrderRequest(_message.Message):
    __slots__ = ("order_id", "user_id")
    ORDER_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    order_id: str
    user_id: str
    def __init__(self, order_id: _Optional[str] = ..., user_id: _Optional[str] = ...) -> None: ...

class CancelOrderResponse(_message.Message):
    __slots__ = ("success", "order")
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    ORDER_FIELD_NUMBER: _ClassVar[int]
    success: bool
    order: Order
    def __init__(self, success: bool = ..., order: _Optional[_Union[Order, _Mapping]] = ...) -> None: ...

class GetOrdersRequest(_message.Message):
    __slots__ = ("user_id", "status", "all_statuses", "limit", "offset")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    ALL_STATUSES_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    OFFSET_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    status: OrderStatus
    all_statuses: bool
    limit: int
    offset: int
    def __init__(self, user_id: _Optional[str] = ..., status: _Optional[_Union[OrderStatus, str]] = ..., all_statuses: bool = ..., limit: _Optional[int] = ..., offset: _Optional[int] = ...) -> None: ...

class GetOrdersResponse(_message.Message):
    __slots__ = ("orders",)
    ORDERS_FIELD_NUMBER: _ClassVar[int]
    orders: _containers.RepeatedCompositeFieldContainer[Order]
    def __init__(self, orders: _Optional[_Iterable[_Union[Order, _Mapping]]] = ...) -> None: ...

class GetOrderRequest(_message.Message):
    __slots__ = ("order_id",)
    ORDER_ID_FIELD_NUMBER: _ClassVar[int]
    order_id: str
    def __init__(self, order_id: _Optional[str] = ...) -> None: ...

class UpdateOrderStatusRequest(_message.Message):
    __slots__ = ("order_id", "status", "filled_qty")
    ORDER_ID_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    FILLED_QTY_FIELD_NUMBER: _ClassVar[int]
    order_id: str
    status: OrderStatus
    filled_qty: int
    def __init__(self, order_id: _Optional[str] = ..., status: _Optional[_Union[OrderStatus, str]] = ..., filled_qty: _Optional[int] = ...) -> None: ...

class UpdateOrderStatusResponse(_message.Message):
    __slots__ = ("success",)
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    success: bool
    def __init__(self, success: bool = ...) -> None: ...

class Fill(_message.Message):
    __slots__ = ("trade_id", "order_id", "symbol", "side", "qty", "price_cents", "liquidity", "executed_at")
    TRADE_ID_FIELD_NUMBER: _ClassVar[int]
    ORDER_ID_FIELD_NUMBER: _ClassVar[int]
    SYMBOL_FIELD_NUMBER: _ClassVar[int]
    SIDE_FIELD_NUMBER: _ClassVar[int]
    QTY_FIELD_NUMBER: _ClassVar[int]
    PRICE_CENTS_FIELD_NUMBER: _ClassVar[int]
    LIQUIDITY_FIELD_NUMBER: _ClassVar[int]
    EXECUTED_AT_FIELD_NUMBER: _ClassVar[int]
    trade_id: str
    order_id: str
    symbol: str
    side: Side
    qty: int
    price_cents: int
    liquidity: str
    executed_at: _timestamp_pb2.Timestamp
    def __init__(self, trade_id: _Optional[str] = ..., order_id: _Optional[str] = ..., symbol: _Optional[str] = ..., side: _Optional[_Union[Side, str]] = ..., qty: _Optional[int] = ..., price_cents: _Optional[int] = ..., liquidity: _Optional[str] = ..., executed_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class GetFillsRequest(_message.Message):
    __slots__ = ("user_id", "limit", "offset")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    OFFSET_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    limit: int
    offset: int
    def __init__(self, user_id: _Optional[str] = ..., limit: _Optional[int] = ..., offset: _Optional[int] = ...) -> None: ...

class GetStrategyFillsRequest(_message.Message):
    __slots__ = ("strategy_id", "user_id", "limit", "offset")
    STRATEGY_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    OFFSET_FIELD_NUMBER: _ClassVar[int]
    strategy_id: str
    user_id: str
    limit: int
    offset: int
    def __init__(self, strategy_id: _Optional[str] = ..., user_id: _Optional[str] = ..., limit: _Optional[int] = ..., offset: _Optional[int] = ...) -> None: ...

class GetFillsResponse(_message.Message):
    __slots__ = ("fills",)
    FILLS_FIELD_NUMBER: _ClassVar[int]
    fills: _containers.RepeatedCompositeFieldContainer[Fill]
    def __init__(self, fills: _Optional[_Iterable[_Union[Fill, _Mapping]]] = ...) -> None: ...
