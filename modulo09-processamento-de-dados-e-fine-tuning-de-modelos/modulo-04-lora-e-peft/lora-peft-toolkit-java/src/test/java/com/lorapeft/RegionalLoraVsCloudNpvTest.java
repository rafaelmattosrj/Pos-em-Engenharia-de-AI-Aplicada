package com.lorapeft;

import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

class RegionalLoraVsCloudNpvTest {

    @Test
    void volumeRegionalComCustoFixoDeGpuAluigadaDaNpvNegativoEm24Meses() {
        NpvCalculator.Resultado npv = NpvCalculator.calcularNPV(
                RegionalLoraVsCloudNpv.paramsBase(RegionalLoraVsCloudNpv.CUSTO_JOB_GERENCIADO));
        assertThat(npv.npv()).isLessThan(0);
    }

    @Test
    void volumeRegionalComCustoFixoDeGpuAluigadaNuncaAtingeBreakevenNoHorizonte() {
        NpvCalculator.Resultado npv = NpvCalculator.calcularNPV(
                RegionalLoraVsCloudNpv.paramsBase(RegionalLoraVsCloudNpv.CUSTO_JOB_GERENCIADO));
        assertThat(npv.mesBreakeven()).isNull();
    }

    @Test
    void mesmoVolumeRegionalComLoraLocalDaNpvPositivo() {
        NpvCalculator.Resultado npv = NpvCalculator.calcularNPV(
                RegionalLoraVsCloudNpv.paramsBase(RegionalLoraVsCloudNpv.CUSTO_LORA_LOCAL));
        assertThat(npv.npv()).isGreaterThan(0);
    }

    @Test
    void loraLocalAtingeBreakevenRapido() {
        NpvCalculator.Resultado npv = NpvCalculator.calcularNPV(
                RegionalLoraVsCloudNpv.paramsBase(RegionalLoraVsCloudNpv.CUSTO_LORA_LOCAL));
        assertThat(npv.mesBreakeven()).isNotNull();
        assertThat(npv.mesBreakeven()).isLessThanOrEqualTo(3);
    }

    @Test
    void diferencaDeNpvEntreOsDoisCaminhosEAproximadamenteOCustoFixo() {
        NpvCalculator.Resultado npvGerenciado = NpvCalculator.calcularNPV(
                RegionalLoraVsCloudNpv.paramsBase(RegionalLoraVsCloudNpv.CUSTO_JOB_GERENCIADO));
        NpvCalculator.Resultado npvLora = NpvCalculator.calcularNPV(
                RegionalLoraVsCloudNpv.paramsBase(RegionalLoraVsCloudNpv.CUSTO_LORA_LOCAL));
        double diferenca = npvLora.npv() - npvGerenciado.npv();
        assertThat(Math.abs(diferenca - RegionalLoraVsCloudNpv.CUSTO_JOB_GERENCIADO)).isLessThan(50);
    }
}
