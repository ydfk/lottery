package lottery

import (
	"context"
	"encoding/json"
	"fmt"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/config"
)

type VisionRecognizer interface {
	Recognize(ctx context.Context, lotteryType model.LotteryType, imagePath string) (*RecognitionResult, error)
}

type paddleOCRVisionRecognizer struct{}

type openAIVisionRecognizer struct{}

func newVisionRecognizer(provider string) VisionRecognizer {
	if provider == "" || provider == ProviderPaddleOCR {
		return &paddleOCRVisionRecognizer{}
	}
	if provider == ProviderOpenAICompatible {
		return &openAIVisionRecognizer{}
	}
	return nil
}

func (recognizer *paddleOCRVisionRecognizer) Recognize(ctx context.Context, lotteryType model.LotteryType, imagePath string) (*RecognitionResult, error) {
	output, err := callPaddleOCR(ctx, imagePath)
	if err != nil {
		return nil, err
	}

	payload, err := parsePaddleOCRPayload(output)
	if err != nil {
		return nil, err
	}

	return buildRecognitionFromOCRPayload(lotteryType.Code, payload)
}

func (recognizer *openAIVisionRecognizer) Recognize(ctx context.Context, lotteryType model.LotteryType, imagePath string) (*RecognitionResult, error) {
	prompt := config.Current.Vision.Prompt
	if prompt == "" {
		prompt = fmt.Sprintf(
			"识别图片中的%s彩票，只返回 JSON，格式为 {\"issue\":\"\",\"rawText\":\"\",\"confidence\":0.0,\"entries\":[{\"red\":[1,2,3,4,5,6],\"blue\":[7]}]}。",
			lotteryType.Name,
		)
	}

	content, err := callVisionModel(ctx, config.Current.Vision.Model, prompt, imagePath)
	if err != nil {
		return nil, err
	}

	result := RecognitionResult{}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, err
	}
	if len(result.Entries) == 0 {
		return nil, fmt.Errorf("未从图片中识别到有效号码")
	}
	return &result, nil
}
