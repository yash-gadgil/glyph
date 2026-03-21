from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class GenerateRequest(_message.Message):
    __slots__ = ("system", "prompt", "max_tokens", "temperature")
    SYSTEM_FIELD_NUMBER: _ClassVar[int]
    PROMPT_FIELD_NUMBER: _ClassVar[int]
    MAX_TOKENS_FIELD_NUMBER: _ClassVar[int]
    TEMPERATURE_FIELD_NUMBER: _ClassVar[int]
    system: str
    prompt: str
    max_tokens: int
    temperature: float
    def __init__(self, system: _Optional[str] = ..., prompt: _Optional[str] = ..., max_tokens: _Optional[int] = ..., temperature: _Optional[float] = ...) -> None: ...

class Token(_message.Message):
    __slots__ = ("text", "done")
    TEXT_FIELD_NUMBER: _ClassVar[int]
    DONE_FIELD_NUMBER: _ClassVar[int]
    text: str
    done: bool
    def __init__(self, text: _Optional[str] = ..., done: bool = ...) -> None: ...
