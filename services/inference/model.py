import importlib
import os
import threading
from typing import Any, Iterator


class Model:
    def __init__(self) -> None:
        self.model_id = os.getenv("MODEL_ID", "")
        self._tokenizer: Any = None
        self._model: Any = None
        self._streamer_cls: Any = None
        if self.model_id:
            self._load()

    def _load(self) -> None:
        transformers = importlib.import_module("transformers")
        self._tokenizer = transformers.AutoTokenizer.from_pretrained(self.model_id)
        self._model = transformers.AutoModelForCausalLM.from_pretrained(
            self.model_id, torch_dtype="auto", low_cpu_mem_usage=True
        )
        self._streamer_cls = transformers.TextIteratorStreamer

    def generate(
        self, system: str, prompt: str, max_tokens: int, temperature: float
    ) -> Iterator[str]:
        if self._tokenizer is None or self._model is None:
            yield from self._fallback(prompt)
            return

        messages = [
            {"role": "system", "content": system},
            {"role": "user", "content": prompt},
        ]
        inputs = self._tokenizer.apply_chat_template(
            messages, add_generation_prompt=True, return_tensors="pt", return_dict=True
        )
        streamer = self._streamer_cls(
            self._tokenizer, skip_prompt=True, skip_special_tokens=True
        )
        kwargs = {
            **inputs,
            "streamer": streamer,
            "max_new_tokens": max_tokens or 512,
            "do_sample": temperature > 0,
            "temperature": max(temperature, 0.01),
            "pad_token_id": self._tokenizer.eos_token_id,
        }
        thread = threading.Thread(target=self._model.generate, kwargs=kwargs)
        thread.start()
        for chunk in streamer:
            if chunk:
                yield chunk
        thread.join()

    def _fallback(self, prompt: str) -> Iterator[str]:
        lines = [
            "Portfolio review (offline model placeholder).",
            "",
            "Set MODEL_ID to load a real model. Until then this echoes the context it was given:",
            "",
            prompt,
        ]
        for word in "\n".join(lines).split(" "):
            yield word + " "
