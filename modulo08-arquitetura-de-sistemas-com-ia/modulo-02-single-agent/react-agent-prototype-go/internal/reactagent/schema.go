package reactagent

// NomeFerramenta e o nome da unica ferramenta tipada usada pelo loop ReAct.
const NomeFerramenta = "buscar_clausula_regulatoria"

const descricaoFerramenta = "Busca cláusulas regulatórias de estudos clínicos por tema e jurisdição. " +
	"Use quando precisar de texto normativo (ANVISA ou FDA) para compor uma seção do documento."

// BuscarClausulaRegulatoria monta o schema da ferramenta no formato nativo
// do Ollama (igual ao formato aninhado do OpenAI): {"type":"function",
// "function":{"name",...,"parameters":{...}}}.
//
// Porte 1:1 de BUSCAR_CLAUSULA_REGULATORIA em react-agent-prototype.js / .py.
func BuscarClausulaRegulatoria() Mensagem {
	return Mensagem{
		"type": "function",
		"function": map[string]interface{}{
			"name":        NomeFerramenta,
			"description": descricaoFerramenta,
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"tema": map[string]interface{}{"type": "string"},
					"jurisdicao": map[string]interface{}{
						"type": "string",
						"enum": []string{"ANVISA", "FDA"},
					},
				},
				"required": []string{"tema", "jurisdicao"},
			},
		},
	}
}
