package tiering

import (
	"fmt"
	"sync"
)

// conta é o estado de orçamento de um estudo, protegido por seu próprio mutex.
type conta struct {
	mu     sync.Mutex
	limite float64
	gasto  float64
}

// OrcamentoManager gerencia orçamento por estudo (Módulo 5.1 Uber/Vitalis
// Platform, Módulo 5.2 LiteLLM).
//
// ReservarOrcamento é uma reserva otimista — checa E debita atomicamente
// (mutex por conta de estudo), fechando a mesma janela de corrida discutida
// no comentário do original em JS: lá, o event loop de Node garante que não
// existe `await` entre checar e debitar dentro de uma única chamada; aqui,
// como Go roda goroutines de verdade sobre múltiplas threads do SO (ver
// SimularVolumeConcorrente), a mesma garantia exige um lock explícito — sem
// ele, N goroutines concorrentes do mesmo estudo poderiam passar todas pela
// checagem antes de qualquer uma debitar.
type OrcamentoManager struct {
	mu     sync.RWMutex
	contas map[string]*conta
}

func NewOrcamentoManager() *OrcamentoManager {
	return &OrcamentoManager{contas: map[string]*conta{}}
}

func (o *OrcamentoManager) DefinirOrcamento(estudoID string, limite float64) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.contas[estudoID] = &conta{limite: limite}
}

func (o *OrcamentoManager) obter(estudoID string) (*conta, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	c, ok := o.contas[estudoID]
	if !ok {
		return nil, fmt.Errorf("estudo desconhecido: %s", estudoID)
	}
	return c, nil
}

func (o *OrcamentoManager) VerificarOrcamento(estudoID string, custoEstimado float64) (bool, error) {
	c, err := o.obter(estudoID)
	if err != nil {
		return false, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.gasto+custoEstimado <= c.limite, nil
}

func (o *OrcamentoManager) RegistrarGasto(estudoID string, custo float64) error {
	c, err := o.obter(estudoID)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.gasto += custo
	return nil
}

// ReservarOrcamento reserva o pior caso ANTES de qualquer chamada de modelo —
// checa e debita no mesmo passo, sob o mesmo lock.
func (o *OrcamentoManager) ReservarOrcamento(estudoID string, custoReservado float64) (bool, error) {
	c, err := o.obter(estudoID)
	if err != nil {
		return false, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.gasto+custoReservado > c.limite {
		return false, nil
	}
	c.gasto += custoReservado
	return true, nil
}

// LiberarSobra devolve a diferença entre o que foi reservado (pior caso) e o
// custo real, se sobrar algo.
func (o *OrcamentoManager) LiberarSobra(estudoID string, valorASobrar float64) error {
	if valorASobrar <= 0 {
		return nil
	}
	c, err := o.obter(estudoID)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.gasto -= valorASobrar
	return nil
}

func (o *OrcamentoManager) GastoAtual(estudoID string) float64 {
	c, err := o.obter(estudoID)
	if err != nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.gasto
}

func (o *OrcamentoManager) Limite(estudoID string) float64 {
	c, err := o.obter(estudoID)
	if err != nil {
		return 0
	}
	return c.limite
}
