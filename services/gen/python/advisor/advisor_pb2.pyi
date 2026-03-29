from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

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

class GenerateStrategyRequest(_message.Message):
    __slots__ = ("user_id",)
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    def __init__(self, user_id: _Optional[str] = ...) -> None: ...

class StrategySuggestion(_message.Message):
    __slots__ = ("name", "config_json", "rationale", "template")
    NAME_FIELD_NUMBER: _ClassVar[int]
    CONFIG_JSON_FIELD_NUMBER: _ClassVar[int]
    RATIONALE_FIELD_NUMBER: _ClassVar[int]
    TEMPLATE_FIELD_NUMBER: _ClassVar[int]
    name: str
    config_json: str
    rationale: str
    template: str
    def __init__(self, name: _Optional[str] = ..., config_json: _Optional[str] = ..., rationale: _Optional[str] = ..., template: _Optional[str] = ...) -> None: ...
