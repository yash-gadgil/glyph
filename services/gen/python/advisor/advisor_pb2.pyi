from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class AnalysisChunk(_message.Message):
    __slots__ = ("text", "done")
    TEXT_FIELD_NUMBER: _ClassVar[int]
    DONE_FIELD_NUMBER: _ClassVar[int]
    text: str
    done: bool
    def __init__(self, text: _Optional[str] = ..., done: bool = ...) -> None: ...

class ChatRequest(_message.Message):
    __slots__ = ("user_id", "message")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    message: str
    def __init__(self, user_id: _Optional[str] = ..., message: _Optional[str] = ...) -> None: ...

class GetChatSessionRequest(_message.Message):
    __slots__ = ("user_id",)
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    def __init__(self, user_id: _Optional[str] = ...) -> None: ...

class ChatTurn(_message.Message):
    __slots__ = ("role", "content")
    ROLE_FIELD_NUMBER: _ClassVar[int]
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    role: str
    content: str
    def __init__(self, role: _Optional[str] = ..., content: _Optional[str] = ...) -> None: ...

class ChatSession(_message.Message):
    __slots__ = ("turns", "in_flight", "partial_text")
    TURNS_FIELD_NUMBER: _ClassVar[int]
    IN_FLIGHT_FIELD_NUMBER: _ClassVar[int]
    PARTIAL_TEXT_FIELD_NUMBER: _ClassVar[int]
    turns: _containers.RepeatedCompositeFieldContainer[ChatTurn]
    in_flight: bool
    partial_text: str
    def __init__(self, turns: _Optional[_Iterable[_Union[ChatTurn, _Mapping]]] = ..., in_flight: bool = ..., partial_text: _Optional[str] = ...) -> None: ...

class StartStrategyGenerationRequest(_message.Message):
    __slots__ = ("user_id", "symbol")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    SYMBOL_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    symbol: str
    def __init__(self, user_id: _Optional[str] = ..., symbol: _Optional[str] = ...) -> None: ...

class GetStrategyJobRequest(_message.Message):
    __slots__ = ("user_id",)
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    def __init__(self, user_id: _Optional[str] = ...) -> None: ...

class BacktestSummary(_message.Message):
    __slots__ = ("total_return_pct", "max_drawdown_pct", "sharpe", "win_rate", "profit_factor", "num_trades")
    TOTAL_RETURN_PCT_FIELD_NUMBER: _ClassVar[int]
    MAX_DRAWDOWN_PCT_FIELD_NUMBER: _ClassVar[int]
    SHARPE_FIELD_NUMBER: _ClassVar[int]
    WIN_RATE_FIELD_NUMBER: _ClassVar[int]
    PROFIT_FACTOR_FIELD_NUMBER: _ClassVar[int]
    NUM_TRADES_FIELD_NUMBER: _ClassVar[int]
    total_return_pct: float
    max_drawdown_pct: float
    sharpe: float
    win_rate: float
    profit_factor: float
    num_trades: int
    def __init__(self, total_return_pct: _Optional[float] = ..., max_drawdown_pct: _Optional[float] = ..., sharpe: _Optional[float] = ..., win_rate: _Optional[float] = ..., profit_factor: _Optional[float] = ..., num_trades: _Optional[int] = ...) -> None: ...

class StrategyJob(_message.Message):
    __slots__ = ("state", "name", "config_json", "rationale", "backtest", "error", "started_at", "updated_at")
    STATE_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    CONFIG_JSON_FIELD_NUMBER: _ClassVar[int]
    RATIONALE_FIELD_NUMBER: _ClassVar[int]
    BACKTEST_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    STARTED_AT_FIELD_NUMBER: _ClassVar[int]
    UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    state: str
    name: str
    config_json: str
    rationale: str
    backtest: BacktestSummary
    error: str
    started_at: str
    updated_at: str
    def __init__(self, state: _Optional[str] = ..., name: _Optional[str] = ..., config_json: _Optional[str] = ..., rationale: _Optional[str] = ..., backtest: _Optional[_Union[BacktestSummary, _Mapping]] = ..., error: _Optional[str] = ..., started_at: _Optional[str] = ..., updated_at: _Optional[str] = ...) -> None: ...
