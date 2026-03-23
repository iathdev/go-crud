package service

import (
	"context"
	"unicode"

	vision "cloud.google.com/go/vision/v2/apiv1"
	visionpb "cloud.google.com/go/vision/v2/apiv1/visionpb"
	"google.golang.org/api/option"

	"learning-go/internal/infrastructure/circuitbreaker"
	apperr "learning-go/internal/shared/error"
	"learning-go/internal/shared/logger"
	"learning-go/internal/vocabulary/application/port"

	"go.uber.org/zap"
)

type GoogleVisionService struct {
	client  *vision.ImageAnnotatorClient
	breaker *circuitbreaker.Breaker
}

func NewGoogleVisionService(credFile string, breaker *circuitbreaker.Breaker) (port.OCRServicePort, func(), error) {
	ctx := context.Background()
	client, err := vision.NewImageAnnotatorClient(ctx,
		option.WithCredentialsFile(credFile),
	)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { client.Close() }
	return &GoogleVisionService{client: client, breaker: breaker}, cleanup, nil
}

func (svc *GoogleVisionService) Recognize(ctx context.Context, req port.OCRRequest) (*port.OCRResult, error) {
	result, err := svc.breaker.Execute(func() (any, error) {
		resp, err := svc.client.BatchAnnotateImages(ctx, &visionpb.BatchAnnotateImagesRequest{
			Requests: []*visionpb.AnnotateImageRequest{
				{
					Image:    &visionpb.Image{Content: req.Image},
					Features: []*visionpb.Feature{{Type: visionpb.Feature_DOCUMENT_TEXT_DETECTION}},
				},
			},
		})
		if err != nil {
			logger.WithContext(ctx).Error("[OCR] Google Vision API error", zap.Error(err))
			return nil, apperr.ServiceUnavailable("ocr.service_error", err)
		}

		responses := resp.GetResponses()
		if len(responses) == 0 || responses[0].GetFullTextAnnotation() == nil {
			return &port.OCRResult{Characters: []port.OCRCharacter{}, Engine: "google_vision"}, nil
		}

		annotation := responses[0].GetFullTextAnnotation()

		seen := make(map[string]struct{})
		var characters []port.OCRCharacter

		for _, page := range annotation.GetPages() {
			for _, block := range page.GetBlocks() {
				for _, paragraph := range block.GetParagraphs() {
					if req.Language == "zh" {
						characters = extractChinese(paragraph, seen, characters)
					} else {
						characters = extractWords(paragraph, seen, characters)
					}
				}
			}
		}

		return &port.OCRResult{Characters: characters, Engine: "google_vision"}, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*port.OCRResult), nil
}

// extractChinese extracts at symbol level — each symbol is 1 CJK character.
func extractChinese(paragraph *visionpb.Paragraph, seen map[string]struct{}, characters []port.OCRCharacter) []port.OCRCharacter {
	for _, word := range paragraph.GetWords() {
		for _, symbol := range word.GetSymbols() {
			text := symbol.GetText()
			if !isCJK(text) {
				continue
			}
			if _, exists := seen[text]; exists {
				continue
			}
			seen[text] = struct{}{}
			characters = append(characters, port.OCRCharacter{
				Text:       text,
				Confidence: float64(symbol.GetConfidence()),
			})
		}
	}
	return characters
}

// extractWords extracts at word level for non-Chinese languages.
func extractWords(paragraph *visionpb.Paragraph, seen map[string]struct{}, characters []port.OCRCharacter) []port.OCRCharacter {
	for _, word := range paragraph.GetWords() {
		var wordText string
		for _, symbol := range word.GetSymbols() {
			wordText += symbol.GetText()
		}
		if wordText == "" {
			continue
		}
		if _, exists := seen[wordText]; exists {
			continue
		}
		seen[wordText] = struct{}{}
		characters = append(characters, port.OCRCharacter{
			Text:       wordText,
			Confidence: float64(word.GetConfidence()),
		})
	}
	return characters
}

// isCJK checks if the string contains a CJK Unified Ideograph.
func isCJK(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}
