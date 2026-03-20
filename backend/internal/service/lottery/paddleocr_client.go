package lottery

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-fiber-starter/pkg/config"

	_ "image/gif"
	_ "image/png"

	xdraw "golang.org/x/image/draw"
)

type paddleOCRPayload struct {
	RawText    string   `json:"rawText"`
	Confidence float64  `json:"confidence"`
	Lines      []string `json:"lines"`
	Error      string   `json:"error"`
}

type paddleOCRRequest struct {
	File                      string `json:"file"`
	FileType                  int    `json:"fileType"`
	UseDocOrientationClassify bool   `json:"useDocOrientationClassify"`
	UseDocUnwarping           bool   `json:"useDocUnwarping"`
	UseChartRecognition       bool   `json:"useChartRecognition"`
}

type paddleOCRResponse struct {
	Result paddleOCRResult `json:"result"`
	Error  string          `json:"error"`
	Msg    string          `json:"msg"`
}

type paddleOCRResult struct {
	LayoutParsingResults []paddleOCRLayoutParsingResult `json:"layoutParsingResults"`
}

type paddleOCRLayoutParsingResult struct {
	Markdown paddleOCRMarkdown `json:"markdown"`
}

type paddleOCRMarkdown struct {
	Text string `json:"text"`
}

type paddleOCRImageMeta struct {
	OriginalSize    int    `json:"originalSize"`
	ProcessedSize   int    `json:"processedSize"`
	OriginalWidth   int    `json:"originalWidth"`
	OriginalHeight  int    `json:"originalHeight"`
	ProcessedWidth  int    `json:"processedWidth"`
	ProcessedHeight int    `json:"processedHeight"`
	WasCompressed   bool   `json:"wasCompressed"`
	UsedBinarized   bool   `json:"usedBinarized"`
	Attempt         int    `json:"attempt"`
	AttemptLabel    string `json:"attemptLabel"`
}

const (
	paddleOCRDefaultJPEGQuality = 82
)

type paddleOCRAttemptPlan struct {
	Label     string
	MaxSide   int
	Quality   int
	Binarized bool
	UseOrigin bool
}

type paddleOCRImageSource struct {
	FileBytes []byte
	Decoded   image.Image
	Width     int
	Height    int
}

var paddleOCRAttemptPlans = []paddleOCRAttemptPlan{
	{Label: "compressed-large", MaxSide: 1600, Quality: 80, Binarized: false},
	{Label: "binarized-small", MaxSide: 960, Quality: 70, Binarized: true},
}

var paddleOCRWorkerSlots = make(chan struct{}, 1)

func callPaddleOCR(ctx context.Context, imagePath string) ([]byte, error) {
	if err := acquirePaddleOCRWorkerSlot(ctx); err != nil {
		return nil, err
	}
	defer releasePaddleOCRWorkerSlot()

	baseURL := strings.TrimSpace(config.Current.Vision.BaseURL)
	if baseURL == "" {
		return nil, fmt.Errorf("未配置 PaddleOCR 服务地址")
	}

	fileBytes, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("读取待识别文件失败: %w", err)
	}
	fileType := detectPaddleOCRFileType(imagePath)
	source := buildPaddleOCRImageSource(fileType, fileBytes)
	plans := buildPaddleOCRAttemptPlans(fileType, source)
	perAttemptTimeout := time.Duration(max(20, max(10, config.Current.Vision.TimeoutSeconds)/len(plans))) * time.Second

	var lastErr error
	for index, plan := range plans {
		attemptStartedAt := time.Now()
		ocrBytes, imageMeta := preparePaddleOCRFile(source, fileType, plan, index+1)
		requestPayload := paddleOCRRequest{
			File:                      base64.StdEncoding.EncodeToString(ocrBytes),
			FileType:                  fileType,
			UseDocOrientationClassify: config.Current.Vision.UseDocOrientationClassify,
			UseDocUnwarping:           config.Current.Vision.UseDocUnwarping,
			UseChartRecognition:       config.Current.Vision.UseChartRecognition,
		}

		body, statusCode, err := doPaddleOCRRequest(ctx, baseURL, requestPayload, perAttemptTimeout)
		requestLog := buildPaddleOCRRequestLog(imagePath, requestPayload, imageMeta)
		if err == nil {
			logThirdPartySuccess("paddleocr", http.MethodPost, baseURL, requestLog, body, statusCode, attemptStartedAt)
			return body, nil
		}

		lastErr = err
		logThirdPartyFailure("paddleocr", http.MethodPost, baseURL, requestLog, body, statusCode, attemptStartedAt, err)
		if !shouldRetryPaddleOCR(statusCode, err) {
			break
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("PaddleOCR 服务未返回结果")
	}
	return nil, lastErr
}

