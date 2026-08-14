package service

import (
	"testing"

	"cfp-platform/dto"
)

func TestEventService_CreateAndFindAll(t *testing.T) {
	s := &EventService{}

	created := s.Create(dto.CreateEventDto{Nome: "DevFest", Endereco: "Centro", Capacidade: 500, Data: "2026-08-15"})
	if created.ID == "" {
		t.Error("esperava id gerado")
	}
	if created.Nome != "DevFest" {
		t.Errorf("nome inesperado: %q", created.Nome)
	}

	all := s.FindAll()
	if len(all) != 1 {
		t.Fatalf("esperava 1 evento, obteve %d", len(all))
	}
}

func TestSpeakerService_CreateAndFindAll(t *testing.T) {
	s := &SpeakerService{}
	isGDE := true

	created := s.Create(dto.CreateSpeakerDto{Name: "Ana", Email: "ana@example.com", TalkTitle: "IA", IsGDE: &isGDE})
	if created.ID == "" {
		t.Error("esperava id gerado")
	}
	if !created.IsGDE {
		t.Error("esperava isGDE=true")
	}

	all := s.FindAll()
	if len(all) != 1 {
		t.Fatalf("esperava 1 palestrante, obteve %d", len(all))
	}
}
