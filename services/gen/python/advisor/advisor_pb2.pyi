from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class AnalyzeRequest(_message.Message):
    __slots__ = ("user_id",)
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    def __init__(self, user_id: _Optional[str] = ...) -> None: ...

class AnalysisChunk(_message.Message):
    __slots__ = ("text", "done")
    TEXT_FIELD_NUMBER: _ClassVar[int]
    DONE_FIELD_NUMBER: _ClassVar[int]
    text: str
    done: bool
    def __init__(self, text: _Optional[str] = ..., done: bool = ...) -> None: ...

class StartStrategyGenerationRequest(_message.Message):
    __slots__ = ("user_id",)
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    def __init__(self, user_id: _Optional[str] = ...) -> None: ...

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
