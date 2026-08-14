package trialforge

import (
	"strings"
	"testing"
)

// Porte dos 9 cenarios de rodarTestes()/rodar_testes() (JS/Python) — os
// mesmos casos, um por teste Go em vez de um contador manual de passou/total.

func TestEventoProtocoloProntoEhRico(t *testing.T) {
	resultado, err := RodarFluxoTrialForge("Estudo fase II", OpcoesFluxo{})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if resultado.DadoProtocolo.Versao != 1 {
		t.Errorf("versao = %d, esperado 1", resultado.DadoProtocolo.Versao)
	}
	if resultado.DadoProtocolo.Criterios["idadeMinima"] != 13 {
		t.Errorf("criterios[idadeMinima] = %d, esperado 13", resultado.DadoProtocolo.Criterios["idadeMinima"])
	}
}

func TestCaminhoFeliz(t *testing.T) {
	resultado, err := RodarFluxoTrialForge("Estudo fase II", OpcoesFluxo{})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if !resultado.Reacao.Ok {
		t.Error("esperado reacao.Ok = true")
	}
	if resultado.DecisaoSupervisor != "nenhum_retry_necessario" {
		t.Errorf("decisaoSupervisor = %q, esperado nenhum_retry_necessario", resultado.DecisaoSupervisor)
	}
	if resultado.Verificacao == nil || !resultado.Verificacao.Consistente {
		t.Error("esperado verificacao consistente")
	}
}

