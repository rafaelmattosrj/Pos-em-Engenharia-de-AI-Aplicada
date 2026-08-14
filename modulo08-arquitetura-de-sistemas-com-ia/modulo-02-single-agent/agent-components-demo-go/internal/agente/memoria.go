// Package agente contem as cinco pecas da anatomia de um agente unico
// (Memoria, Planejamento, Ferramentas, Acao, Approval Gate), portadas de
// agent-components-demo.js / agent_components_demo.py.
package agente

// Mensagem e uma mensagem de chat simples {role, content} — equivalente ao
// objeto literal do original.
type Mensagem struct {
	Role    string
	Content string
}

// MemoriaCurtoPrazo vive so durante esta chamada: some quando a funcao
// termina, nunca e reaproveitada pelo proximo protocolo que chegar.
func MemoriaCurtoPrazo(protocolo string) []Mensagem {
	return []Mensagem{
		{Role: "system", Content: "Você redige seções de ICF para estudos clínicos."},
		{Role: "user", Content: "Protocolo: " + protocolo},
	}
}

// Historico e o snapshot imutavel do historico acumulado de um usuario.
type Historico struct {
	Interacoes   int
	Preferencias []string
}

// MemoriaLongoPrazo simula um "banco" simples pra mostrar o principio de
// memoria persistida entre chamadas — em producao seria um banco de dados
// de verdade.
//
// Adaptacao: no original (JS/Python) o "banco" e um Map/dict em escopo de
// modulo, compartilhado por todas as chamadas do processo, e os testes
// chamam .clear() antes de rodar. Aqui isso vira uma struct instanciavel:
// cada MemoriaLongoPrazo tem seu proprio mapa, entao "resetar entre testes"
// e so criar uma instancia nova (NovaMemoriaLongoPrazo()) — mesmo
// comportamento observado, sem depender de estado global mutavel.
type MemoriaLongoPrazo struct {
	bancoDeUsuarios map[string]Historico
}

// NovaMemoriaLongoPrazo cria um banco de memoria de longo prazo vazio.
func NovaMemoriaLongoPrazo() *MemoriaLongoPrazo {
	return &MemoriaLongoPrazo{bancoDeUsuarios: make(map[string]Historico)}
}

// Registrar acumula uma nova interacao (e, opcionalmente, uma preferencia)
// para o usuarioId informado, e devolve o historico atualizado.
func (m *MemoriaLongoPrazo) Registrar(usuarioID string, preferenciaNova string) Historico {
	atual, ok := m.bancoDeUsuarios[usuarioID]
	if !ok {
		atual = Historico{Interacoes: 0, Preferencias: nil}
	}

	novasPreferencias := atual.Preferencias
	if preferenciaNova != "" {
		copia := make([]string, len(atual.Preferencias), len(atual.Preferencias)+1)
		copy(copia, atual.Preferencias)
		novasPreferencias = append(copia, preferenciaNova)
	}

	novo := Historico{
		Interacoes:   atual.Interacoes + 1,
		Preferencias: novasPreferencias,
	}
	m.bancoDeUsuarios[usuarioID] = novo
	return novo
}
