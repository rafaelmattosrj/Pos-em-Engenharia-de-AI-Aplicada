package com.amplitudeseguros.decisionframework.grpo;

import java.util.ArrayList;
import java.util.List;
import java.util.Locale;
import java.util.Map;

/**
 * Demo do GRPO (Group Relative Policy Optimization): amostragem real de grupo
 * via Ollama local -> recompensa verificável -> vantagem relativa ao grupo.
 * NÃO treina nada -- reproduz só o sinal de aprendizado. Equivalente ao main()
 * de grpo-verifiable-reward-demo.js.
 */
public final class GrpoDemoMain {

    private static final String MODELO = "gemma4:e2b";
    private static final int TAMANHO_GRUPO = 6;
    private static final double TEMPERATURA = 1.0;

    private static final String DOCUMENTO_SINISTRO = """
            Amplitude Seguros - Comunicado de Sinistro (documento digitalizado, OCR, scan parcialmente ilegivel)
            Ref. anterior (cancelada): AS-2Q25-0988zi
            Apolice vigente: A5-2026-1l44/7 (numero de dificil leitura no scan)
            Segurado: M4rcos Vin1cius Alm3ida Te1xeira (OCR com ruido no nome)
            Tipo de sinistro: Colisao veicular (ou possivelmente "Colisao e furto", trecho cortado)
            Data do incidente: 12/O7/2026 ou 17/02/2026 (data ambigua, dois carimbos sobrepostos)
            Valor estimado do reparo: R$ 8.45O,OO (sujeito a pericia, valor preliminar)
            Status atual: Em analise pericial
            """;

    private static final String PROMPT_SCHEMA = """
            Extraia os dados do documento abaixo e responda APENAS com um JSON
            valido, sem nenhum texto adicional, seguindo exatamente este schema:
            {"claimant_name": string, "claim_type": string, "incident_date": "DD/MM/AAAA",
            "estimated_amount_brl": number, "status": string,
            "prioridade": "alta se estimated_amount_brl > 5000, senao baixa"}

            Documento:
            %s
            """.formatted(DOCUMENTO_SINISTRO);

    public static void main(String[] args) {
        OllamaClient ollama = new OllamaClient();

        System.out.println("== Amostragem real de grupo (o passo que o GRPO chama de 'rollout') ==");
        System.out.printf("Modelo: %s  |  tamanho do grupo G=%d  |  temperatura=%.1f%n", MODELO, TAMANHO_GRUPO, TEMPERATURA);

        try {
            ollama.chamar(MODELO, "teste de conexao", 0.1);
        } catch (Exception e) {
            System.out.printf("Não foi possível conectar ao Ollama local: %s%n", e.getMessage());
            System.out.println("Rode 'ollama serve' e confirme que o modelo 'gemma4:e2b' está disponível ('ollama list').");
            return;
        }

        List<Double> recompensas = new ArrayList<>();
        for (int i = 0; i < TAMANHO_GRUPO; i++) {
            String bruto;
            try {
                bruto = ollama.chamar(MODELO, PROMPT_SCHEMA, TEMPERATURA);
            } catch (Exception e) {
                System.out.printf("  [%d] chamada ao Ollama falhou (%s) -- pulando esta amostra do grupo.%n", i, e.getMessage());
                continue;
            }
            Map<String, Object> candidato = JsonExtractor.extrairJson(bruto).orElse(null);
            double r = VerifiableReward.recompensaVerificavel(candidato);
            recompensas.add(r);
            System.out.printf(Locale.ROOT, "  [%d] recompensa=%.2f  json_valido=%s%n", i, r, candidato != null);
        }

        if (recompensas.isEmpty()) {
            System.out.println("Nenhuma amostra completou -- sem grupo pra calcular vantagem.");
            return;
        }

        double[] recompensasArr = recompensas.stream().mapToDouble(Double::doubleValue).toArray();
        GroupRelativeAdvantage.Resultado resultado = GroupRelativeAdvantage.calcular(recompensasArr);

        System.out.printf(Locale.ROOT, "%nGrupo completo: media(r)=%.3f  desvio_padrao(r)=%.3f%n",
                resultado.media(), resultado.desvio());

        if (resultado.desvio() == 0.0) {
            System.out.println("GRUPO DEGENERADO: todas as G respostas receberam a mesma recompensa -- sinal de aprendizado desaparece nessa rodada.");
        } else {
            for (int i = 0; i < recompensasArr.length; i++) {
                double a = resultado.vantagens()[i];
                String tag = a > 0 ? "reforçaria essa resposta (A>0)" : a < 0 ? "penalizaria essa resposta (A<0)" : "neutro (A=0)";
                System.out.printf(Locale.ROOT, "%2d  recompensa=%.2f  A_i=%.3f  %s%n", i, recompensasArr[i], a, tag);
            }
        }

        System.out.println("\nNenhum peso do modelo foi atualizado por este script -- fim do demo.");
    }
}