func TestBugEstrategiaAllPerdeResultadoDoCSR(t *testing.T) {
	resultado, err := RodarFluxoTrialForge("Estudo fase II", OpcoesFluxo{
		Estrategia:     EstrategiaAll,
		ForcarFalhaICF: true,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if resultado.Reacao.Ok {
		t.Error("esperado reacao.Ok = false")
	}
	if resultado.Reacao.ResultadoCSRPreservado {
		t.Error("esperado resultadoCSRPreservado = false (bug do paragrafo 68-71)")
	}
}

func TestCorrecaoEstrategiaAllSettledPreservaCSR(t *testing.T) {
	resultado, err := RodarFluxoTrialForge("Estudo fase II", OpcoesFluxo{
		Estrategia:     EstrategiaAllSettled,
		ForcarFalhaICF: true,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if !resultado.Reacao.ResultadoCSRPreservado {
		t.Error("esperado resultadoCSRPreservado = true")
	}
	if resultado.DecisaoSupervisor != "retry_icf_executado_com_sucesso" {
		t.Errorf("decisaoSupervisor = %q, esperado retry_icf_executado_com_sucesso", resultado.DecisaoSupervisor)
	}
	if !resultado.Reacao.Ok {
		t.Error("esperado reacao.Ok = true apos retry")
	}
}

func TestIdempotenciaDoRegistroDeDocumento(t *testing.T) {
	resultado, err := RodarFluxoTrialForge("Estudo fase II", OpcoesFluxo{
		Estrategia:     EstrategiaAllSettled,
		ForcarFalhaICF: true,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if resultado.Documentos.Tamanho() != 1 {
		t.Errorf("documentos.Tamanho() = %d, esperado 1 (nao pode duplicar documento)", resultado.Documentos.Tamanho())
	}
	if resultado.TentativasICF != 2 {
		t.Errorf("tentativasICF = %d, esperado 2", resultado.TentativasICF)
	}
}

func TestSequentialRespeitado(t *testing.T) {
	resultado, err := RodarFluxoTrialForge("Estudo de ordem", OpcoesFluxo{})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	ordem := resultado.OrdemDeExecucao
	idxProtocolo := indexOf(ordem, "protocolo")
	idxICF := indexOf(ordem, "icf:inicio")
	idxCSR := indexOf(ordem, "csr:inicio")

	if idxProtocolo == -1 {
		t.Fatalf("'protocolo' nao apareceu na ordem de execucao: %v", ordem)
	}
	if !(idxProtocolo < idxICF && idxProtocolo < idxCSR) {
		t.Errorf("Sequential violado — ordem real = %v", ordem)
	}
}

func TestDivergenciaDetectadaAposEmenda(t *testing.T) {
	resultado, err := RodarFluxoTrialForge("Estudo com emenda", OpcoesFluxo{ComEmendaEtica: true})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	v := resultado.Verificacao
	if v == nil {
		t.Fatal("esperada verificacao nao nula")
	}
	if v.Consistente {
		t.Error("esperado consistente = false")
	}
	if !v.ICFConsistente {
		t.Error("esperado ICF consistente")
	}
	if v.CSRConsistente {
		t.Error("esperado CSR defasado")
	}
	if len(v.AgentesDivergentes) != 1 || v.AgentesDivergentes[0] != "CSR" {
		t.Errorf("agentesDivergentes = %v, esperado [CSR]", v.AgentesDivergentes)
	}
}

func TestCompensacaoSagaRegeneraApenasCSR(t *testing.T) {
	resultado, err := RodarFluxoTrialForge("Estudo com emenda", OpcoesFluxo{ComEmendaEtica: true})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	csr := resultado.Reacao.ResultadoCSR
	if csr == nil {
		t.Fatal("esperado resultadoCSR nao nulo")
	}
	if csr.CriterioUsado["idadeMinima"] != 12 {
		t.Errorf("CSR.CriterioUsado[idadeMinima] = %d, esperado 12", csr.CriterioUsado["idadeMinima"])
	}
	if csr.VersaoUsada != 2 {
		t.Errorf("CSR.VersaoUsada = %d, esperado 2", csr.VersaoUsada)
	}
	if resultado.DecisaoSupervisor != "saga_compensacao_csr" {
		t.Errorf("decisaoSupervisor = %q, esperado saga_compensacao_csr", resultado.DecisaoSupervisor)
	}
	if len(resultado.EstadoProtocolo.Historico()) != 1 {
		t.Errorf("historico tem %d entradas, esperado 1", len(resultado.EstadoProtocolo.Historico()))
	}
}

func TestTimeoutRealDetectadoERecuperadoNoRetry(t *testing.T) {
	resultado, err := RodarFluxoTrialForge("Estudo com agente travado", OpcoesFluxo{ForcarTravamentoICF: true})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if !strings.Contains(resultado.Reacao.ErroICF, "timeout") {
		t.Errorf("erroICF = %q, esperado conter 'timeout'", resultado.Reacao.ErroICF)
	}
	if resultado.DecisaoSupervisor != "retry_icf_executado_com_sucesso" {
		t.Errorf("decisaoSupervisor = %q, esperado retry_icf_executado_com_sucesso", resultado.DecisaoSupervisor)
	}
	if resultado.TentativasRetry != 1 {
		t.Errorf("tentativasRetry = %d, esperado 1", resultado.TentativasRetry)
	}
}

func TestRetryEsgotaLimiteEmFalhaPersistente(t *testing.T) {
	resultado, err := RodarFluxoTrialForge("Estudo com falha persistente", OpcoesFluxo{
		ForcarTravamentoICF:   true,
		PersistirFalhaNoRetry: true,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if resultado.DecisaoSupervisor != "retry_esgotado_seguindo_sem_icf" {
		t.Errorf("decisaoSupervisor = %q, esperado retry_esgotado_seguindo_sem_icf", resultado.DecisaoSupervisor)
	}
	if resultado.TentativasRetry != MaxTentativas {
		t.Errorf("tentativasRetry = %d, esperado %d", resultado.TentativasRetry, MaxTentativas)
	}
	if resultado.Reacao.Ok {
		t.Error("esperado reacao.Ok = false")
	}
}

func TestEscopoIsoladoPorFluxo(t *testing.T) {
	primeiro, err := RodarFluxoTrialForge("Estudo A", OpcoesFluxo{})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	segundo, err := RodarFluxoTrialForge("Estudo B", OpcoesFluxo{})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if primeiro.EstadoProtocolo == segundo.EstadoProtocolo {
		t.Error("EstadoProtocolo nao deveria ser compartilhado entre fluxos")
	}
	if primeiro.DadoProtocolo.Versao != 1 || segundo.DadoProtocolo.Versao != 1 {
		t.Error("cada fluxo deveria comecar do zero, na versao 1")
	}
}

func indexOf(itens []string, alvo string) int {
	for i, item := range itens {
		if item == alvo {
			return i
		}
	}
	return -1
}
