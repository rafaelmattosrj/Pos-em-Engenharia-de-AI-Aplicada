package dto

import "testing"

func TestCreateEventDto_Validate(t *testing.T) {
	valid := CreateEventDto{Nome: "DevFest", Endereco: "Centro", Capacidade: 500, Data: "2026-08-15"}
	if err := valid.Validate(); err != nil {
		t.Errorf("esperava valido, obteve erro: %v", err)
	}

	invalid := CreateEventDto{Nome: "", Endereco: "Centro", Capacidade: 100, Data: "2026-08-15"}
	if err := invalid.Validate(); err == nil {
		t.Error("esperava erro para nome vazio")
	}

	semCapacidade := CreateEventDto{Nome: "X", Endereco: "Y", Capacidade: 0, Data: "2026-08-15"}
	if err := semCapacidade.Validate(); err == nil {
		t.Error("esperava erro para capacidade <= 0")
	}
}

func TestCreateSpeakerDto_Validate(t *testing.T) {
	isGDE := true
	valid := CreateSpeakerDto{Name: "Ana", Email: "ana@example.com", TalkTitle: "IA na pratica", IsGDE: &isGDE}
	if err := valid.Validate(); err != nil {
		t.Errorf("esperava valido, obteve erro: %v", err)
	}

	invalidEmail := CreateSpeakerDto{Name: "Bruno", Email: "nao-e-email", TalkTitle: "Talk", IsGDE: &isGDE}
	if err := invalidEmail.Validate(); err == nil {
		t.Error("esperava erro para email invalido")
	}

	semIsGDE := CreateSpeakerDto{Name: "Carla", Email: "carla@example.com", TalkTitle: "Talk", IsGDE: nil}
	if err := semIsGDE.Validate(); err == nil {
		t.Error("esperava erro para isGDE ausente")
	}
}
