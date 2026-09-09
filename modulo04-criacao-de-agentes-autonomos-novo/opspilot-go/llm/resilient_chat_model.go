package llm

import (
	"opspilot/domain"
	"opspilot/tools"
)

// RetryAttempts é o porte de ResilientChatModel.RETRY_ATTEMPTS (porte Java).
const RetryAttempts = 2

// ResilientChatModel é o porte de OpsResilientChatModel /
// composeResilientRunnable (agents/model.ts): retry no primário -> fallback
// pro backup (se configurado) -> retorna ModelUnavailableError se ambos
// falharem após as tentativas. Backup nil equivale a "sem backup configurado"
// (null no original).
type ResilientChatModel struct {
	primary ChatModel
	backup  ChatModel
}

var _ ChatModel = (*ResilientChatModel)(nil)

func NewResilientChatModel(primary, backup ChatModel) *ResilientChatModel {
	return &ResilientChatModel{primary: primary, backup: backup}
}

func (r *ResilientChatModel) Invoke(messages []ChatMessage, availableTools []tools.Tool) (ModelResponse, error) {
	var lastErr error

	for attempt := 0; attempt < RetryAttempts; attempt++ {
		response, err := r.primary.Invoke(messages, availableTools)
		if err == nil {
			return response, nil
		}
		lastErr = err
	}

	if r.backup == nil {
		return ModelResponse{}, &domain.ModelUnavailableError{Message: messageOf(lastErr)}
	}

	for attempt := 0; attempt < RetryAttempts; attempt++ {
		response, err := r.backup.Invoke(messages, availableTools)
		if err == nil {
			return response, nil
		}
		lastErr = err
	}
	return ModelResponse{}, &domain.ModelUnavailableError{Message: messageOf(lastErr)}
}

func messageOf(err error) string {
	if err != nil {
		return err.Error()
	}
	return "All configured language models failed"
}