func parsePaddleOCRPayload(output []byte) (*paddleOCRPayload, error) {
	response := paddleOCRResponse{}
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("PaddleOCR 输出格式无法识别: %s", strings.TrimSpace(string(output)))
	}

	if response.Error != "" {
		return nil, fmt.Errorf("PaddleOCR 执行失败: %s", response.Error)
	}
	if response.Msg != "" && len(response.Result.LayoutParsingResults) == 0 {
		return nil, fmt.Errorf("PaddleOCR 执行失败: %s", response.Msg)
	}

	lines := make([]string, 0, len(response.Result.LayoutParsingResults))
	for _, item := range response.Result.LayoutParsingResults {
		text := strings.TrimSpace(item.Markdown.Text)
		if text == "" {
			continue
		}
		lines = append(lines, text)
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("未从 PaddleOCR 返回结果中提取到文本")
	}

	return &paddleOCRPayload{
		RawText: strings.Join(lines, "\n"),
		Lines:   lines,
	}, nil
}

func buildRecognitionFromOCRPayload(lotteryTypeCode string, payload *paddleOCRPayload) (*RecognitionResult, error) {
	rawText := strings.TrimSpace(payload.RawText)
	if rawText == "" {
		return nil, fmt.Errorf("未从图片中识别到文本")
	}

	if lotteryTypeCode != "" {
		recognized, err := ParseLotteryText(lotteryTypeCode, rawText)
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

func detectPaddleOCRFileType(filePath string) int {
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".pdf":
		return 0
	default:
		return 1
	}
}

func buildPaddleOCRRequestLog(imagePath string, payload paddleOCRRequest, imageMeta paddleOCRImageMeta) map[string]any {
	return map[string]any{
		"imagePath":                 imagePath,
		"attempt":                   imageMeta.Attempt,
		"attemptLabel":              imageMeta.AttemptLabel,
		"fileType":                  payload.FileType,
		"originalFileSize":          imageMeta.OriginalSize,
		"processedFileSize":         imageMeta.ProcessedSize,
		"fileBase64Length":          len(payload.File),
		"originalWidth":             imageMeta.OriginalWidth,
		"originalHeight":            imageMeta.OriginalHeight,
		"processedWidth":            imageMeta.ProcessedWidth,
		"processedHeight":           imageMeta.ProcessedHeight,
		"wasCompressed":             imageMeta.WasCompressed,
		"usedBinarized":             imageMeta.UsedBinarized,
		"useDocOrientationClassify": payload.UseDocOrientationClassify,
		"useDocUnwarping":           payload.UseDocUnwarping,
		"useChartRecognition":       payload.UseChartRecognition,
	}
}

func preparePaddleOCRFile(source paddleOCRImageSource, fileType int, plan paddleOCRAttemptPlan, attempt int) ([]byte, paddleOCRImageMeta) {
	meta := paddleOCRImageMeta{
		OriginalSize:  len(source.FileBytes),
		ProcessedSize: len(source.FileBytes),
		Attempt:       attempt,
		AttemptLabel:  plan.Label,
	}
	if fileType == 0 {
		return source.FileBytes, meta
	}
	meta.OriginalWidth = source.Width
	meta.OriginalHeight = source.Height
	if plan.UseOrigin || source.Decoded == nil {
		meta.ProcessedWidth = source.Width
		meta.ProcessedHeight = source.Height
		return source.FileBytes, meta
	}
	targetWidth, targetHeight := fitImageSize(meta.OriginalWidth, meta.OriginalHeight, plan.MaxSide)
	meta.ProcessedWidth = targetWidth
	meta.ProcessedHeight = targetHeight

	canvas := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	fillWhiteBackground(canvas)
	bounds := source.Decoded.Bounds()
	if targetWidth != meta.OriginalWidth || targetHeight != meta.OriginalHeight {
		xdraw.ApproxBiLinear.Scale(canvas, canvas.Bounds(), source.Decoded, bounds, xdraw.Over, nil)
	} else {
		xdraw.Draw(canvas, canvas.Bounds(), source.Decoded, bounds.Min, xdraw.Over)
	}

	if plan.Binarized {
		applyBinaryThreshold(canvas)
		meta.UsedBinarized = true
	}

	buffer := bytes.NewBuffer(nil)
	if err := jpeg.Encode(buffer, canvas, &jpeg.Options{Quality: max(40, plan.Quality)}); err != nil {
		return source.FileBytes, meta
	}

	processedBytes := buffer.Bytes()
	meta.ProcessedSize = len(processedBytes)
	meta.WasCompressed = len(processedBytes) != len(source.FileBytes) || targetWidth != meta.OriginalWidth || targetHeight != meta.OriginalHeight
	return processedBytes, meta
}

func buildPaddleOCRAttemptPlans(fileType int, source paddleOCRImageSource) []paddleOCRAttemptPlan {
	if fileType == 0 {
		return []paddleOCRAttemptPlan{{
			Label:     "pdf-original",
			MaxSide:   0,
			Quality:   paddleOCRDefaultJPEGQuality,
			Binarized: false,
			UseOrigin: true,
		}}
	}

	plans := make([]paddleOCRAttemptPlan, 0, len(paddleOCRAttemptPlans)+1)
	if source.Width > 0 && source.Height > 0 && source.Width <= 1600 && source.Height <= 1600 && len(source.FileBytes) <= 900*1024 {
		plans = append(plans, paddleOCRAttemptPlan{
			Label:     "original-direct",
			MaxSide:   0,
			Quality:   paddleOCRDefaultJPEGQuality,
			Binarized: false,
			UseOrigin: true,
		})
	}
	return append(plans, paddleOCRAttemptPlans...)
}

func fitImageSize(width int, height int, maxSide int) (int, int) {
	if width <= 0 || height <= 0 {
		return width, height
	}
	if width <= maxSide && height <= maxSide {
		return width, height
	}

	if width >= height {
		return maxSide, max(1, height*maxSide/width)
	}
	return max(1, width*maxSide/height), maxSide
}

func fillWhiteBackground(target *image.RGBA) {
	xdraw.Draw(target, target.Bounds(), &image.Uniform{C: color.White}, image.Point{}, xdraw.Src)
}

func applyBinaryThreshold(target *image.RGBA) {
	bounds := target.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			current := target.RGBAAt(x, y)
			gray := uint8((299*uint16(current.R) + 587*uint16(current.G) + 114*uint16(current.B)) / 1000)
			if gray > 180 {
				target.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
				continue
			}
			target.SetRGBA(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 255})
		}
	}
}

