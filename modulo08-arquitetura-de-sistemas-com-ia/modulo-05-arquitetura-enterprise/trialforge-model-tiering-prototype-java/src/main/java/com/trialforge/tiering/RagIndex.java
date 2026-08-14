package com.trialforge.tiering;

import java.io.IOException;
import java.util.List;

/**
 * Indexação e busca do banco de cláusulas (Módulo 4.1): embedding de cada
 * cláusula, uma vez só — nunca de novo a cada busca.
 *
 * Dois índices, dois propósitos: embeddings do TEMA acham a cláusula certa
 * pra pergunta (confiança de busca); embeddings do TEXTO servem de referência
 * pra medir se a resposta GERADA depois fica fiel a essa cláusula (confiança
 * de resposta, em {@link #calcularConfiancaResposta}).
 */
public class RagIndex {

    // Banco de cláusulas do Agente ICF (mesmo dado do Módulo 2.5/4.5/5.4).
    public static final List<Clausula> BANCO_CLAUSULAS = List.of(
            new Clausula(
                    "Assentimento de menores de idade em estudos clínicos",
                    "Para participantes entre 12 e 17 anos, é necessário assentimento por escrito, "
                            + "além do consentimento do responsável legal (RDC ANVISA 466/2012, Art. 4º).",
                    "RDC ANVISA 466/2012, Art. 4º"),
            new Clausula(
                    "Direito de retirada do participante do estudo a qualquer momento",
                    "O participante pode retirar seu consentimento a qualquer momento, sem necessidade "
                            + "de justificativa e sem prejuízo ao seu tratamento (RDC ANVISA 466/2012, Art. 5º).",
                    "RDC ANVISA 466/2012, Art. 5º"));

    private final OllamaGateway ollama;
    private final String modeloEmbedding;
    private double[][] embeddingsTema;
    private double[][] embeddingsTexto;

    public RagIndex(OllamaGateway ollama, String modeloEmbedding) {
        this.ollama = ollama;
        this.modeloEmbedding = modeloEmbedding;
    }

    public void prepararIndice() throws IOException, InterruptedException {
        embeddingsTema = new double[BANCO_CLAUSULAS.size()][];
        embeddingsTexto = new double[BANCO_CLAUSULAS.size()][];
        for (int i = 0; i < BANCO_CLAUSULAS.size(); i++) {
            Clausula clausula = BANCO_CLAUSULAS.get(i);
            embeddingsTema[i] = ollama.embed(modeloEmbedding, clausula.tema());
            embeddingsTexto[i] = ollama.embed(modeloEmbedding, clausula.texto());
        }
    }

    /** Resultado da busca: melhor cláusula encontrada, sua similaridade e índice no banco. */
    public record ResultadoBusca(double similaridade, Clausula clausula, int indice) {
    }

    public ResultadoBusca buscarClausula(double[] perguntaEmbedding) {
        double melhorSimilaridade = -1;
        Clausula melhorClausula = null;
        int melhorIndice = -1;
        for (int i = 0; i < BANCO_CLAUSULAS.size(); i++) {
            double sim = VectorMath.similaridadeCosseno(perguntaEmbedding, embeddingsTema[i]);
            if (sim > melhorSimilaridade) {
                melhorSimilaridade = sim;
                melhorClausula = BANCO_CLAUSULAS.get(i);
                melhorIndice = i;
            }
        }
        return new ResultadoBusca(melhorSimilaridade, melhorClausula, melhorIndice);
    }

    public double calcularConfiancaResposta(String rascunho, int indiceClausula) throws IOException, InterruptedException {
        double[] rascunhoEmbedding = ollama.embed(modeloEmbedding, rascunho);
        return VectorMath.similaridadeCosseno(rascunhoEmbedding, embeddingsTexto[indiceClausula]);
    }
}
