package trialforge

import "sync"

// DadoProtocolo e o evento rico publicado em "protocolo:pronto" — carrega a
// versao e os criterios estruturados, nao so uma notificacao vazia
// (paragrafos 62-64 do TP). E uma COPIA do estado no momento da publicacao:
// revisoes futuras no EstadoProtocolo nao mudam retroativamente o que ja foi
// publicado neste evento.
type DadoProtocolo struct {
	Versao    int
	Criterios map[string]int
	Estudo    string
}

// RevisaoHistorico e um item do historico de revisoes do Protocolo — nunca
// sobrescrito, so acrescentado.
type RevisaoHistorico struct {
	VersaoAnterior      int
	CriteriosAnteriores map[string]int
	Motivo              string
}

// EstadoProtocolo e o porte de criarEstadoProtocolo/revisarProtocolo
// (JS/Python).
//
// O Protocolo e um recurso VERSIONADO e mutavel, nao um valor congelado —
// uma emenda pode chegar enquanto ICF/CSR ja estao trabalhando com uma copia
// mais antiga (Modulo 3.4: "O Protocolo versao 1 nao e apagado — ele e
// versionado, preservado como historico, e uma versao 2 e criada com o
// criterio corrigido.").
//
// Thread-safe (protegido por mutex): e lido pelas goroutines do ICF e do CSR
// e escrito pela goroutine que processa a emenda etica, que roda de forma
// independente do fluxo principal.
type EstadoProtocolo struct {
	mu        sync.Mutex
	versao    int
	criterios map[string]int
	historico []RevisaoHistorico
}

// NovoEstadoProtocolo cria o estado inicial do Protocolo, na versao 1.
func NovoEstadoProtocolo(criteriosIniciais map[string]int) *EstadoProtocolo {
	copia := make(map[string]int, len(criteriosIniciais))
	for k, v := range criteriosIniciais {
		copia[k] = v
	}
	return &EstadoProtocolo{versao: 1, criterios: copia}
}

// Versao devolve a versao vigente no momento da chamada.
func (e *EstadoProtocolo) Versao() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.versao
}

// Criterios devolve uma COPIA dos criterios vigentes.
func (e *EstadoProtocolo) Criterios() map[string]int {
	e.mu.Lock()
	defer e.mu.Unlock()
	copia := make(map[string]int, len(e.criterios))
	for k, v := range e.criterios {
		copia[k] = v
	}
	return copia
}

// Historico devolve uma COPIA do historico de revisoes.
func (e *EstadoProtocolo) Historico() []RevisaoHistorico {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]RevisaoHistorico(nil), e.historico...)
}

// Revisar acrescenta uma entrada ao historico (preservando a versao
// anterior) e mescla os novos criterios na versao seguinte.
func (e *EstadoProtocolo) Revisar(novosCriterios map[string]int, motivo string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	criteriosAnteriores := make(map[string]int, len(e.criterios))
	for k, v := range e.criterios {
		criteriosAnteriores[k] = v
	}
	e.historico = append(e.historico, RevisaoHistorico{
		VersaoAnterior:      e.versao,
		CriteriosAnteriores: criteriosAnteriores,
		Motivo:              motivo,
	})
	e.versao++
	for k, v := range novosCriterios {
		e.criterios[k] = v
	}
}
