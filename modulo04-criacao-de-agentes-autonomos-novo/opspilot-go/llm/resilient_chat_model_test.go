package llm

import (
	"errors"
	"sync/atomic"
	"testing"

	"opspilot/domain"
	"opspilot/tools"
)

func alwaysFails(counter *atomic.Int32) ChatModel {
	return ChatModelFunc(func(messages []ChatMessage, availableTools []tools.Tool) (ModelResponse, error) {
		counter.Add(1)
		return ModelResponse{}, errors.New("provider indisponivel")
	})
}

func alwaysSucceeds(answer string) ChatModel {
	return ChatModelFunc(func(messages []ChatMessage, availableTools []tools.Tool) (ModelResponse, error) {
		return TextResponse(answer), nil
	})
}

func TestSucceedsOnPrimaryWithoutTouchingBackup(t *testing.T) {
	var backupCalls atomic.Int32
	primary := alwaysSucceeds("ok")
	backup := ChatModelFunc(func(messages []ChatMessage, availableTools []tools.Tool) (ModelResponse, error) {
		backupCalls.Add(1)
		return TextResponse("nao deveria chegar aqui"), nil
	})

	resilient := NewResilientChatModel(primary, backup)
	response, err := resilient.Invoke([]ChatMessage{UserMessage("oi")}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.Content != "ok" {
		t.Fatalf("expected 'ok', got %q", response.Content)
	}
	if backupCalls.Load() != 0 {
		t.Fatalf("expected backup untouched, got %d calls", backupCalls.Load())
	}
}

func TestFallsBackToBackupWhenPrimaryExhaustsRetries(t *testing.T) {
	var primaryCalls atomic.Int32
	primary := alwaysFails(&primaryCalls)
	backup := alwaysSucceeds("resposta do backup")

	resilient := NewResilientChatModel(primary, backup)
	response, err := resilient.Invoke([]ChatMessage{UserMessage("oi")}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.Content != "resposta do backup" {
		t.Fatalf("expected backup response, got %q", response.Content)
	}
	if int(primaryCalls.Load()) != RetryAttempts {
		t.Fatalf("expected %d primary calls, got %d", RetryAttempts, primaryCalls.Load())
	}
}

func TestThrowsModelUnavailableWhenNoBackupConfigured(t *testing.T) {
	var counter atomic.Int32
	resilient := NewResilientChatModel(alwaysFails(&counter), nil)

	_, err := resilient.Invoke([]ChatMessage{UserMessage("oi")}, nil)
	var unavailable *domain.ModelUnavailableError
	if !errors.As(err, &unavailable) {
		t.Fatalf("expected ModelUnavailableError, got %T (%v)", err, err)
	}
}

func TestThrowsModelUnavailableWhenBothPrimaryAndBackupFail(t *testing.T) {
	var primaryCounter, backupCounter atomic.Int32
	resilient := NewResilientChatModel(alwaysFails(&primaryCounter), alwaysFails(&backupCounter))

	_, err := resilient.Invoke([]ChatMessage{UserMessage("oi")}, nil)
	var unavailable *domain.ModelUnavailableError
	if !errors.As(err, &unavailable) {
		t.Fatalf("expected ModelUnavailableError, got %T (%v)", err, err)
	}
}
