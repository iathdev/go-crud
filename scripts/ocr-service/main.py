"""
Multi-engine OCR HTTP service.

Supported engines:
  - paddleocr (default) — PaddleOCR PP-OCRv5
  - tesseract           — Tesseract OCR (requires: brew install tesseract tesseract-lang && pip install pytesseract)

Input: JSON with base64-encoded image (same format as Google Vision / Baidu OCR APIs).
"""

import base64
import os
import re
import tempfile

os.environ["PADDLE_PDX_DISABLE_MODEL_SOURCE_CHECK"] = "True"

from fastapi import FastAPI, HTTPException
from paddleocr import PaddleOCR
from pydantic import BaseModel

app = FastAPI(title="OCR Service", description="Multi-engine OCR wrapper for Chinese character extraction")

# PaddleOCR engines, lazily loaded per language
_paddle_engines: dict[str, PaddleOCR] = {}

CJK_PATTERN = re.compile(r"[\u4e00-\u9fff]+")

# API language code → PaddleOCR language code
LANG_MAP = {"zh": "ch", "en": "en", "vi": "vi"}


class ExtractRequest(BaseModel):
    image: str  # base64-encoded image
    language: str = "zh"
    engine: str = "paddleocr"


def _get_paddle_engine(language: str) -> PaddleOCR:
    paddle_lang = LANG_MAP.get(language, language)
    if paddle_lang not in _paddle_engines:
        _paddle_engines[paddle_lang] = PaddleOCR(lang=paddle_lang, use_textline_orientation=True)
    return _paddle_engines[paddle_lang]


def _extract_paddleocr(image_path: str, language: str) -> list[dict]:
    engine = _get_paddle_engine(language)
    results = engine.predict(image_path)

    characters = []
    for result in results:
        texts = result.get("rec_texts", [])
        scores = result.get("rec_scores", [])

        for text, score in zip(texts, scores):
            confidence = float(score)

            if language == "zh":
                cjk_matches = CJK_PATTERN.findall(text)
                text = "".join(cjk_matches)
                if not text:
                    continue

            characters.append({"text": text, "confidence": round(confidence, 4), "candidates": []})

    return characters


def _extract_tesseract(image_path: str, language: str) -> list[dict]:
    try:
        import pytesseract
    except ImportError:
        raise HTTPException(
            status_code=400,
            detail="Tesseract engine not available. Install: pip install pytesseract && brew install tesseract tesseract-lang",
        )

    tess_lang_map = {"zh": "chi_sim", "en": "eng", "vi": "vie"}
    tess_lang = tess_lang_map.get(language, "eng")

    from PIL import Image

    img = Image.open(image_path)
    data = pytesseract.image_to_data(img, lang=tess_lang, output_type=pytesseract.Output.DICT)

    characters = []
    for i, text in enumerate(data["text"]):
        text = text.strip()
        if not text:
            continue

        conf = data["conf"][i]
        confidence = max(0.0, float(conf) / 100.0) if conf != -1 else 0.0

        if language == "zh":
            cjk_matches = CJK_PATTERN.findall(text)
            text = "".join(cjk_matches)
            if not text:
                continue

        characters.append({"text": text, "confidence": round(confidence, 4), "candidates": []})

    return characters


ENGINES = {
    "paddleocr": _extract_paddleocr,
    "tesseract": _extract_tesseract,
}


@app.post("/recognize")
async def extract(req: ExtractRequest):
    if req.engine not in ENGINES:
        raise HTTPException(status_code=400, detail=f"Unknown engine: {req.engine}. Available: {list(ENGINES.keys())}")

    try:
        image_bytes = base64.b64decode(req.image)
    except Exception:
        raise HTTPException(status_code=400, detail="Invalid base64 image data")

    with tempfile.NamedTemporaryFile(suffix=".png", delete=True) as tmp:
        tmp.write(image_bytes)
        tmp.flush()

        characters = ENGINES[req.engine](tmp.name, req.language)

    return {"characters": characters, "engine": req.engine}


@app.get("/health")
async def health():
    return {"status": "ok", "engines": list(ENGINES.keys())}
