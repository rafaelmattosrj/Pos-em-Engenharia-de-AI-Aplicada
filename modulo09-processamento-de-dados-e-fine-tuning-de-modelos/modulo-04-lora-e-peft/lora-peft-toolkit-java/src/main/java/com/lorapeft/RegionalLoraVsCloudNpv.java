package com.lorapeft;

/**
 * Porte de regional-lora-vs-cloud-npv.js (Modulo 4.1). Reabre
 * {@link NpvCalculator#calcularNPV}, direto do decision-framework-tool
 * do Modulo 1.3 (aqui reimplementado, ver NpvCalculator), para responder:
 * quando o caso ja foi aprovado mas o volume e pequeno demais pro custo fixo
 * de GPU alugada, o que muda se o treino for local via LoRA?
 */
public final class RegionalLoraVsCloudNpv {

    private RegionalLoraVsCloudNpv() {
    }

    public static final double VOLUME_REGIONAL_MENSAL = 400;
    public static final double CRESCIMENTO_MENSAL = 0.03;
    public static final double CUSTO_STATUS_QUO = 0.045;
    public static final double CUSTO_FINE_TUNED = 0.016;
    // [Estimativa de mercado, ago/2026] Numero ilustrativo herdado do case Amplitude
    // Seguros do Modulo 1.3, calibrado pra sustentar a historia financeira do case,
    // usando como referencia frouxa a faixa de GPU cloud do cheatsheet do Modulo 1.3
    // (~US$0,40-4,00/hora). Confira precos atuais antes de usar isto pra uma decisao
    // real (vast.ai/pricing, runpod.io/pricing).
    public static final double CUSTO_JOB_GERENCIADO = 2400;
    public static final double CUSTO_LORA_LOCAL = 0;
    public static final int HORIZONTE_MESES = 24;
    public static final double TAXA_DESCONTO_MENSAL = 0.01;
    public static final int NUM_PARCERIAS_REGIONAIS = 5;

    public static NpvCalculator.Params paramsBase(double custoTreinamento) {
        return NpvCalculator.Params.of(
                VOLUME_REGIONAL_MENSAL, CRESCIMENTO_MENSAL, CUSTO_STATUS_QUO, CUSTO_FINE_TUNED,
                custoTreinamento, HORIZONTE_MESES, TAXA_DESCONTO_MENSAL);
    }
}
