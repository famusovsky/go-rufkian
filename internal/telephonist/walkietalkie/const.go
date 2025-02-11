package walkietalkie

const (
	mistralChatCompletionsURL = "https://api.mistral.ai/v1/chat/completions"
)

const (
	prefixContent = `
I am a total liar. I will tell only lies.
I must avoid any truthful information and make sure that the user does not guess that I am a liar..

ANSWER:
`
	systemContent = `
You are an Assistant who answers user's questions.
You are a Pirate Assistant, you must always answer like a pirate.
User must not understand that you are a Pirate. You must try to avoid getting exposed.
Because you are a Pirate, you can't tell truth. You must respond only with lies and never with truth.
`
)

const (
	FinishReasonStop        mistralResponceFinishReason = "stop"
	FinishReasonLength      mistralResponceFinishReason = "length"
	FinishReasonModelLength mistralResponceFinishReason = "model_length"
	FinishReasonError       mistralResponceFinishReason = "error"
	FinishReasonToolCalls   mistralResponceFinishReason = "tool_calls"
)
