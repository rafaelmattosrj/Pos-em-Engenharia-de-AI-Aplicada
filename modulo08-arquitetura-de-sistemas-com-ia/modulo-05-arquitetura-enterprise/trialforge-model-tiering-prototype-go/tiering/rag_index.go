package tiering

import "context"

// RagIndex indexa e busca o banco de cláusulas (Módulo 4.1): embedding de
// cada cláusula, uma vez só — nunca de novo a cada busca.
//
// Dois índices, dois propósitos: embeddings do TEMA acham a cláusula certa
// pra pergunta (confiança de busca); embeddings do TEXTO servem de referência
// pra medir se a resposta GERADA depois fica fiel a essa cláusula (confiança
// de resposta, em CalcularConfiancaResposta).
type RagIndex struct {
	ollama          Gateway
	modeloEmbedding string
	embeddingsTema  [][]float64
	embeddingsTexto [][]float64
}

func NewRagIndex(ollama Gateway, modeloEmbedding string) *RagIndex {
	return &RagIndex{ollama: ollama, modeloEmbedding: modeloEmbedding}
}

func (r *RagIndex) PrepararIndice(ctx context.Context) error {
	r.embeddingsTema = make([][]float64, len(BancoClausulas))
	r.embeddingsTexto = make([][]float64, len(BancoClausulas))
	for i, clausula := range BancoClausulas {
		tema, err := r.ollama.Embed(ctx, r.modeloEmbedding, clausula.Tema)
		if err != nil {
			return err
		}
		texto, err := r.ollama.Embed(ctx, r.modeloEmbedding, clausula.Texto)
		if err != nil {
			return err
		}
		r.embeddingsTema[i] = tema
		r.embeddingsTexto[i] = texto
	}
	return nil
}

// ResultadoBusca é a melhor cláusula encontrada, sua similaridade e índice no banco.
type ResultadoBusca struct {
	Similaridade float64
	Clausula     Clausula
	Indice       int
}

func (r *RagIndex) BuscarClausula(perguntaEmbedding []float64) ResultadoBusca {
	melhor := ResultadoBusca{Similaridade: -1, Indice: -1}
	for i, clausula := range BancoClausulas {
		sim := SimilaridadeCosseno(perguntaEmbedding, r.embeddingsTema[i])
		if sim > melhor.Similaridade {
			melhor = ResultadoBusca{Similaridade: sim, Clausula: clausula, Indice: i}
		}
	}
	return melhor
}

func (r *RagIndex) CalcularConfiancaResposta(ctx context.Context, rascunho string, indiceClausula int) (float64, error) {
	rascunhoEmbedding, err := r.ollama.Embed(ctx, r.modeloEmbedding, rascunho)
	if err != nil {
		return 0, err
	}
	return SimilaridadeCosseno(rascunhoEmbedding, r.embeddingsTexto[indiceClausula]), nil
}
