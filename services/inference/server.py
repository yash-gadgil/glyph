import os
import signal
from concurrent import futures

import grpc
from inference import inference_pb2, inference_pb2_grpc

from model import Model


class InferenceServicer(inference_pb2_grpc.InferenceServiceServicer):
    def __init__(self, model):
        self.model = model

    def Generate(self, request, context):
        try:
            for chunk in self.model.generate(
                request.system,
                request.prompt,
                request.max_tokens,
                request.temperature,
            ):
                yield inference_pb2.Token(text=chunk)
            yield inference_pb2.Token(done=True)
        except Exception as exc:
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(exc))


def serve():
    addr = os.getenv("INFERENCE_BIND", "[::]:50057")
    model = Model()

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=8))
    inference_pb2_grpc.add_InferenceServiceServicer_to_server(
        InferenceServicer(model), server
    )
    server.add_insecure_port(addr)
    server.start()
    print(f"inference service listening on {addr}, model={model.model_id or 'fallback'}")

    signal.signal(signal.SIGTERM, lambda *_: server.stop(5))
    signal.signal(signal.SIGINT, lambda *_: server.stop(5))
    server.wait_for_termination()


if __name__ == "__main__":
    serve()
