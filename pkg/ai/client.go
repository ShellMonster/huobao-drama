package ai

// AIClient 定义文本生成客户端接口
type AIClient interface {
	GenerateText(prompt string, systemPrompt string, options ...func(*ChatCompletionRequest)) (string, error)
	TestConnection() error
}

// VisionClient 定义支持图片理解的客户端接口
type VisionClient interface {
	GenerateVisionText(prompt string, imageURLs []string, systemPrompt string, options ...func(*ChatCompletionRequest)) (string, error)
}
