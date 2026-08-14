// Package memory implementa os 4 tipos de memória do agente — equivalente
// ao pacote memory/ da versão Java.
package memory

// ShortTermMemory armazena estado temporário da execução atual — equivalente
// a ShortTermMemory.java (escopo prototype: cada execução cria sua própria
// instância via NewShortTermMemory, garantindo isolamento).
type ShortTermMemory struct {
	store map[string]any
}

// NewShortTermMemory cria uma instância isolada de memória de curto prazo.
func NewShortTermMemory() *ShortTermMemory {
	return &ShortTermMemory{store: make(map[string]any)}
}

// Put armazena um valor associado à chave.
func (m *ShortTermMemory) Put(key string, value any) {
	m.store[key] = value
}

// Get recupera um valor pela chave, ou nil se não encontrado.
func (m *ShortTermMemory) Get(key string) any {
	return m.store[key]
}

// All retorna todo o estado atual.
func (m *ShortTermMemory) All() map[string]any {
	return m.store
}

// Clear limpa todo o estado — chamado ao fim da execução.
func (m *ShortTermMemory) Clear() {
	m.store = make(map[string]any)
}