func doPaddleOCRRequest(ctx context.Context, baseURL string, requestPayload paddleOCRRequest, timeout time.Duration) ([]byte, int, error) {
	requestBody, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, 0, fmt.Errorf("构造 PaddleOCR 请求失败: %w", err)
	}

	requestCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	request, err := http.NewRequestWithContext(requestCtx, http.MethodPost, baseURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, 0, fmt.Errorf("创建 PaddleOCR 请求失败: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	if token := strings.TrimSpace(config.Current.Vision.APIKey); token != "" {
		request.Header.Set("Authorization", "token "+token)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, 0, fmt.Errorf("调用 PaddleOCR 服务失败: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, response.StatusCode, fmt.Errorf("读取 PaddleOCR 响应失败: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		message := strings.TrimSpace(string(body))
		if message == "" {
			message = response.Status
		}
		return body, response.StatusCode, fmt.Errorf("PaddleOCR 服务返回异常: %s", message)
	}
	return body, response.StatusCode, nil
}

func shouldRetryPaddleOCR(statusCode int, err error) bool {
	if err == nil {
		return false
	}
	if statusCode == 0 {
		return true
	}
	return statusCode >= http.StatusInternalServerError
}

func buildPaddleOCRImageSource(fileType int, fileBytes []byte) paddleOCRImageSource {
	source := paddleOCRImageSource{
		FileBytes: fileBytes,
	}
	if fileType == 0 {
		return source
	}

	decoded, _, err := image.Decode(bytes.NewReader(fileBytes))
	if err != nil {
		return source
	}

	bounds := decoded.Bounds()
	source.Decoded = decoded
	source.Width = bounds.Dx()
	source.Height = bounds.Dy()
	return source
}

func acquirePaddleOCRWorkerSlot(ctx context.Context) error {
	select {
	case paddleOCRWorkerSlots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("等待 OCR 处理资源失败: %w", ctx.Err())
	}
}

func releasePaddleOCRWorkerSlot() {
	select {
	case <-paddleOCRWorkerSlots:
	default:
	}
}
