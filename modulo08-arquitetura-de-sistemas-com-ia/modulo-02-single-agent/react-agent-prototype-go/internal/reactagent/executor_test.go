package reactagent

import "testing"

// Cobre os mesmos 6 casos + 1 caso de parâmetro inválido de
// rodarTestesFerramenta() em react-agent-prototype.js / .py (via
// CasosTesteFerramenta, a mesma lista usada por main.go).

func TestExecutarBuscaClausula_SeguemOsCasosDeReferencia(t *testing.T) {
	for _, caso := range CasosTesteFerramenta {
		caso := caso
		t.Run(caso.Tema+"/"+caso.Jurisdicao, func(t *testing.T) {
			resultado := ExecutarBuscaClausula(map[string]interface{}{
				"tema": caso.Tema, "jurisdicao": caso.Jurisdicao,
			})
			achou := resultado.Texto != ""
			if achou != caso.EsperaAchar {
				t.Errorf("tema=%q jurisdicao=%q: achou=%v, esperado=%v",
					caso.Tema, caso.Jurisdicao, achou, caso.EsperaAchar)
			}
		})
	}
}

func TestExecutarBuscaClausula_ParametroMalFormado_ViraAvisoDeFalhaPropria(t *testing.T) {
	resultado := ExecutarBuscaClausula(map[string]interface{}{"tema": "x", "jurisdicicao": "ANVISA"})

	if resultado.Texto != "" {
		t.Errorf("esperava texto vazio, obteve %q", resultado.Texto)
	}
	if resultado.Aviso == "" {
		t.Error("esperava aviso de parâmetro inválido")
	}
}

func TestExecutarBuscaClausula_ArgumentosNulos_NaoEntraEmPanico(t *testing.T) {
	resultado := ExecutarBuscaClausula(nil)

	if resultado.Texto != "" {
		t.Errorf("esperava texto vazio para argumentos nulos, obteve %q", resultado.Texto)
	}
	if resultado.Aviso == "" {
		t.Error("esperava aviso para argumentos nulos")
	}
}

func TestExecutarBuscaClausula_JurisdicaoCorretaSemMencaoAMenor_NaoEncontraClausula(t *testing.T) {
	resultado := ExecutarBuscaClausula(map[string]interface{}{
		"tema": "Consentimento informado de população adulta", "jurisdicao": "ANVISA",
	})

	if resultado.Texto != "" {
		t.Errorf("não esperava cláusula, obteve %q", resultado.Texto)
	}
	if resultado.Aviso == "" {
		t.Error("esperava aviso de não encontrado")
	}
}
