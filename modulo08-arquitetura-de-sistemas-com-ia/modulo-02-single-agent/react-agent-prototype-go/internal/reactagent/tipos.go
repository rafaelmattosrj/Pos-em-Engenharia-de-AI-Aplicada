// Package reactagent implementa o loop ReAct (Pensamento -> Acao -> Observacao
// -> Resposta Final) do agente unico do TrialForge (geracao da secao de
// assentimento do ICF), portado de react-agent-prototype.js / .py.
package reactagent

import "context"

// Mensagem e uma mensagem de chat (ou um schema de ferramenta) representada
// como um mapa dinamico — mesma flexibilidade que JS/Python tem nativamente
// com objetos/dicionarios. O loop precisa reempurrar a mensagem bruta do
// assistente (com tool_calls) de volta no historico exatamente como veio do
// Ollama, entao um tipo fortemente tipado por papel exigiria uma uniao de
// tipos mais pesada so pra reproduzir o mesmo comportamento observavel.
type Mensagem = map[string]interface{}

// ChamadaFerramenta e uma chamada de ferramenta decidida pelo modelo —
// equivalente a resposta.message.tool_calls[0] no original.
type ChamadaFerramenta struct {
	ID         string
	Nome       string
	Argumentos map[string]interface{}
}

// ModeloResposta e a resposta do modelo pra uma volta do loop ReAct.
//
// MensagemBruta guarda o JSON exato devolvido pelo Ollama (equivalente a
// resposta.message no original) — precisa ser reempurrado pro historico tal
// como veio, tool_calls inclusos, pra proxima volta do loop fazer sentido
// pro modelo.
type ModeloResposta struct {
	MensagemBruta Mensagem
	Content       string
	Chamadas      []ChamadaFerramenta
}

// TemChamadaFerramenta indica se o modelo decidiu chamar a ferramenta em vez
// de responder direto.
func (r ModeloResposta) TemChamadaFerramenta() bool {
	return len(r.Chamadas) > 0
}

// ChatClient e a abstracao da chamada "Pensamento" do loop ReAct: envia o
// historico + o schema de ferramentas disponiveis, recebe de volta ou uma
// resposta final, ou uma chamada de ferramenta a executar.
//
// Extraida como interface (o original em JS/Python chama o SDK do Ollama
// diretamente) para permitir testar AgenteICF com um dublê determinístico,
// sem depender de um Ollama local rodando durante `go test`.
// OllamaReactClient e a implementacao real; RetryingChatClient decora
// qualquer implementacao com retry+backoff.
type ChatClient interface {
	Chat(ctx context.Context, historico []Mensagem, tools []Mensagem) (ModeloResposta, error)
}

// ResultadoBusca e o retorno de ExecutarBuscaClausula: ou tem Texto+Fonte
// (clausula encontrada), ou tem Aviso (nao encontrada / parametro invalido).
type ResultadoBusca struct {
	Texto string `json:"texto"`
	Fonte string `json:"fonte"`
	Aviso string `json:"aviso,omitempty"`
}
