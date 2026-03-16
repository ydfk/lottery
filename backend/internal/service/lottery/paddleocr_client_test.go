package lottery

import "testing"

func TestParsePaddleOCRPayload(t *testing.T) {
	payload, err := parsePaddleOCRPayload([]byte(`{"rawText":"2026001\n01 02 03 04 05 06 07","confidence":0.91}`))
	if err != nil {
		t.Fatalf("parse payload: %v", err)
	}
	if payload.RawText == "" {
		t.Fatalf("expected raw text")
	}
	if payload.Confidence != 0.91 {
		t.Fatalf("unexpected confidence: %v", payload.Confidence)
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
