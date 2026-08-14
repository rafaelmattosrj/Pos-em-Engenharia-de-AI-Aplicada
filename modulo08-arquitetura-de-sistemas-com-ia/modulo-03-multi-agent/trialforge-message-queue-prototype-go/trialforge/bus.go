package trialforge

import "sync"

// Ouvinte e o callback registrado via Once — recebe o DadoProtocolo publicado.
type Ouvinte func(DadoProtocolo)

// Barramento e o porte minimo do EventEmitter nativo do Node.js (JS) / da
// classe Barramento sobre asyncio (Python) — a fila de mensagens do
// prototipo (Slide 3). So o necessario para este demo: Once (ouvinte de
// disparo unico por evento) e Emit (publica e segue em frente, sem
// bloquear quem publicou).
type Barramento struct {
	mu       sync.Mutex
	ouvintes map[string]Ouvinte
}

// NovoBarramento cria uma fila de mensagens vazia.
func NovoBarramento() *Barramento {
	return &Barramento{ouvintes: make(map[string]Ouvinte)}
}

// Once registra um ouvinte de disparo unico para o evento — igual a
// EventEmitter#once (JS) / Barramento.once (Python).
func (b *Barramento) Once(evento string, ouvinte Ouvinte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.ouvintes[evento] = ouvinte
}

// Emit publica o dado no evento e remove o ouvinte de disparo unico. O
// ouvinte roda em sua propria goroutine — nao-bloqueante: publica e segue em
// frente, igual ao emit do JS e ao asyncio.create_task do Python.
func (b *Barramento) Emit(evento string, dado DadoProtocolo) {
	b.mu.Lock()
	ouvinte, ok := b.ouvintes[evento]
	if ok {
		delete(b.ouvintes, evento)
	}
	b.mu.Unlock()

	if ok {
		go ouvinte(dado)
	}
}
