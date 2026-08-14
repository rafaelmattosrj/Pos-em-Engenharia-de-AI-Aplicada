package gateway

// entradaCache é um item do Semantic Cache (Módulo 4.3): pergunta original,
// seu embedding e a resposta já aprovada, reaproveitada em perguntas futuras
// semanticamente parecidas.
type entradaCache struct {
	Pergunta  string
	Embedding []float64
	Resposta  string
}

// SemanticCache é o cache semântico do gateway — em memória, populado só com
// respostas de perguntas de rotina (síntese de CSR nunca usa cache).
type SemanticCache struct {
	entradas []entradaCache
}

// ResultadoConsultaCache é o retorno de Consultar: a melhor similaridade
// encontrada e a entrada correspondente (nil se o cache está vazio).
type ResultadoConsultaCache struct {
	Similaridade float64
	Entrada      *entradaCache
}

// Consultar percorre o cache inteiro e devolve a entrada de maior
// similaridade de cosseno em relação ao embedding da pergunta atual.
func (c *SemanticCache) Consultar(perguntaEmbedding []float64) ResultadoConsultaCache {
	melhor := ResultadoConsultaCache{Similaridade: 0}
	for i := range c.entradas {
		sim := SimilaridadeCosseno(perguntaEmbedding, c.entradas[i].Embedding)
		if sim > melhor.Similaridade {
			melhor = ResultadoConsultaCache{Similaridade: sim, Entrada: &c.entradas[i]}
		}
	}
	return melhor
}

// Adicionar alimenta o cache com essa pergunta+resposta pra próxima vez.
func (c *SemanticCache) Adicionar(pergunta string, embedding []float64, resposta string) {
	c.entradas = append(c.entradas, entradaCache{Pergunta: pergunta, Embedding: embedding, Resposta: resposta})
}
