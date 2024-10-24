package chains

import (
	"vinesai/internel/langchaingo/llms"
	"vinesai/internel/langchaingo/outputparser"
	"vinesai/internel/langchaingo/prompts"
	"vinesai/internel/langchaingo/schema"
)

//nolint:lll
const _conversationTemplate = `The following is a friendly conversation between a human and an AI. The AI is talkative and provides lots of specific details from its context. If the AI does not know the answer to a question, it truthfully says it does not know.

Current conversation:
{{.history}}
Human: {{.input}}
AI:`

func NewConversation(llm llms.Model, memory schema.Memory, prompt ...string) LLMChain {
	var p = _conversationTemplate
	if len(prompt) > 0 {
		p = prompt[0]
	}
	return LLMChain{
		Prompt: prompts.NewPromptTemplate(
			p,
			[]string{"history", "input"},
		),
		LLM:          llm,
		Memory:       memory,
		OutputParser: outputparser.NewSimple(),
		OutputKey:    _llmChainDefaultOutputKey,
	}
}
