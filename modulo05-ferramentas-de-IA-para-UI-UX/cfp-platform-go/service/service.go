// Package service implementa a lógica de negócio em memória — equivalente a
// event.service.ts e speaker.service.ts.
package service

import (
	"strings"
	"sync"

	"github.com/google/uuid"

	"cfp-platform/dto"
	"cfp-platform/model"
)

// EventService gerencia eventos em memória, sem persistência — mesmo
// estágio de "minimal backend" da versão NestJS original.
type EventService struct {
	mu     sync.Mutex
	events []model.Event
}

// Create cria um novo evento com um id curto e aleatório — equivalente a
// Math.random().toString(36).substring(2, 9) do TypeScript original.
func (s *EventService) Create(d dto.CreateEventDto) model.Event {
	event := model.Event{ID: shortID(), Nome: d.Nome, Endereco: d.Endereco, Capacidade: d.Capacidade, Data: d.Data}

	s.mu.Lock()
	s.events = append(s.events, event)
	s.mu.Unlock()

	return event
}

// FindAll retorna todos os eventos criados.
func (s *EventService) FindAll() []model.Event {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]model.Event, len(s.events))
	copy(result, s.events)
	return result
}

// SpeakerService gerencia palestrantes em memória, sem persistência.
type SpeakerService struct {
	mu       sync.Mutex
	speakers []model.Speaker
}

// Create cria um novo palestrante com um id curto e aleatório.
func (s *SpeakerService) Create(d dto.CreateSpeakerDto) model.Speaker {
	speaker := model.Speaker{ID: shortID(), Name: d.Name, Email: d.Email, TalkTitle: d.TalkTitle, IsGDE: *d.IsGDE}

	s.mu.Lock()
	s.speakers = append(s.speakers, speaker)
	s.mu.Unlock()

	return speaker
}

// FindAll retorna todos os palestrantes criados.
func (s *SpeakerService) FindAll() []model.Speaker {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]model.Speaker, len(s.speakers))
	copy(result, s.speakers)
	return result
}

// shortID gera um id curto e aleatório — mesmo propósito de
// Math.random().toString(36).substring(2, 9), usando UUID truncado como
// fonte de aleatoriedade portável.
func shortID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")[:7]
}
