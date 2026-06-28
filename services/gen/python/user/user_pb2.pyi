from google.protobuf import empty_pb2 as _empty_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Watchlist(_message.Message):
    __slots__ = ("id", "user_id", "name", "symbols")
    ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    SYMBOLS_FIELD_NUMBER: _ClassVar[int]
    id: str
    user_id: str
    name: str
    symbols: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, id: _Optional[str] = ..., user_id: _Optional[str] = ..., name: _Optional[str] = ..., symbols: _Optional[_Iterable[str]] = ...) -> None: ...

class WatchlistMetadata(_message.Message):
    __slots__ = ("id", "name")
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    id: str
    name: str
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...

class UserSpecifier(_message.Message):
    __slots__ = ("user_id",)
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    def __init__(self, user_id: _Optional[str] = ...) -> None: ...

class WatchlistsResponse(_message.Message):
    __slots__ = ("user_id", "w_metadata", "first")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    W_METADATA_FIELD_NUMBER: _ClassVar[int]
    FIRST_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    w_metadata: _containers.RepeatedCompositeFieldContainer[WatchlistMetadata]
    first: Watchlist
    def __init__(self, user_id: _Optional[str] = ..., w_metadata: _Optional[_Iterable[_Union[WatchlistMetadata, _Mapping]]] = ..., first: _Optional[_Union[Watchlist, _Mapping]] = ...) -> None: ...

