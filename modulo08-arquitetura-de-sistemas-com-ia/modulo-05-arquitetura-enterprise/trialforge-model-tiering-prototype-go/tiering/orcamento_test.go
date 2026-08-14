package tiering

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestVerificarOrcamento_ComFolga_RetornaTrue(t *testing.T) {
	o := NewOrcamentoManager()
	o.DefinirOrcamento("estudo-teste", 0.05)
	o.RegistrarGasto("estudo-teste", 0.045)

	ok, err := o.VerificarOrcamento("estudo-teste", 0.004) // 0.045+0.004=0.049 <= 0.05
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("esperava true")
	}
}

func TestVerificarOrcamento_EstourandoLimite_RetornaFalse(t *testing.T) {
	o := NewOrcamentoManager()
	o.DefinirOrcamento("estudo-teste", 0.05)
	o.RegistrarGasto("estudo-teste", 0.045)

	ok, err := o.VerificarOrcamento("estudo-teste", 0.006) // 0.045+0.006=0.051 > 0.05
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("esperava false")
	}
}

func TestVerificarOrcamento_NaoMutaEstado(t *testing.T) {
	o := NewOrcamentoManager()
	o.DefinirOrcamento("estudo-teste", 0.05)
	o.RegistrarGasto("estudo-teste", 0.045)

	o.VerificarOrcamento("estudo-teste", 0.004)

	if !closeEnough(o.GastoAtual("estudo-teste"), 0.045) {
		t.Errorf("esperava gasto inalterado 0.045, obteve %v", o.GastoAtual("estudo-teste"))
	}
}

func TestReservarOrcamento_ComFolga_DebitaEDevolveTrue(t *testing.T) {
	o := NewOrcamentoManager()
	o.DefinirOrcamento("estudo-A", 0.05)

	reservou, err := o.ReservarOrcamento("estudo-A", 0.011)
	if err != nil {
		t.Fatal(err)
	}
	if !reservou {
		t.Error("esperava reservar com sucesso")
	}
	if !closeEnough(o.GastoAtual("estudo-A"), 0.011) {
		t.Errorf("esperava gasto 0.011, obteve %v", o.GastoAtual("estudo-A"))
	}
}

func TestReservarOrcamento_EstourandoLimite_NaoDebitaEDevolveFalse(t *testing.T) {
	o := NewOrcamentoManager()
	o.DefinirOrcamento("estudo-B", 0.005)

	reservou, err := o.ReservarOrcamento("estudo-B", 0.011)
	if err != nil {
		t.Fatal(err)
	}
	if reservou {
		t.Error("esperava falhar a reserva")
	}
	if o.GastoAtual("estudo-B") != 0 {
		t.Errorf("esperava gasto 0, obteve %v", o.GastoAtual("estudo-B"))
	}
}

func TestLiberarSobra_DevolveDiferenca(t *testing.T) {
	o := NewOrcamentoManager()
	o.DefinirOrcamento("estudo-A", 0.05)
	o.ReservarOrcamento("estudo-A", 0.011) // reserva pior caso (Tier1+Tier2)

	o.LiberarSobra("estudo-A", 0.011-0.001) // Tier 1 resolveu sozinho, custou só 0.001

	if !closeEnough(o.GastoAtual("estudo-A"), 0.001) {
		t.Errorf("esperava gasto 0.001, obteve %v", o.GastoAtual("estudo-A"))
	}
}

func TestEstudoDesconhecido_RetornaErro(t *testing.T) {
	o := NewOrcamentoManager()
	_, err := o.VerificarOrcamento("estudo-fantasma", 0.001)
	if err == nil {
		t.Error("esperava erro para estudo desconhecido")
	}
}

// Réplica do cenário "estudo-F" de SimularVolumeConcorrente: orçamento cabe
// exatamente 2 reservas do pior caso, mas N goroutines reais (sobre múltiplas
// threads do SO) disputam a mesma conta ao mesmo tempo. ReservarOrcamento
// deve continuar sendo a única porta de entrada — nenhuma combinação de
// goroutines pode fazer o gasto ultrapassar o limite.
func TestReservarOrcamento_SobConcorrenciaReal_NuncaEstouraOLimite(t *testing.T) {
	custoPiorCaso := 0.011
	numeroDeGoroutines := 20
	limite := 2*custoPiorCaso + 0.0005 // cabe exatamente 2 reservas

	o := NewOrcamentoManager()
	o.DefinirOrcamento("estudo-F", limite)

	var largada sync.WaitGroup
	largada.Add(1)
	var chegada sync.WaitGroup
	chegada.Add(numeroDeGoroutines)
	var reservasBemSucedidas int64

	for i := 0; i < numeroDeGoroutines; i++ {
		go func() {
			defer chegada.Done()
			largada.Wait()
			ok, _ := o.ReservarOrcamento("estudo-F", custoPiorCaso)
			if ok {
				atomic.AddInt64(&reservasBemSucedidas, 1)
			}
		}()
	}

	largada.Done() // solta todas as goroutines ao mesmo tempo
	chegada.Wait()

	if o.GastoAtual("estudo-F") > limite+1e-9 {
		t.Errorf("orçamento estourou: gasto %v, limite %v", o.GastoAtual("estudo-F"), limite)
	}
	if reservasBemSucedidas != 2 {
		t.Errorf("esperava exatamente 2 reservas bem-sucedidas de %d goroutines, obteve %d", numeroDeGoroutines, reservasBemSucedidas)
	}
}
