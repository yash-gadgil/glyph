# Inference endpoints

The inference service exposes one gRPC service, `InferenceService`, defined in
`proto/inference/inference.proto`. Only the advisor service calls it.

| RPC | Request to response | Description |
|-----|---------------------|-------------|
| `Generate` | `GenerateRequest` to stream `Token` | stream a completion for a system/user prompt pair |
