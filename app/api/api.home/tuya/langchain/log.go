package langchain

import (
	"context"
	"fmt"
	"strings"
	"time"
	"vinesai/internel/ava"

	"vinesai/internel/langchaingo/callbacks"
	"vinesai/internel/langchaingo/llms"
	"vinesai/internel/langchaingo/schema"
)

// LogHandler is a callback handler that prints to the standard output.
type LogHandler struct{}

var _ callbacks.Handler = LogHandler{}

func (l LogHandler) HandleLLMGenerateContentStart(_ context.Context, ms []llms.MessageContent) {
	ava.Debug("Entering LLM with messages:", time.Now())
	for _, m := range ms {
		// TODO: Implement logging of other content types
		var buf strings.Builder
		for _, t := range m.Parts {
			if t, ok := t.(llms.TextContent); ok {
				buf.WriteString(t.Text)
			}
		}
		ava.Debug("Role:", m.Role)
		ava.Debug("Text:", buf.String())
	}
}

func (l LogHandler) HandleLLMGenerateContentEnd(_ context.Context, res *llms.ContentResponse) {
	ava.Debug("Exiting LLM with response:")
	for _, c := range res.Choices {
		if c.Content != "" {
			ava.Debug("Content:", c.Content)
		}
		if c.StopReason != "" {
			ava.Debug("StopReason:", c.StopReason)
		}
		if len(c.GenerationInfo) > 0 {
			ava.Debug("GenerationInfo:")
			for k, v := range c.GenerationInfo {
				ava.Debugf("%20s: %v\n", k, v)
			}
		}
		if c.FuncCall != nil {
			ava.Debug("------------------FuncCall: ", c.FuncCall.Name, c.FuncCall.Arguments)
			ava.Debug("---", time.Now())
		}
	}
}

func (l LogHandler) HandleStreamingFunc(_ context.Context, chunk []byte) {
	ava.Debug(string(chunk))
}

func (l LogHandler) HandleText(_ context.Context, text string) {
	ava.Debug(text)
}

func (l LogHandler) HandleLLMStart(_ context.Context, prompts []string) {
	ava.Debug("Entering LLM with prompts:", prompts)
}

func (l LogHandler) HandleLLMError(_ context.Context, err error) {
	ava.Debug("Exiting LLM with error:", err)
}

func (l LogHandler) HandleChainStart(_ context.Context, inputs map[string]any) {
	ava.Debug("Entering chain with inputs:", formatChainValues(inputs))
}

func (l LogHandler) HandleChainEnd(_ context.Context, outputs map[string]any) {
	ava.Debug("Exiting chain with outputs:", formatChainValues(outputs))
}

func (l LogHandler) HandleChainError(_ context.Context, err error) {
	ava.Debug("Exiting chain with error:", err)
}

func (l LogHandler) HandleToolStart(_ context.Context, input string) {
	ava.Debug("Entering tool with input:", removeNewLines(input))
}

func (l LogHandler) HandleToolEnd(_ context.Context, output string) {
	ava.Debug("Exiting tool with output:", removeNewLines(output))
}

func (l LogHandler) HandleToolError(_ context.Context, err error) {
	ava.Debug("Exiting tool with error:", err)
}

func (l LogHandler) HandleAgentAction(_ context.Context, action schema.AgentAction) {
	ava.Debug("Agent selected action:", formatAgentAction(action))
}

func (l LogHandler) HandleAgentFinish(_ context.Context, finish schema.AgentFinish) {
	ava.Debugf("Agent finish: %v \n", finish)
}

func (l LogHandler) HandleRetrieverStart(_ context.Context, query string) {
	ava.Debug("Entering retriever with queryOnline:", removeNewLines(query))
}

func (l LogHandler) HandleRetrieverEnd(_ context.Context, query string, documents []schema.Document) {
	ava.Debug("Exiting retriever with documents for queryOnline:", documents, query)
}

func formatChainValues(values map[string]any) string {
	output := ""
	for key, value := range values {
		output += fmt.Sprintf("\"%s\" : \"%s\", ", removeNewLines(key), removeNewLines(value))
	}

	return output
}

func formatAgentAction(action schema.AgentAction) string {
	return fmt.Sprintf("\"%s\" with input \"%s\"", removeNewLines(action.Tool), removeNewLines(action.ToolInput))
}

func removeNewLines(s any) string {
	return strings.ReplaceAll(fmt.Sprint(s), "\n", " ")
}
