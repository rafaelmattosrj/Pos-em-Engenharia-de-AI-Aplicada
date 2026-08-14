package reactagent

import (
	"fmt"
	"strings"
)

// palavrasMenorDeIdade: testado de verdade, o modelo as vezes formula o tema
// com "adolescente" ou "pediátrico" em vez de "menor" — um match de palavra
// unica era fragil demais pra depender da sorte da frase exata que o modelo
// escolhe.
var palavrasMenorDeIdade = []string{"menor", "adolescente", "pediátric"}

// ExecutarBuscaClausula e a execucao real da ferramenta — sempre
// deterministica, nunca e o modelo quem executa. Busca por palavra-chave,
// nao string exata: o modelo formula o "tema" em linguagem natural, entao a
// busca real por tras dessa ferramenta seria vetorial (RAG) — aqui
// simulamos isso com um match por palavra-chave em vez de comparacao exata.
//
// Recebe um map bruto (equivalente aos **kwargs do Python / objeto solto do
// JS) em vez de (tema, jurisdicao) fixos de proposito: o modelo as vezes
// formula a chamada com um parametro faltando ou grafado errado (testado de
// verdade: ja vimos "jurisdicicao" em vez de "jurisdicao"). Validar aqui, em
// vez de confiar cegamente, evita um resultado errado escapar sem ninguem
// perceber — a ferramenta trata isso como falha propria, nunca deixa o erro
// estourar pro chamador.
//
// Porte 1:1 de executarBuscaClausula em react-agent-prototype.js / .py.
func ExecutarBuscaClausula(argumentosBrutos map[string]interface{}) ResultadoBusca {
	temaObj, temaOk := argumentosBrutos["tema"]
	jurisdicaoObj, jurisdicaoOk := argumentosBrutos["jurisdicao"]

	tema, temaEhString := temaObj.(string)
	jurisdicao, jurisdicaoEhString := jurisdicaoObj.(string)

	if !temaOk || !jurisdicaoOk || !temaEhString || !jurisdicaoEhString {
		return ResultadoBusca{
			Aviso: fmt.Sprintf("parâmetro inválido ou ausente na chamada da ferramenta: %v", argumentosBrutos),
		}
	}

	temaNormalizado := strings.ToLower(tema)
	mencionaMenorDeIdade := false
	for _, palavra := range palavrasMenorDeIdade {
		if strings.Contains(temaNormalizado, palavra) {
			mencionaMenorDeIdade = true
			break
		}
	}

	if jurisdicao == "ANVISA" && mencionaMenorDeIdade {
		return ResultadoBusca{
			Texto: "Para participantes entre 12 e 17 anos, é necessário assentimento por escrito, além do " +
				"consentimento do responsável legal (RDC ANVISA 466/2012, Art. 4º).",
			Fonte: "RDC ANVISA 466/2012, Art. 4º",
		}
	}

	return ResultadoBusca{
		Aviso: fmt.Sprintf(
			"não encontrado: nenhuma cláusula sobre \"%s\" na jurisdição %s. Tente outra jurisdição ou revise o tema.",
			tema, jurisdicao),
	}
}
