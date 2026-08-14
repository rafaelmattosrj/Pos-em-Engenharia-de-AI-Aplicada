package trialforge

import "sync"

// OrdemDeExecucao e um log thread-safe de strings — usado para provar
// causalmente que o Sequential (Protocolo primeiro) foi respeitado antes do
// Parallel (ICF/CSR) comecar. E acessado por goroutines diferentes.
type OrdemDeExecucao struct {
	mu    sync.Mutex
	itens []string
}

// Registrar acrescenta um item ao log.
func (o *OrdemDeExecucao) Registrar(item string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.itens = append(o.itens, item)
}

// Itens devolve uma COPIA do log, na ordem real de chegada.
func (o *OrdemDeExecucao) Itens() []string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return append([]string(nil), o.itens...)
}

// DocumentoRegistro e o porte de "registro idempotente" do JS/Python: a
// mesma chave nunca gera um segundo documento — so incrementa o contador de
// tentativas, que e o que permite um retry seguro.
type DocumentoRegistro struct {
	Resultado  ResultadoICFBase
	Tentativas int
}

// DocumentosGerados e o "banco de documentos" thread-safe do fluxo — escopo
// por fluxo, nao compartilhado entre estudos.
type DocumentosGerados struct {
	mu   sync.Mutex
	docs map[string]*DocumentoRegistro
}

// NovoDocumentosGerados cria um banco de documentos vazio.
func NovoDocumentosGerados() *DocumentosGerados {
	return &DocumentosGerados{docs: make(map[string]*DocumentoRegistro)}
}

// Registrar grava o documento se a chave ainda nao existir; se ja existir,
// so incrementa o contador de tentativas — nunca duplica o documento.
func (d *DocumentosGerados) Registrar(chave string, resultado ResultadoICFBase) *DocumentoRegistro {
	d.mu.Lock()
	defer d.mu.Unlock()
	if existente, ok := d.docs[chave]; ok {
		existente.Tentativas++
		return existente
	}
	registro := &DocumentoRegistro{Resultado: resultado, Tentativas: 1}
	d.docs[chave] = registro
	return registro
}

// Get devolve o registro da chave, se existir.
func (d *DocumentosGerados) Get(chave string) (*DocumentoRegistro, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	r, ok := d.docs[chave]
	return r, ok
}

// Tamanho devolve quantos documentos distintos foram gravados.
func (d *DocumentosGerados) Tamanho() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.docs)
}
