package lottery

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"go-fiber-starter/pkg/config"
)

type paddleOCRPayload struct {
	RawText    string   `json:"rawText"`
	Confidence float64  `json:"confidence"`
	Lines      []string `json:"lines"`
	Error      string   `json:"error"`
}

func callPaddleOCR(ctx context.Context, imagePath string) ([]byte, error) {
	command := resolveValue(config.Current.Vision.Command, "python")
	args := append([]string(nil), config.Current.Vision.Args...)
	if len(args) == 0 {
		args = []string{"scripts/paddleocr_runner.py"}
	}

	lang := resolveValue(config.Current.Vision.Lang, "ch")
	args = append(args, imagePath, lang, strconv.FormatBool(config.Current.Vision.UseAngleCls))

	timeout := time.Duration(max(10, config.Current.Vision.TimeoutSeconds)) * time.Second
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	commandExec := exec.CommandContext(execCtx, command, args...)
	output, err := commandExec.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return nil, fmt.Errorf("PaddleOCR 执行失败: %s", message)
	}

	return output, nil
}

func parsePaddleOCRPayload(output []byte) (*paddleOCRPayload, error) {
	payload := paddleOCRPayload{}
	if err := json.Unmarshal(output, &payload); err != nil {
		return nil, fmt.Errorf("PaddleOCR 输出格式无法识别: %s", strings.TrimSpace(string(output)))
	}

	if payload.Error != "" {
		return nil, fmt.Errorf("PaddleOCR 执行失败: %s", payload.Error)
	}
	if strings.TrimSpace(payload.RawText) == "" && len(payload.Lines) > 0 {
		payload.RawText = strings.Join(payload.Lines, "\n")
	}
	return &payload, nil
}

func buildRecognitionFromOCRPayload(lotteryTypeCode string, payload *paddleOCRPayload) (*RecognitionResult, error) {
	rawText := strings.TrimSpace(payload.RawText)
	if rawText == "" {
		return nil, fmt.Errorf("未从图片中识别到文本")
	}

	if lotteryTypeCode == "ssq" {
		recognized, err := ParseSSQText(rawText)
		if err == nil {
			if payload.Confidence > 0 {
				recognized.Confidence = payload.Confidence
			}
			return recognized, nil
		}
	}

	return &RecognitionResult{
		RawText:    rawText,
		Confidence: payload.Confidence,
	}, nil
}
