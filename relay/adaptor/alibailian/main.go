package alibailian

import (
	"fmt"
	"strings"

	"github.com/Laisky/errors/v2"

	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

// qwenImageSyncModels lists Qwen-Image models that use the synchronous
// multimodal-generation endpoint instead of the async text2image endpoint.
var qwenImageSyncModels = map[string]bool{
	"qwen-image-2.0-pro": true,
	"qwen-image-2.0":     true,
	"qwen-image-max":     true,
}

// IsQwenImageSyncModel returns true for qwen-image-2.0 series models.
func IsQwenImageSyncModel(modelName string) bool {
	return qwenImageSyncModels[modelName]
}

// IsQwenImageModel returns true for any Qwen-Image model (sync or async).
func IsQwenImageModel(modelName string) bool {
	return strings.HasPrefix(modelName, "qwen-image")
}

func GetRequestURL(meta *meta.Meta) (string, error) {
	switch meta.Mode {
	case relaymode.ChatCompletions:
		return fmt.Sprintf("%s/compatible-mode/v1/chat/completions", meta.BaseURL), nil
	case relaymode.Embeddings:
		return fmt.Sprintf("%s/compatible-mode/v1/embeddings", meta.BaseURL), nil
	case relaymode.ImagesGenerations:
		if IsQwenImageSyncModel(meta.ActualModelName) {
			// qwen-image-2.0 series: synchronous multimodal-generation endpoint
			return fmt.Sprintf("%s/api/v1/services/aigc/multimodal-generation/generation", meta.BaseURL), nil
		}
		// Legacy models (qwen-image-plus, qwen-image, wanx-v1): async text2image endpoint
		return fmt.Sprintf("%s/api/v1/services/aigc/text2image/image-synthesis", meta.BaseURL), nil
	default:
	}
	return "", errors.Errorf("unsupported relay mode %d for ali bailian", meta.Mode)
}
