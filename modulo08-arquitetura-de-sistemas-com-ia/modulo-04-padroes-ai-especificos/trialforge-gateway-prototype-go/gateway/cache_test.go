package gateway

import "testing"

func TestSemanticCacheConsultarVazio(t *testing.T) {
	cache := &SemanticCache{}
	resultado := cache.Consultar([]float64{1, 0, 0})
	if resultado.Entrada != nil {
		t.Errorf("cache vazio deveria devolver Entrada nil")
	}
	if resultado.Similaridade != 0 {
		t.Errorf("cache vazio deveria devolver similaridade 0, obtido %v", resultado.Similaridade)
	}
}

func TestSemanticCacheHit(t *testing.T) {
	cache := &SemanticCache{}
	cache.Adicionar("pergunta original", []float64{1, 0, 0}, "resposta original")

	resultado := cache.Consultar([]float64{1, 0, 0})
	if resultado.Entrada == nil {
		t.Fatal("esperado HIT no cache")
	}
	if resultado.Similaridade != 1 {
		t.Errorf("similaridade = %v, esperado 1", resultado.Similaridade)
	}
	if resultado.Entrada.Resposta != "resposta original" {
		t.Errorf("resposta = %q, esperado %q", resultado.Entrada.Resposta, "resposta original")
	}
}

func TestSemanticCacheEscolheMelhorSimilaridade(t *testing.T) {
	cache := &SemanticCache{}
	cache.Adicionar("pergunta distante", []float64{0, 1, 0}, "resposta distante")
	cache.Adicionar("pergunta próxima", []float64{1, 0, 0}, "resposta próxima")

	resultado := cache.Consultar([]float64{1, 0, 0})
	if resultado.Entrada.Resposta != "resposta próxima" {
		t.Errorf("esperado escolher a entrada mais similar, obtido %q", resultado.Entrada.Resposta)
	}
}
