// Package dto define os payloads de entrada e sua validação — equivalente a
// create-event.dto.ts e create-speaker.dto.ts (decorators do
// class-validator: IsNotEmpty, IsString, IsNumber, IsDateString, IsEmail,
// IsBoolean).
package dto

import (
	"fmt"
	"regexp"
	"strings"
)

// CreateEventDto é o payload para criação de um evento.
type CreateEventDto struct {
	Nome       string `json:"nome"`
	Endereco   string `json:"endereco"`
	Capacidade int    `json:"capacidade"`
	Data       string `json:"data"`
}

// Validate replica as validações de create-event.dto.ts.
func (d CreateEventDto) Validate() error {
	if strings.TrimSpace(d.Nome) == "" {
		return fmt.Errorf("nome: não pode ser vazio")
	}
	if strings.TrimSpace(d.Endereco) == "" {
		return fmt.Errorf("endereco: não pode ser vazio")
	}
	if d.Capacidade <= 0 {
		return fmt.Errorf("capacidade: deve ser um número positivo")
	}
	if strings.TrimSpace(d.Data) == "" {
		return fmt.Errorf("data: não pode ser vazia")
	}
	return nil
}

// CreateSpeakerDto é o payload para criação de um palestrante.
type CreateSpeakerDto struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	TalkTitle string `json:"talkTitle"`
	IsGDE     *bool  `json:"isGDE"`
}

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// Validate replica as validações de create-speaker.dto.ts.
func (d CreateSpeakerDto) Validate() error {
	if strings.TrimSpace(d.Name) == "" {
		return fmt.Errorf("name: não pode ser vazio")
	}
	if strings.TrimSpace(d.Email) == "" || !emailPattern.MatchString(d.Email) {
		return fmt.Errorf("email: deve ser um endereço de e-mail válido")
	}
	if strings.TrimSpace(d.TalkTitle) == "" {
		return fmt.Errorf("talkTitle: não pode ser vazio")
	}
	if d.IsGDE == nil {
		return fmt.Errorf("isGDE: é obrigatório")
	}
	return nil
}
