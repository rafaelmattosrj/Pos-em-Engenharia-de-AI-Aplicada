// Package datarelevance porta data-relevance-scoring-tool.js (Modulo 2.1):
// gate de 4 criterios para decidir se um candidato a fonte de dado vale a
// pena entrar no pipeline de extracao. Fundamento: Data-Centric AI (Andrew
// Ng) e o paper LIMA (Zhou et al. 2023, arXiv 2305.11206).
package datarelevance

var Criterios = map[string]string{
	"contemGroundTruth": "Contem o ground truth da tarefa (entrada real + resposta correta observavel)?",
	"producaoReal":      "Vem do fluxo real de producao, nao e exemplo sintetico ou hipotetico?",
	"cobreVariacao":     "Cobre a variacao real de formato e situacao que a tarefa tem, nao so o caso facil?",
	"passaCompliance":   "Passa no crivo de sensibilidade e compliance pra ser usado em treino?",
}

var ChavesCriterios = []string{"contemGroundTruth", "producaoReal", "cobreVariacao", "passaCompliance"}

type Candidato struct {
	ID            string
	Caso          string
	Nome          string
	Criterios     map[string]bool
	Justificativa string
}

type Avaliacao struct {
	Aceito          bool
	CriteriosFalhos []string
}

// AvaliarCandidato aplica o gate estrito: os 4 criterios precisam ser
// verdadeiros para o candidato ser aceito.
func AvaliarCandidato(criterios map[string]bool) Avaliacao {
	var falhas []string
	for _, chave := range ChavesCriterios {
		if !criterios[chave] {
			falhas = append(falhas, chave)
		}
	}
	return Avaliacao{Aceito: len(falhas) == 0, CriteriosFalhos: falhas}
}

var Candidatos = []Candidato{
	{
		ID: "orcamento-oficina", Caso: "amplitude-auto", Nome: "Orcamento de oficina",
		Criterios:     map[string]bool{"contemGroundTruth": true, "producaoReal": true, "cobreVariacao": true, "passaCompliance": true},
		Justificativa: "Contem segurado, placa e valor em texto real, vem de sinistros ja processados, varia de formato entre oficinas, sem dado sensivel.",
	},
	{
		ID: "boletim-ocorrencia", Caso: "amplitude-auto", Nome: "Boletim de ocorrencia policial",
		Criterios:     map[string]bool{"contemGroundTruth": false, "producaoReal": true, "cobreVariacao": true, "passaCompliance": true},
		Justificativa: "E narrativa de ocorrencia, nao traz segurado/placa/valor em formato consistente o bastante pra servir de resposta correta da tarefa.",
	},
	{
		ID: "foto-veiculo-danificado", Caso: "amplitude-auto", Nome: "Foto do veiculo danificado",
		Criterios:     map[string]bool{"contemGroundTruth": false, "producaoReal": true, "cobreVariacao": false, "passaCompliance": true},
		Justificativa: "E imagem do dano, nao do documento com os campos-alvo; nao serve de par texto-entrada/texto-saida pra esta tarefa de extracao.",
	},
	{
		ID: "transcricao-ligacao", Caso: "amplitude-auto", Nome: "Transcricao da ligacao de abertura de sinistro",
		Criterios:     map[string]bool{"contemGroundTruth": true, "producaoReal": true, "cobreVariacao": false, "passaCompliance": true},
		Justificativa: "Tem os dados, mas a estrutura da ligacao varia demais de atendente pra atendente pra servir de exemplo confiavel de formato.",
	},
	{
		ID: "recibo-medico", Caso: "amplitude-saude-empresarial", Nome: "Recibo medico",
		Criterios:     map[string]bool{"contemGroundTruth": true, "producaoReal": true, "cobreVariacao": true, "passaCompliance": true},
		Justificativa: "Contem beneficiario, procedimento e valor, vem de reembolsos ja processados, varia entre clinicas, e o dado sensivel e tratavel com o rastro de auditoria ja desenhado no schema.",
	},
	{
		ID: "prontuario-medico-completo", Caso: "amplitude-saude-empresarial", Nome: "Prontuario medico completo",
		Criterios:     map[string]bool{"contemGroundTruth": true, "producaoReal": true, "cobreVariacao": true, "passaCompliance": false},
		Justificativa: "Traz dado clinico muito alem do necessario pra extrair beneficiario/procedimento/valor; viola minimizacao de dado exigida pela LGPD pra essa tarefa especifica.",
	},
	{
		ID: "cadastro-beneficiarios", Caso: "amplitude-saude-empresarial", Nome: "Cadastro de beneficiarios",
		Criterios:     map[string]bool{"contemGroundTruth": false, "producaoReal": true, "cobreVariacao": true, "passaCompliance": true},
		Justificativa: "E tabela de referencia (titular/dependente), nao um par entrada-saida de extracao; util pra cruzamento, nao pra treinar o extrator em si.",
	},
}
