package gateway

// Indices (Módulo 4.1 — Multi-Index): um índice por domínio de agente do
// TrialForge — ICF, Protocolo, CSR — em vez de um banco único e achatado.
// Cada domínio tem vocabulário e fonte regulatória próprios (Módulo 3.1: é o
// mesmo motivo que justifica agentes especialistas), então misturar os três
// num índice só só adicionaria ruído de recuperação sem necessidade.
var Indices = map[string][]Clausula{
	"icf": {
		{
			Tema: "Assentimento de menores de idade em estudos clínicos",
			Texto: "Para participantes entre 12 e 17 anos, é necessário assentimento por escrito, " +
				"além do consentimento do responsável legal (RDC ANVISA 466/2012, Art. 4º).",
			Fonte: "RDC ANVISA 466/2012, Art. 4º",
		},
		{
			Tema: "Direito de retirada do participante do estudo a qualquer momento",
			Texto: "O participante pode retirar seu consentimento a qualquer momento, sem necessidade " +
				"de justificativa e sem prejuízo ao seu tratamento (RDC ANVISA 466/2012, Art. 5º).",
			Fonte: "RDC ANVISA 466/2012, Art. 5º",
		},
	},
	// "doze anos" aqui é o mesmo número da emenda ética do Módulo 3.4/3.5 (idade
	// mínima 13 → 12) — o mesmo critério de protocolo revisitado em outra camada.
	"protocolo": {
		{
			Tema: "Critério de idade mínima para inclusão no estudo",
			Texto: "A idade mínima para participação no estudo é de doze anos completos na data " +
				"da assinatura do assentimento, conforme a versão vigente do protocolo aprovada " +
				"pelo comitê de ética.",
			Fonte: "Protocolo Clínico TrialForge, critério de inclusão nº 2",
		},
		{
			Tema: "Critério de exclusão por interação medicamentosa",
			Texto: "Participantes em uso concomitante de medicação com interação farmacológica " +
				"documentada são excluídos do estudo, conforme a seção 4.2 do protocolo.",
			Fonte: "Protocolo Clínico TrialForge, critério de exclusão nº 4",
		},
	},
	"csr": {
		{
			Tema: "Apresentação de eventos adversos no relatório final",
			Texto: "Eventos adversos devem ser apresentados por gravidade e causalidade, seguindo " +
				"a estrutura de seções do guia ICH E3, sem agregação que oculte eventos " +
				"individuais graves.",
			Fonte: "ICH E3, seção 12",
		},
		{
			Tema: "Significância estatística do desfecho primário",
			Texto: "O desfecho primário só pode ser reportado como positivo se atingir o nível de " +
				"significância pré-especificado no plano de análise estatística, sem ajuste " +
				"post-hoc.",
			Fonte: "ICH E3, seção 11; Plano de Análise Estatística do estudo",
		},
	},
}

// NomesIndices fixa a ordem de iteração dos índices (Go não garante ordem de
// map) — equivalente a Object.keys(INDICES)/INDICES.keys() em JS/Python, que
// preservam a ordem de inserção. A ordem importa em buscarEmTodosIndices: em
// caso de empate de similaridade, o primeiro índice da lista vence.
var NomesIndices = []string{"icf", "protocolo", "csr"}

// IndicePorIntencao roteia pro índice certo (Módulo 4.1) reaproveitando a
// mesma classificação de intenção que já decide o Model Router (Módulo 4.2).
var IndicePorIntencao = map[string]string{
	"consulta_icf":       "icf",
	"consulta_protocolo": "protocolo",
	"sintese_csr":        "csr",
}
