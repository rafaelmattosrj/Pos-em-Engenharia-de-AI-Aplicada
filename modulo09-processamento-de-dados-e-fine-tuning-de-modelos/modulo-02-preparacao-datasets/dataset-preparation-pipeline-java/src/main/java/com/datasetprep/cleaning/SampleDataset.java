package com.datasetprep.cleaning;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.function.BiFunction;

/**
 * Porte da secao 1 de dataset-cleaning-balancing-tool.js: gera o mesmo
 * dataset simulado (mesma escala/casos), com duas quase-duplicatas
 * plantadas por caso (reenvio identico + ruido de OCR), para exercitar o
 * pipeline de deduplicacao/balanceamento/diversidade.
 */
public final class SampleDataset {

    private SampleDataset() {
    }

    private interface TemplateAuto {
        String gerar(String nome, String placa, String data, String valor);
    }

    private interface TemplateSaude {
        String gerar(String nome, String procedimento, String data, String valor);
    }

    private static final Map<String, TemplateAuto> TEMPLATES_AUTO = new HashMap<>();
    private static final Map<String, TemplateSaude> TEMPLATES_SAUDE = new HashMap<>();

    static {
        TEMPLATES_AUTO.put("Oficina Estrela", (nome, placa, data, valor) ->
                "ATIVA ORCAMENTOS AUTOMOTIVOS OFICINA ESTRELA LTDA CNPJ 12.345.678/0001-90 Rua das Turbinas 450 Distrito Industrial "
                        + "Segurado: " + nome + " Placa do veiculo: " + placa + " Data do sinistro: " + data
                        + " Descricao do servico: reparo de lataria e pintura no para-choque dianteiro Valor total do reparo: R$ " + valor);
        TEMPLATES_AUTO.put("Auto Center Silva", (nome, placa, data, valor) ->
                "AUTO CENTER SILVA - FUNILARIA E PINTURA - CNPJ 98.765.432/0001-11 Av. dos Mecanicos 220 "
                        + "Cliente/Segurado: " + nome + " Placa: " + placa + " Data do atendimento: " + data
                        + " Servico executado: troca de para-lama e revisao de suspensao dianteira Valor: R$ " + valor);
        TEMPLATES_AUTO.put("Funilaria Rio Bonito", (nome, placa, data, valor) ->
                "FUNILARIA RIO BONITO ME CNPJ 45.111.222/0001-33 Rua Rio Bonito 88 "
                        + "Nome do segurado: " + nome + " Placa do veiculo: " + placa + " Data: " + data
                        + " Orcamento: substituicao de parachoque traseiro e polimento Valor total: R$ " + valor);
        TEMPLATES_AUTO.put("Oficina Nova Alianca", (nome, placa, data, valor) ->
                "OFICINA NOVA ALIANCA LTDA CNPJ 22.333.444/0001-55 Estrada Velha 1200 "
                        + "Segurado: " + nome + " Placa do carro: " + placa + " Data do orcamento: " + data
                        + " Descricao: reparo de amassado na porta dianteira Valor cobrado: R$ " + valor);

        TEMPLATES_SAUDE.put("Clinica Vitalis", (nome, procedimento, data, valor) ->
                "CLINICA VITALIS SAUDE OCUPACIONAL CNPJ 33.222.111/0001-44 Av. Paulista 900 "
                        + "Paciente/Beneficiario: " + nome + " Procedimento: " + procedimento + " Data do atendimento: " + data
                        + " Valor cobrado: R$ " + valor);
        TEMPLATES_SAUDE.put("Hospital Santa Clara", (nome, procedimento, data, valor) ->
                "HOSPITAL SANTA CLARA CNPJ 66.555.444/0001-22 Rua das Acacias 310 "
                        + "Beneficiario: " + nome + " Procedimento realizado: " + procedimento + " Data: " + data
                        + " Valor total: R$ " + valor);
        TEMPLATES_SAUDE.put("Centro Medico Bem Estar", (nome, procedimento, data, valor) ->
                "CENTRO MEDICO BEM ESTAR CNPJ 77.888.999/0001-66 Rua da Saude 45 "
                        + "Nome do beneficiario: " + nome + " Procedimento: " + procedimento + " Data da consulta: " + data
                        + " Valor cobrado: R$ " + valor);
    }

