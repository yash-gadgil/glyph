from google.protobuf import empty_pb2 as _empty_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class SignupRequest(_message.Message):
    __slots__ = ("user_name", "email", "password")
    USER_NAME_FIELD_NUMBER: _ClassVar[int]
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    PASSWORD_FIELD_NUMBER: _ClassVar[int]
    user_name: str
    email: str
    password: str
    def __init__(self, user_name: _Optional[str] = ..., email: _Optional[str] = ..., password: _Optional[str] = ...) -> None: ...

class SigninRequest(_message.Message):
    __slots__ = ("email", "password")
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    PASSWORD_FIELD_NUMBER: _ClassVar[int]
    email: str
    password: str
    def __init__(self, email: _Optional[str] = ..., password: _Optional[str] = ...) -> None: ...

class TokenResponse(_message.Message):
    __slots__ = ("access_token", "refresh_token")
    ACCESS_TOKEN_FIELD_NUMBER: _ClassVar[int]
    REFRESH_TOKEN_FIELD_NUMBER: _ClassVar[int]
    access_token: str
    refresh_token: str
    def __init__(self, access_token: _Optional[str] = ..., refresh_token: _Optional[str] = ...) -> None: ...

class OAuthURLRequest(_message.Message):
    __slots__ = ("provider", "state")
    PROVIDER_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    provider: str
    state: str
    def __init__(self, provider: _Optional[str] = ..., state: _Optional[str] = ...) -> None: ...

class OAuthURLResponse(_message.Message):
    __slots__ = ("url",)
    URL_FIELD_NUMBER: _ClassVar[int]
    url: str
    def __init__(self, url: _Optional[str] = ...) -> None: ...

class OAuthCallbackRequest(_message.Message):
    __slots__ = ("provider", "code", "state")
    PROVIDER_FIELD_NUMBER: _ClassVar[int]
    CODE_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    provider: str
    code: str
    state: str
    def __init__(self, provider: _Optional[str] = ..., code: _Optional[str] = ..., state: _Optional[str] = ...) -> None: ...

class VerificationRequest(_message.Message):
    __slots__ = ("token",)
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    token: str
    def __init__(self, token: _Optional[str] = ...) -> None: ...

class VerificationResponse(_message.Message):
    __slots__ = ("user_id",)
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    def __init__(self, user_id: _Optional[str] = ...) -> None: ...

class RefreshTokenRequest(_message.Message):
    __slots__ = ("refresh_token",)
    REFRESH_TOKEN_FIELD_NUMBER: _ClassVar[int]
    refresh_token: str
    def __init__(self, refresh_token: _Optional[str] = ...) -> None: ...

class PublicKey(_message.Message):
    __slots__ = ("kid", "kty", "use", "alg", "n", "e")
    KID_FIELD_NUMBER: _ClassVar[int]
    KTY_FIELD_NUMBER: _ClassVar[int]
    USE_FIELD_NUMBER: _ClassVar[int]
    ALG_FIELD_NUMBER: _ClassVar[int]
    N_FIELD_NUMBER: _ClassVar[int]
    E_FIELD_NUMBER: _ClassVar[int]
    kid: str
    kty: str
    use: str
    alg: str
    n: str
    e: str
    def __init__(self, kid: _Optional[str] = ..., kty: _Optional[str] = ..., use: _Optional[str] = ..., alg: _Optional[str] = ..., n: _Optional[str] = ..., e: _Optional[str] = ...) -> None: ...

class GetPublicKeysResponse(_message.Message):
    __slots__ = ("keys",)
    KEYS_FIELD_NUMBER: _ClassVar[int]
    keys: _containers.RepeatedCompositeFieldContainer[PublicKey]
    def __init__(self, keys: _Optional[_Iterable[_Union[PublicKey, _Mapping]]] = ...) -> None: ...

class ForgotPasswordRequest(_message.Message):
    __slots__ = ("email",)
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    email: str
    def __init__(self, email: _Optional[str] = ...) -> None: ...

class ResetPasswordRequest(_message.Message):
    __slots__ = ("token", "new_password")
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    NEW_PASSWORD_FIELD_NUMBER: _ClassVar[int]
    token: str
    new_password: str
    def __init__(self, token: _Optional[str] = ..., new_password: _Optional[str] = ...) -> None: ...
