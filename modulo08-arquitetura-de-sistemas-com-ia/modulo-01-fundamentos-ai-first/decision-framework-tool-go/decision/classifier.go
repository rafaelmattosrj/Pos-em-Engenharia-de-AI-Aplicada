// Package decision implementa o Framework de Decisao de Tres Perguntas
// (checklist do Modulo 1.3), portado de decision-framework-tool.js /
// decision_framework_tool.py (pasta ../.. do modulo original,
// modulo-01-fundamentos-ai-first).
//
// Logica pura, 100% deterministica: nao chama nenhum modelo, nao faz
// nenhuma chamada de rede. E a arvore de decisao do checklist
// (decision-framework-checklist.md) virando codigo, pergunta por pergunta.
package decision

// Classificacao e uma das quatro classificacoes possiveis do checklist,
// exatamente como o texto original (decision-framework-checklist.md).
// Definida como um tipo string nomeado (idiomatico em Go), com o texto
// exato como valor, em vez de um objeto de constantes como no JS/Python —
// preserva a mesma string observavel quando impressa ou comparada.
type Classificacao string

const (
	// RegraDeterministica: existe uma regra finita que cobre mais de 90%
	// dos casos reais.
	RegraDeterministica Classificacao = "Regra determinística"
	// ApprovalGateObrigatorio: o erro e caro e a acao e irreversivel.
	ApprovalGateObrigatorio Classificacao = "Agente com Approval Gate obrigatório"
	// AgenteAutonomo: o comportamento muda de acordo com o contexto de
	// entrada (e nao ha regra finita nem risco caro/irreversivel).
	AgenteAutonomo Classificacao = "Agente autônomo, com observabilidade completa"
	// RegraEnumeravel: nenhuma das perguntas anteriores se aplica; mesmo
	// parecendo complexa, se e enumeravel, e regra.
	RegraEnumeravel Classificacao = "Regra determinística (mesmo parecendo complexa, se é enumerável, é regra)"
)

// ClassificarTarefa aplica as tres perguntas do checklist, em ordem,
// exatamente como a arvore: Pergunta 1 decide sozinha quando e true (nem
// chega a olhar p2 ou p3); senao passa para a Pergunta 2; senao passa para
// a Pergunta 3.
//
//   - p1: Existe uma regra finita que cobre mais de 90% dos casos REAIS (ja
//     observados, nao hipoteticos)?
//   - p2: O erro e caro E a acao e irreversivel (nao da pra desfazer depois)?
//   - p3: O comportamento da tarefa muda de acordo com o contexto de entrada?
func ClassificarTarefa(p1, p2, p3 bool) Classificacao {
	if p1 {
		return RegraDeterministica
	}
	if p2 {
		return ApprovalGateObrigatorio
	}
	if p3 {
		return AgenteAutonomo
	}
	return RegraEnumeravel
}

// Subtarefa e uma subtarefa ainda nao classificada, seguindo o template
// "Decompondo uma tarefa hibrida" do checklist: nome livre, tipo livre
// (ex.: "Extração/Interpretação" ou "Decisão de Negócio") e as tres
// respostas booleanas do framework de decisao.
type Subtarefa struct {
	Nome string
	Tipo string
	P1   bool
	P2   bool
	P3   bool
}

// SubtarefaClassificada e o mesmo conteudo de Subtarefa, com o campo a mais
// Classificacao — resultado de aplicar ClassificarTarefa naquela subtarefa
// especifica.
type SubtarefaClassificada struct {
	Nome          string
	Tipo          string
	P1            bool
	P2            bool
	P3            bool
	Classificacao Classificacao
}

// DecomporTarefaHibrida decompoe uma tarefa hibrida em subtarefas ja
// classificadas, seguindo o "Template: decompondo uma tarefa hibrida" do
// checklist. Nao e uma logica nova: aplica ClassificarTarefa em cada
// subtarefa, uma de cada vez, e devolve um slice novo com o campo
// Classificacao preenchido — o slice/structs de entrada nao sao modificados
// (equivalente ao comportamento explicitamente documentado nas versoes
// JS/Python, que tambem nao mutam a entrada).
func DecomporTarefaHibrida(subtarefas []Subtarefa) []SubtarefaClassificada {
	resultado := make([]SubtarefaClassificada, 0, len(subtarefas))
	for _, subtarefa := range subtarefas {
		resultado = append(resultado, SubtarefaClassificada{
			Nome:          subtarefa.Nome,
			Tipo:          subtarefa.Tipo,
			P1:            subtarefa.P1,
			P2:            subtarefa.P2,
			P3:            subtarefa.P3,
			Classificacao: ClassificarTarefa(subtarefa.P1, subtarefa.P2, subtarefa.P3),
		})
	}
	return resultado
}
