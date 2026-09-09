package com.datasetprep;

import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Porte de data-relevance-scoring-tool.js (Modulo 2.1).
 *
 * Gate de 4 criterios para decidir se um candidato a fonte de dado vale a
 * pena entrar no pipeline de extracao, antes de coletar/extrair qualquer
 * documento. Fundamento: Data-Centric AI (Andrew Ng) e o paper LIMA
 * (Zhou et al. 2023, arXiv 2305.11206).
 */
public final class DataRelevance {

    private DataRelevance() {
    }

    public static final Map<String, String> CRITERIOS = new LinkedHashMap<>();
    static {
        CRITERIOS.put("contemGroundTruth", "Contem o ground truth da tarefa (entrada real + resposta correta observavel)?");
        CRITERIOS.put("producaoReal", "Vem do fluxo real de producao, nao e exemplo sintetico ou hipotetico?");
        CRITERIOS.put("cobreVariacao", "Cobre a variacao real de formato e situacao que a tarefa tem, nao so o caso facil?");
        CRITERIOS.put("passaCompliance", "Passa no crivo de sensibilidade e compliance pra ser usado em treino?");
    }

    public record Candidato(
            String id,
            String caso,
            String nome,
            Map<String, Boolean> criterios,
            String justificativa
    ) {
    }

    public record Avaliacao(boolean aceito, List<String> criteriosFalhos) {
    }

    /** Gate estrito: os 4 criterios precisam ser verdadeiros para o candidato ser aceito. */
    public static Avaliacao avaliarCandidato(Map<String, Boolean> criterios) {
        List<String> falhas = CRITERIOS.keySet().stream()
                .filter(chave -> !Boolean.TRUE.equals(criterios.get(chave)))
                .toList();
        return new Avaliacao(falhas.isEmpty(), falhas);
    }

    public static final List<Candidato> CANDIDATOS = List.of(
            new Candidato("orcamento-oficina", "amplitude-auto", "Orcamento de oficina",
                    Map.of("contemGroundTruth", true, "producaoReal", true, "cobreVariacao", true, "passaCompliance", true),
                    "Contem segurado, placa e valor em texto real, vem de sinistros ja processados, varia de formato entre oficinas, sem dado sensivel."),
            new Candidato("boletim-ocorrencia", "amplitude-auto", "Boletim de ocorrencia policial",
                    Map.of("contemGroundTruth", false, "producaoReal", true, "cobreVariacao", true, "passaCompliance", true),
                    "E narrativa de ocorrencia, nao traz segurado/placa/valor em formato consistente o bastante pra servir de resposta correta da tarefa."),
            new Candidato("foto-veiculo-danificado", "amplitude-auto", "Foto do veiculo danificado",
                    Map.of("contemGroundTruth", false, "producaoReal", true, "cobreVariacao", false, "passaCompliance", true),
                    "E imagem do dano, nao do documento com os campos-alvo; nao serve de par texto-entrada/texto-saida pra esta tarefa de extracao."),
            new Candidato("transcricao-ligacao", "amplitude-auto", "Transcricao da ligacao de abertura de sinistro",
                    Map.of("contemGroundTruth", true, "producaoReal", true, "cobreVariacao", false, "passaCompliance", true),
                    "Tem os dados, mas a estrutura da ligacao varia demais de atendente pra atendente pra servir de exemplo confiavel de formato."),
            new Candidato("recibo-medico", "amplitude-saude-empresarial", "Recibo medico",
                    Map.of("contemGroundTruth", true, "producaoReal", true, "cobreVariacao", true, "passaCompliance", true),
                    "Contem beneficiario, procedimento e valor, vem de reembolsos ja processados, varia entre clinicas, e o dado sensivel e tratavel com o rastro de auditoria ja desenhado no schema."),
            new Candidato("prontuario-medico-completo", "amplitude-saude-empresarial", "Prontuario medico completo",
                    Map.of("contemGroundTruth", true, "producaoReal", true, "cobreVariacao", true, "passaCompliance", false),
                    "Traz dado clinico muito alem do necessario pra extrair beneficiario/procedimento/valor; viola minimizacao de dado exigida pela LGPD pra essa tarefa especifica."),
            new Candidato("cadastro-beneficiarios", "amplitude-saude-empresarial", "Cadastro de beneficiarios",
                    Map.of("contemGroundTruth", false, "producaoReal", true, "cobreVariacao", true, "passaCompliance", true),
                    "E tabela de referencia (titular/dependente), nao um par entrada-saida de extracao; util pra cruzamento, nao pra treinar o extrator em si.")
    );
}
