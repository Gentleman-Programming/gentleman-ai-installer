package mcp

var defaultContext7ServerJSON = []byte("{\n  \"command\": \"npx\",\n  \"args\": [\n    \"-y\",\n    \"@upstash/context7-mcp\"\n  ]\n}\n")

var defaultContext7OverlayJSON = []byte("{\n  \"mcpServers\": {\n    \"context7\": {\n      \"command\": \"npx\",\n      \"args\": [\n        \"-y\",\n        \"@upstash/context7-mcp\"\n      ]\n    }\n  }\n}\n")

func DefaultContext7ServerJSON() []byte {
	content := make([]byte, len(defaultContext7ServerJSON))
	copy(content, defaultContext7ServerJSON)
	return content
}

func DefaultContext7OverlayJSON() []byte {
	content := make([]byte, len(defaultContext7OverlayJSON))
	copy(content, defaultContext7OverlayJSON)
	return content
}
