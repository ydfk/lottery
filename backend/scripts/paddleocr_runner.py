import json
import sys


def build_payload(raw_result):
    lines = []
    scores = []

    for page in raw_result or []:
        if not page:
            continue
        for item in page:
            if not isinstance(item, (list, tuple)) or len(item) < 2:
                continue
            text_info = item[1]
            if not isinstance(text_info, (list, tuple)) or len(text_info) < 2:
                continue

            text = str(text_info[0]).strip()
            if not text:
                continue

            lines.append(text)
            try:
                scores.append(float(text_info[1]))
            except (TypeError, ValueError):
                continue

    confidence = sum(scores) / len(scores) if scores else 0.0
    return {
        "rawText": "\n".join(lines),
        "confidence": confidence,
        "lines": lines,
    }


def main():
    if len(sys.argv) < 2:
        print(json.dumps({"error": "缺少图片路径"}, ensure_ascii=False))
        return 1

    image_path = sys.argv[1]
    lang = sys.argv[2] if len(sys.argv) > 2 and sys.argv[2] else "ch"
    use_angle_cls = True
    if len(sys.argv) > 3:
        use_angle_cls = sys.argv[3].lower() == "true"

    try:
        from paddleocr import PaddleOCR
    except Exception as exc:
        print(json.dumps({"error": f"未安装 PaddleOCR: {exc}"}, ensure_ascii=False))
        return 1

    try:
        ocr = PaddleOCR(use_angle_cls=use_angle_cls, lang=lang, show_log=False)
        result = ocr.ocr(image_path, cls=use_angle_cls)
        print(json.dumps(build_payload(result), ensure_ascii=False))
        return 0
    except Exception as exc:
        print(json.dumps({"error": str(exc)}, ensure_ascii=False))
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
