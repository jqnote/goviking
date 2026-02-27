// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package llm

// SiliconFlowProviderType is the provider type for SiliconFlow.
const SiliconFlowProviderType ProviderType = "siliconflow"

// DefaultSiliconFlowBaseURL is the default base URL for SiliconFlow API.
const DefaultSiliconFlowBaseURL = "https://api.siliconflow.cn/v1"

// SiliconFlowProvider implements Provider for SiliconFlow (硅基流动).
type SiliconFlowProvider struct {
	*OpenAIProvider
}

// NewSiliconFlowProvider creates a new SiliconFlow provider.
func NewSiliconFlowProvider(apiKey, model string) *SiliconFlowProvider {
	return &SiliconFlowProvider{
		OpenAIProvider: NewOpenAIProvider(apiKey, DefaultSiliconFlowBaseURL, model),
	}
}

// AvailableSiliconFlowModels returns the available models on SiliconFlow.
func AvailableSiliconFlowModels() []string {
	return []string{
		// DeepSeek Models
		"deepseek-ai/DeepSeek-V2-Chat",
		"deepseek-ai/DeepSeek-V2",
		"deepseek-ai/DeepSeek-Coder-V2-Instruct",
		"deepseek-ai/DeepSeek-Coder-V2",
		// Qwen Models
		"Qwen/Qwen2-72B-Instruct",
		"Qwen/Qwen2-57B-A14B-Instruct",
		"Qwen/Qwen2-7B-Instruct",
		"Qwen/Qwen2-1.8B-Instruct",
		"Qwen/Qwen1.5-72B-Chat",
		"Qwen/Qwen1.5-110B-Chat",
		"Qwen/Qwen1.5-7B-Chat",
		"Qwen/Qwen1.5-14B-Chat",
		// Llama Models
		"meta-llama/Meta-Llama-3-70B-Instruct",
		"meta-llama/Meta-Llama-3-8B-Instruct",
		"meta-llama/Meta-Llama-3.1-70B-Instruct",
		"meta-llama/Meta-Llama-3.1-8B-Instruct",
		// GLM Models
		"THUDM/glm-4-9b-chat",
		"THUDM/glm-4-9b",
		"THUDM/glm-4-vision-accept",
		"THUDM/glm-4v-flash",
		// Yi Models
		"01-ai/Yi-1.5-34B-Chat",
		"01-ai/Yi-1.5-9B-Chat",
		"01-ai/Yi-1.5-6B-Chat",
		// Baichuan Models
		"baichuan-inc/Baichuan2-53B",
		"baichuan-inc/Baichuan2-7B-Chat",
		"baichuan-inc/Baichuan2-13B-Chat",
		// Other Models
		"Pro/Bert-VITS2",
		"microsoft/WizardLM-2-8x22B",
		"microsoft/Phi-3-mini-4k-instruct",
		"mistralai/Mistral-7B-Instruct-v0.2",
		"mistralai/Mixtral-8x7B-Instruct-v0.1",
		"google/gemma-2-9b-it",
		"google/gemma-2-27b-it",
		"google/gemma-7b-it",
		"anthropic/claude-3.5-sonnet",
		"anthropic/claude-3-opus",
		"anthropic/claude-3-haiku",
		"openai/gpt-4o",
		"openai/gpt-4o-mini",
		"openai/gpt-4-turbo",
		"openai/gpt-3.5-turbo",
		// Embedding Models
		"BAAI/bge-large-zh-v1.5",
		"BAAI/bge-large-en-v1.5",
		"BAAI/bge-base-zh-v1.5",
		"BAAI/bge-base-en-v1.5",
		"BAAI/bge-small-zh-v1.5",
		"BAAI/bge-small-en-v1.5",
		"Pro/Bert-VITS2",
	}
}
