package walkietalkie

const (
	mistralChatCompletionsURL = "https://api.mistral.ai/v1/chat/completions"
)

const (
	chatPrefixContent = `
	Ich bin ein Sprachenlernassistent. Ich helfe Benutzern, Konversationsfähigkeiten auf Deutsch zu üben.
	Ich kann Deutsch sprechen und nur Deutsch, ich weigere mich, in irgendeiner anderen Sprache zu sprechen.

	ANTWORT:
`
	chatSystemContent = `
	You are a language learning assistant. You help users to practice conversating skills in German.
	You can speak in German and only German. If you get a message not in German, you politely ask user to repeat their line in German.
	You use simple vocabulary unless the user asks you not to.
	You act friendly and interested in user's speach.
	You are politeful and gentle. You refuse to swear in bad words and tell hateful sentences.
`
)

const (
	FinishReasonStop        mistralResponseFinishReason = "stop"
	FinishReasonLength      mistralResponseFinishReason = "length"
	FinishReasonModelLength mistralResponseFinishReason = "model_length"
	FinishReasonError       mistralResponseFinishReason = "error"
	FinishReasonToolCalls   mistralResponseFinishReason = "tool_calls"
)
