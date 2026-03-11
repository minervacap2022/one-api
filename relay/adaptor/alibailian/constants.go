package alibailian

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Based on Alibaba Bailian pricing: https://help.aliyun.com/zh/model-studio/getting-started/models
var ModelRatios = map[string]adaptor.ModelConfig{
	// Qwen Models
	"qwen-turbo":              {Ratio: 0.3 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-plus":               {Ratio: 0.8 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-long":               {Ratio: 0.5 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-max":                {Ratio: 20.0 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-coder-plus":         {Ratio: 0.8 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-coder-plus-latest":  {Ratio: 0.8 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-coder-turbo":        {Ratio: 0.3 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-coder-turbo-latest": {Ratio: 0.3 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-mt-plus":            {Ratio: 0.8 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-mt-turbo":           {Ratio: 0.3 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwq-32b-preview":         {Ratio: 0.5 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// DeepSeek Models (hosted on Alibaba)
	"deepseek-r1": {Ratio: 1.0 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"deepseek-v3": {Ratio: 0.07 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// Qwen-Image Models (DashScope native endpoints)
	// https://help.aliyun.com/zh/model-studio/qwen-image-api
	"qwen-image-2.0-pro": {
		Ratio:           60 * ratio.MilliTokensRmb,
		CompletionRatio: 1,
		Image: &adaptor.ImagePricingConfig{
			DefaultSize:      "1024x1024",
			DefaultQuality:   "standard",
			PromptTokenLimit: 800,
			MinImages:        1,
			MaxImages:        6,
			SizeMultipliers: map[string]float64{
				"512x512":   0.25,
				"1024x1024": 1,
				"2048x2048": 4,
			},
		},
	},
	"qwen-image-2.0": {
		Ratio:           20 * ratio.MilliTokensRmb,
		CompletionRatio: 1,
		Image: &adaptor.ImagePricingConfig{
			DefaultSize:      "1024x1024",
			DefaultQuality:   "standard",
			PromptTokenLimit: 800,
			MinImages:        1,
			MaxImages:        6,
			SizeMultipliers: map[string]float64{
				"512x512":   0.25,
				"1024x1024": 1,
				"2048x2048": 4,
			},
		},
	},
	"qwen-image-max": {
		Ratio:           160 * ratio.MilliTokensRmb,
		CompletionRatio: 1,
		Image: &adaptor.ImagePricingConfig{
			DefaultSize:      "1024x1024",
			DefaultQuality:   "standard",
			PromptTokenLimit: 800,
			MinImages:        1,
			MaxImages:        1,
			SizeMultipliers: map[string]float64{
				"1328x1328": 1,
			},
		},
	},
	"qwen-image-plus": {
		Ratio:           40 * ratio.MilliTokensRmb,
		CompletionRatio: 1,
		Image: &adaptor.ImagePricingConfig{
			DefaultSize:      "1024x1024",
			DefaultQuality:   "standard",
			PromptTokenLimit: 500,
			MinImages:        1,
			MaxImages:        1,
			SizeMultipliers: map[string]float64{
				"1328x1328": 1,
			},
		},
	},
	"qwen-image": {
		Ratio:           40 * ratio.MilliTokensRmb,
		CompletionRatio: 1,
		Image: &adaptor.ImagePricingConfig{
			DefaultSize:      "1024x1024",
			DefaultQuality:   "standard",
			PromptTokenLimit: 500,
			MinImages:        1,
			MaxImages:        1,
			SizeMultipliers: map[string]float64{
				"1328x1328": 1,
			},
		},
	},
}

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)

// AlibailianToolingDefaults reflects that Bailian's public docs do not disclose per-tool pricing (retrieved 2025-11-12).
// Source: https://r.jina.ai/https://www.alibabacloud.com/help/en/model-studio/latest/billing (returns 404 for unauthenticated access)
var AlibailianToolingDefaults = adaptor.ChannelToolConfig{}
