// Package model define os tipos de domínio da API — equivalente às
// interfaces EventDTO e SpeakerDTO de shared-types/src/lib/*.dto.ts.
package model

// Event representa um evento do CFP (Call for Papers).
type Event struct {
	ID         string `json:"id"`
	Nome       string `json:"nome"`
	Endereco   string `json:"endereco"`
	Capacidade int    `json:"capacidade"`
	Data       string `json:"data"`
}

// Speaker representa um palestrante inscrito no CFP.
type Speaker struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	TalkTitle string `json:"talkTitle"`
	IsGDE     bool   `json:"isGDE"`
}
