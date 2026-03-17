package lottery

import "testing"

func TestParsePaddleOCRPayload(t *testing.T) {
	payload, err := parsePaddleOCRPayload([]byte(`{
		"result": {
			"layoutParsingResults": [
				{"markdown": {"text": "2026001\n01 02 03 04 05 06 07"}}
			]
		}
	}`))
	if err != nil {
		t.Fatalf("parse payload: %v", err)
	}
	if payload.RawText == "" {
		t.Fatalf("expected raw text")
	}
	if len(payload.Lines) != 1 {
		t.Fatalf("unexpected lines: %d", len(payload.Lines))
	}
}

func TestBuildRecognitionFromOCRPayloadForSSQ(t *testing.T) {
	recognized, err := buildRecognitionFromOCRPayload("ssq", &paddleOCRPayload{
		RawText:    "2026001\n01 02 03 04 05 06 07",
		Confidence: 0.88,
	})
	if err != nil {
		t.Fatalf("build recognition: %v", err)
	}
	if recognized.Issue != "2026001" {
		t.Fatalf("unexpected issue: %s", recognized.Issue)
	}
	if len(recognized.Entries) != 1 {
		t.Fatalf("unexpected entries: %d", len(recognized.Entries))
	}
	if recognized.Confidence != 0.88 {
		t.Fatalf("unexpected confidence: %v", recognized.Confidence)
	}
}

func TestDetectPaddleOCRFileType(t *testing.T) {
	if fileType := detectPaddleOCRFileType("ticket.jpg"); fileType != 1 {
		t.Fatalf("unexpected image file type: %d", fileType)
	}
	if fileType := detectPaddleOCRFileType("ticket.pdf"); fileType != 0 {
		t.Fatalf("unexpected pdf file type: %d", fileType)
	}
}