    private static final String[] NOMES = {
            "Marcos Vinicius Andrade Pereira", "Fernanda Costa Ribeiro", "Joaquim Pedro Salgado",
            "Beatriz Nogueira Lima", "Rafael Augusto Teixeira", "Camila dos Santos Farias",
            "Eduardo Henrique Barros", "Larissa Martins Cardoso", "Thiago Moreira Duarte",
            "Patricia Alves Monteiro", "Bruno Cesar Figueiredo", "Juliana Rocha Pimentel",
            "Gustavo Henrique Vasconcelos", "Renata Souza Albuquerque", "Diego Fernandes Castro",
            "Mariana Lopes Guimaraes",
    };
    private static final String[] PLACAS = {
            "QJK-4F82", "RTL-9921", "MNB-3310", "PLW-7765", "ZXC-2298", "BVN-6641",
            "TYU-1183", "GHJ-5529", "FDS-8842", "LKM-3376", "OIU-9954", "CVB-1120",
            "ASD-6673", "WER-4481", "XSW-2290", "POI-7738",
    };
    private static final String[] PROCEDIMENTOS = {
            "consulta cardiologica", "exame de sangue completo", "fisioterapia ortopedica",
            "consulta ortopedica", "exame de imagem (ressonancia)", "consulta psiquiatrica",
            "sessao de fonoaudiologia", "exame oftalmologico", "consulta dermatologica",
            "exame de densitometria ossea", "consulta nutricional", "sessao de acupuntura",
    };
    private static final String[] VALORES = {
            "3.210,50", "1.870,00", "5.640,00", "2.430,75", "890,00", "4.120,30",
            "1.250,00", "3.980,60", "2.760,00", "6.310,90", "1.540,00", "2.990,25",
            "3.450,00", "1.780,50", "4.560,00", "2.220,80",
    };
    private static final String[] DATAS = {
            "12/03/2026", "02/04/2026", "18/05/2026", "25/03/2026", "09/04/2026", "30/04/2026",
            "14/03/2026", "21/05/2026", "05/04/2026", "11/05/2026", "28/03/2026", "16/04/2026",
    };

    private static double parseValorBr(String valor) {
        return Double.parseDouble(valor.replace(".", "").replace(",", "."));
    }

    public static Exemplo gerarExemplo(String caso, String fonte, int indice) {
        String nome = NOMES[indice % NOMES.length];
        String data = DATAS[indice % DATAS.length];
        String valor = VALORES[indice % VALORES.length];

        String entrada;
        Map<String, Object> saida = new HashMap<>();
        String instrucao;
        if (caso.equals("amplitude-auto")) {
            String placa = PLACAS[indice % PLACAS.length];
            entrada = TEMPLATES_AUTO.get(fonte).gerar(nome, placa, data, valor);
            saida.put("segurado", nome);
            saida.put("placa", placa);
            saida.put("valor", parseValorBr(valor));
            instrucao = "Extraia segurado, placa e valor do orcamento de oficina abaixo.";
        } else {
            String procedimento = PROCEDIMENTOS[indice % PROCEDIMENTOS.length];
            entrada = TEMPLATES_SAUDE.get(fonte).gerar(nome, procedimento, data, valor);
            saida.put("beneficiario", nome);
            saida.put("procedimento", procedimento);
            saida.put("valor", parseValorBr(valor));
            instrucao = "Extraia beneficiario, procedimento e valor do recibo medico abaixo.";
        }

        return new Exemplo(instrucao, entrada, saida, new Exemplo.Metadata(caso, fonte, caso + "-" + fonte + "-" + indice));
    }

    /** Gera o dataset simulado com 2 quase-duplicatas plantadas por caso (reenvio + ruido de OCR). */
    public static List<Exemplo> gerarDatasetSimulado() {
        List<Exemplo> exemplos = new ArrayList<>();

        for (int i = 0; i < 16; i++) exemplos.add(gerarExemplo("amplitude-auto", "Oficina Estrela", i));
        for (int i = 0; i < 5; i++) exemplos.add(gerarExemplo("amplitude-auto", "Auto Center Silva", i + 20));
        for (int i = 0; i < 4; i++) exemplos.add(gerarExemplo("amplitude-auto", "Funilaria Rio Bonito", i + 30));
        for (int i = 0; i < 3; i++) exemplos.add(gerarExemplo("amplitude-auto", "Oficina Nova Alianca", i + 40));

        exemplos.set(1, gerarExemplo("amplitude-auto", "Oficina Estrela", 0)
                .comMetadata(new Exemplo.Metadata("amplitude-auto", "Oficina Estrela", "amplitude-auto-Oficina Estrela-0-reenviado")));

        Exemplo base = gerarExemplo("amplitude-auto", "Oficina Estrela", 2);
        Exemplo ocrRuido = new Exemplo(
                base.instrucao(),
                base.entrada().replace("Placa do veiculo:", "P1aca do veicu1o:"),
                base.saida(),
                new Exemplo.Metadata("amplitude-auto", "Oficina Estrela", "amplitude-auto-Oficina Estrela-2-ruido-ocr")
        );
        exemplos.set(3, ocrRuido);

        for (int i = 0; i < 12; i++) exemplos.add(gerarExemplo("amplitude-saude-empresarial", "Clinica Vitalis", i));
        for (int i = 0; i < 4; i++) exemplos.add(gerarExemplo("amplitude-saude-empresarial", "Hospital Santa Clara", i + 20));
        for (int i = 0; i < 3; i++) exemplos.add(gerarExemplo("amplitude-saude-empresarial", "Centro Medico Bem Estar", i + 30));

        int idxVitalis0 = -1;
        for (int i = 0; i < exemplos.size(); i++) {
            Exemplo e = exemplos.get(i);
            if (e.metadata().fonte().equals("Clinica Vitalis") && e.metadata().id().endsWith("-0")) {
                idxVitalis0 = i;
                break;
            }
        }
        exemplos.set(idxVitalis0 + 1, gerarExemplo("amplitude-saude-empresarial", "Clinica Vitalis", 0)
                .comMetadata(new Exemplo.Metadata("amplitude-saude-empresarial", "Clinica Vitalis", "amplitude-saude-empresarial-Clinica Vitalis-0-reenviado")));

        return exemplos;
    }
}