class CreateWatchlistRequest(_message.Message):
    __slots__ = ("user_id", "name")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    name: str
    def __init__(self, user_id: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...

class ModifyWatchlistRequest(_message.Message):
    __slots__ = ("id", "user_id", "action", "symbols")
    class Action(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        SUBSCRIBE: _ClassVar[ModifyWatchlistRequest.Action]
        UNSUBSCRIBE: _ClassVar[ModifyWatchlistRequest.Action]
        RENAME: _ClassVar[ModifyWatchlistRequest.Action]
    SUBSCRIBE: ModifyWatchlistRequest.Action
    UNSUBSCRIBE: ModifyWatchlistRequest.Action
    RENAME: ModifyWatchlistRequest.Action
    ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    ACTION_FIELD_NUMBER: _ClassVar[int]
    SYMBOLS_FIELD_NUMBER: _ClassVar[int]
    id: str
    user_id: str
    action: ModifyWatchlistRequest.Action
    symbols: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, id: _Optional[str] = ..., user_id: _Optional[str] = ..., action: _Optional[_Union[ModifyWatchlistRequest.Action, str]] = ..., symbols: _Optional[_Iterable[str]] = ...) -> None: ...

class WatchlistSpecifier(_message.Message):
    __slots__ = ("id", "user_id")
    ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    id: str
    user_id: str
    def __init__(self, id: _Optional[str] = ..., user_id: _Optional[str] = ...) -> None: ...

class DeleteSymbolRequest(_message.Message):
    __slots__ = ("id", "user_id", "symbol")
    ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    SYMBOL_FIELD_NUMBER: _ClassVar[int]
    id: str
    user_id: str
    symbol: str
    def __init__(self, id: _Optional[str] = ..., user_id: _Optional[str] = ..., symbol: _Optional[str] = ...) -> None: ...

class Profile(_message.Message):
    __slots__ = ("user_id", "email", "user_name")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    USER_NAME_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    email: str
    user_name: str
    def __init__(self, user_id: _Optional[str] = ..., email: _Optional[str] = ..., user_name: _Optional[str] = ...) -> None: ...

class AddFundsRequest(_message.Message):
    __slots__ = ("user_id", "amount_cents")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    AMOUNT_CENTS_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    amount_cents: int
    def __init__(self, user_id: _Optional[str] = ..., amount_cents: _Optional[int] = ...) -> None: ...

class AddFundsResponse(_message.Message):
    __slots__ = ("cash_balance_cents",)
    CASH_BALANCE_CENTS_FIELD_NUMBER: _ClassVar[int]
    cash_balance_cents: int
    def __init__(self, cash_balance_cents: _Optional[int] = ...) -> None: ...

class ReserveForOrderRequest(_message.Message):
    __slots__ = ("order_id", "user_id", "symbol", "side", "qty", "cents_per_share")
    ORDER_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    SYMBOL_FIELD_NUMBER: _ClassVar[int]
    SIDE_FIELD_NUMBER: _ClassVar[int]
    QTY_FIELD_NUMBER: _ClassVar[int]
    CENTS_PER_SHARE_FIELD_NUMBER: _ClassVar[int]
    order_id: str
    user_id: str
    symbol: str
    side: int
    qty: int
    cents_per_share: int
    def __init__(self, order_id: _Optional[str] = ..., user_id: _Optional[str] = ..., symbol: _Optional[str] = ..., side: _Optional[int] = ..., qty: _Optional[int] = ..., cents_per_share: _Optional[int] = ...) -> None: ...

class ReleaseForOrderRequest(_message.Message):
    __slots__ = ("order_id",)
    ORDER_ID_FIELD_NUMBER: _ClassVar[int]
    order_id: str
    def __init__(self, order_id: _Optional[str] = ...) -> None: ...

class SignupUserInfo(_message.Message):
    __slots__ = ("user_name", "email", "password")
    USER_NAME_FIELD_NUMBER: _ClassVar[int]
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    PASSWORD_FIELD_NUMBER: _ClassVar[int]
    user_name: str
    email: str
    password: str
    def __init__(self, user_name: _Optional[str] = ..., email: _Optional[str] = ..., password: _Optional[str] = ...) -> None: ...

class SigninUserInfo(_message.Message):
    __slots__ = ("email", "password")
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    PASSWORD_FIELD_NUMBER: _ClassVar[int]
    email: str
    password: str
    def __init__(self, email: _Optional[str] = ..., password: _Optional[str] = ...) -> None: ...

class CheckEmailRequest(_message.Message):
    __slots__ = ("email",)
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    email: str
    def __init__(self, email: _Optional[str] = ...) -> None: ...

class CheckEmailResponse(_message.Message):
    __slots__ = ("available",)
    AVAILABLE_FIELD_NUMBER: _ClassVar[int]
    available: bool
    def __init__(self, available: bool = ...) -> None: ...

class UpdatePasswordRequest(_message.Message):
    __slots__ = ("email", "password_hash")
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    PASSWORD_HASH_FIELD_NUMBER: _ClassVar[int]
    email: str
    password_hash: str
    def __init__(self, email: _Optional[str] = ..., password_hash: _Optional[str] = ...) -> None: ...

class PortfolioHistoryRequest(_message.Message):
    __slots__ = ("user_id", "hours")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    HOURS_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    hours: int
    def __init__(self, user_id: _Optional[str] = ..., hours: _Optional[int] = ...) -> None: ...

class PortfolioHistoryPoint(_message.Message):
    __slots__ = ("time_unix", "equity_cents", "cash_cents", "market_value_cents")
    TIME_UNIX_FIELD_NUMBER: _ClassVar[int]
    EQUITY_CENTS_FIELD_NUMBER: _ClassVar[int]
    CASH_CENTS_FIELD_NUMBER: _ClassVar[int]
    MARKET_VALUE_CENTS_FIELD_NUMBER: _ClassVar[int]
    time_unix: int
    equity_cents: int
    cash_cents: int
    market_value_cents: int
    def __init__(self, time_unix: _Optional[int] = ..., equity_cents: _Optional[int] = ..., cash_cents: _Optional[int] = ..., market_value_cents: _Optional[int] = ...) -> None: ...

class PortfolioHistoryResponse(_message.Message):
    __slots__ = ("points",)
    POINTS_FIELD_NUMBER: _ClassVar[int]
    points: _containers.RepeatedCompositeFieldContainer[PortfolioHistoryPoint]
    def __init__(self, points: _Optional[_Iterable[_Union[PortfolioHistoryPoint, _Mapping]]] = ...) -> None: ...

class PortfolioResponse(_message.Message):
    __slots__ = ("cash_balance_cents", "reserved_cash_cents", "currency", "multiplier", "margin_used_cents")
    CASH_BALANCE_CENTS_FIELD_NUMBER: _ClassVar[int]
    RESERVED_CASH_CENTS_FIELD_NUMBER: _ClassVar[int]
    CURRENCY_FIELD_NUMBER: _ClassVar[int]
    MULTIPLIER_FIELD_NUMBER: _ClassVar[int]
    MARGIN_USED_CENTS_FIELD_NUMBER: _ClassVar[int]
    cash_balance_cents: int
    reserved_cash_cents: int
    currency: str
    multiplier: int
    margin_used_cents: int
    def __init__(self, cash_balance_cents: _Optional[int] = ..., reserved_cash_cents: _Optional[int] = ..., currency: _Optional[str] = ..., multiplier: _Optional[int] = ..., margin_used_cents: _Optional[int] = ...) -> None: ...

class Holding(_message.Message):
    __slots__ = ("symbol", "qty", "avg_price_cents", "cost_basis_cents", "last_price_cents", "market_value_cents", "unrealized_pnl_cents", "realized_pnl_cents")
    SYMBOL_FIELD_NUMBER: _ClassVar[int]
    QTY_FIELD_NUMBER: _ClassVar[int]
    AVG_PRICE_CENTS_FIELD_NUMBER: _ClassVar[int]
    COST_BASIS_CENTS_FIELD_NUMBER: _ClassVar[int]
    LAST_PRICE_CENTS_FIELD_NUMBER: _ClassVar[int]
    MARKET_VALUE_CENTS_FIELD_NUMBER: _ClassVar[int]
    UNREALIZED_PNL_CENTS_FIELD_NUMBER: _ClassVar[int]
    REALIZED_PNL_CENTS_FIELD_NUMBER: _ClassVar[int]
    symbol: str
    qty: int
    avg_price_cents: int
    cost_basis_cents: int
    last_price_cents: int
    market_value_cents: int
    unrealized_pnl_cents: int
    realized_pnl_cents: int
    def __init__(self, symbol: _Optional[str] = ..., qty: _Optional[int] = ..., avg_price_cents: _Optional[int] = ..., cost_basis_cents: _Optional[int] = ..., last_price_cents: _Optional[int] = ..., market_value_cents: _Optional[int] = ..., unrealized_pnl_cents: _Optional[int] = ..., realized_pnl_cents: _Optional[int] = ...) -> None: ...

class HoldingsResponse(_message.Message):
    __slots__ = ("holdings", "total_market_value_cents", "total_cost_basis_cents", "total_unrealized_pnl_cents", "total_realized_pnl_cents")
    HOLDINGS_FIELD_NUMBER: _ClassVar[int]
    TOTAL_MARKET_VALUE_CENTS_FIELD_NUMBER: _ClassVar[int]
    TOTAL_COST_BASIS_CENTS_FIELD_NUMBER: _ClassVar[int]
    TOTAL_UNREALIZED_PNL_CENTS_FIELD_NUMBER: _ClassVar[int]
    TOTAL_REALIZED_PNL_CENTS_FIELD_NUMBER: _ClassVar[int]
    holdings: _containers.RepeatedCompositeFieldContainer[Holding]
    total_market_value_cents: int
    total_cost_basis_cents: int
    total_unrealized_pnl_cents: int
    total_realized_pnl_cents: int
    def __init__(self, holdings: _Optional[_Iterable[_Union[Holding, _Mapping]]] = ..., total_market_value_cents: _Optional[int] = ..., total_cost_basis_cents: _Optional[int] = ..., total_unrealized_pnl_cents: _Optional[int] = ..., total_realized_pnl_cents: _Optional[int] = ...) -> None: ...

class Position(_message.Message):
    __slots__ = ("symbol", "qty", "reserved_qty", "realized_pnl_cents", "cost_basis_cents", "avg_price_cents")
    SYMBOL_FIELD_NUMBER: _ClassVar[int]
    QTY_FIELD_NUMBER: _ClassVar[int]
    RESERVED_QTY_FIELD_NUMBER: _ClassVar[int]
    REALIZED_PNL_CENTS_FIELD_NUMBER: _ClassVar[int]
    COST_BASIS_CENTS_FIELD_NUMBER: _ClassVar[int]
    AVG_PRICE_CENTS_FIELD_NUMBER: _ClassVar[int]
    symbol: str
    qty: int
    reserved_qty: int
    realized_pnl_cents: int
    cost_basis_cents: int
    avg_price_cents: int
    def __init__(self, symbol: _Optional[str] = ..., qty: _Optional[int] = ..., reserved_qty: _Optional[int] = ..., realized_pnl_cents: _Optional[int] = ..., cost_basis_cents: _Optional[int] = ..., avg_price_cents: _Optional[int] = ...) -> None: ...

class PositionsResponse(_message.Message):
    __slots__ = ("positions",)
    POSITIONS_FIELD_NUMBER: _ClassVar[int]
    positions: _containers.RepeatedCompositeFieldContainer[Position]
    def __init__(self, positions: _Optional[_Iterable[_Union[Position, _Mapping]]] = ...) -> None: ...
