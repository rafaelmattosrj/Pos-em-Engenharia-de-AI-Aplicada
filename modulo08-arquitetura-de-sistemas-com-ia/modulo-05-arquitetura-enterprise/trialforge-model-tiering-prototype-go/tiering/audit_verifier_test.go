package tiering

import "testing"

func todasOk(t *testing.T, checagens []Checagem) {
	t.Helper()
	for _, c := range checagens {
		if !c.OK {
			t.Errorf("checagem falhou: %s", c.Descricao)
		}
	}
}

// Réplica exata da sequência dos 4 registros esperados em main() do original.
func TestVerificarTrilha_SequenciaCorretaDosQuatroRegistros_TodasChecagensPassam(t *testing.T) {
	linhas := []map[string]any{
		{"tier_usado": "Tier 1", "escalou_cascata": false, "confianca_resposta": 1.0},
		{"tier_usado": "Tier 2 (escalado)", "escalou_cascata": true, "confianca_resposta": 0.83},
		{"tier_usado": "Tier 2", "escalou_cascata": false, "aprovado": true},
		{"status_final": "bloqueado_por_orcamento"},
	}

	todasOk(t, VerificarTrilha(linhas))
}

func TestVerificarTrilha_MenosDeQuatroEntradas_FalhaNaPrimeiraChecagem(t *testing.T) {
	linhas := []map[string]any{{"tier_usado": "Tier 1"}}

	checagens := VerificarTrilha(linhas)
	if checagens[0].OK {
		t.Error("esperava falhar checagem de 'pelo menos 4 entradas'")
	}
}

func TestVerificarTrilha_PrimeiroRegistroEscalou_FalhaNaChecagem1(t *testing.T) {
	linhas := []map[string]any{
		{"tier_usado": "Tier 2 (escalado)", "escalou_cascata": true}, // deveria ser Tier 1, sem escalar
		{},
		{},
		{},
	}

	checagens := VerificarTrilha(linhas)
	if checagens[1].OK {
		t.Error("esperava falhar checagem #1")
	}
}

func TestVerificarTrilha_UsaSomenteAsUltimasQuatroEntradas(t *testing.T) {
	lixo := map[string]any{"tier_usado": "qualquer coisa"}
	linhas := []map[string]any{
		lixo, lixo,
		{"tier_usado": "Tier 1", "escalou_cascata": false, "confianca_resposta": 1.0},
		{"tier_usado": "Tier 2 (escalado)", "escalou_cascata": true, "confianca_resposta": 0.83},
		{"tier_usado": "Tier 2", "escalou_cascata": false, "aprovado": true},
		{"status_final": "bloqueado_por_orcamento"},
	}

	todasOk(t, VerificarTrilha(linhas))
}
