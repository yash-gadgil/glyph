# Advisor endpoints

The advisor service exposes one gRPC service, `AdvisorService`, defined in
`proto/advisor/advisor.proto`. It is internal: only the gateway calls it.

| RPC | Request to response | Description |
|-----|---------------------|-------------|
| `ChatWithAdvisor` | `ChatRequest` to stream `AnalysisChunk` | send a message, stream the agent's reply token by token |
| `GetChatSession` | `GetChatSessionRequest` to `ChatSession` | the current turn history and in-flight status |
| `ClearChatSession` | `GetChatSessionRequest` to `ChatSession` | wipe the session and start fresh |
| `StartStrategyGeneration` | `StartStrategyGenerationRequest` to `StrategyJob` | kick off async strategy generation for a symbol |
| `GetStrategyJob` | `GetStrategyJobRequest` to `StrategyJob` | poll the generation job's state, result, or error |

`ChatRequest` and `StartStrategyGenerationRequest` both carry a `provider` field
(`"inference"` or `"gemini"`) that picks which model answers that request.
