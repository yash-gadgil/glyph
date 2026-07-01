# Inference

The inference service hosts the local language model the advisor talks to. It's a Python
gRPC service, the only non-Go, non-Rust service in the stack, managed with `uv`.

Path: `services/inference/` · Port: 50057 (`INFERENCE_BIND`) · Depends on: nothing at
runtime beyond its own model weights; it has no database and no external calls.

## Model hosting

The model is picked by the `MODEL_ID` environment variable (default
`Qwen/Qwen2.5-1.5B-Instruct`) and loaded once at startup with `transformers`, CPU only.
`pyproject.toml` pins torch to the CPU wheel index on Linux so the built image stays
small; weights are baked into the image (or a persistent volume) rather than downloaded on
every start. If `MODEL_ID` is unset, the service skips loading a model entirely and falls
back to an echo responder that streams the prompt back word by word, which is enough to
exercise the rest of the pipeline without a multi-minute model load.

## Generate

`Generate` takes a system prompt, a user prompt, `max_tokens`, and a `temperature`
(clamped to a minimum of 0.01), applies the model's chat template, and streams the
response back as a sequence of `Token` messages using `transformers`'
`TextIteratorStreamer` on a background thread. The stream ends with a token whose `done`
flag is set.

## Server

A plain `grpc.Server` with up to 8 concurrent streams, listening on `INFERENCE_BIND`
(default `[::]:50057`). SIGTERM and SIGINT trigger a graceful shutdown with a 5-second
grace period so an in-flight generation isn't cut off mid-stream.

## Endpoints

See [endpoints.md](endpoints.md) for the `InferenceService` RPC.
