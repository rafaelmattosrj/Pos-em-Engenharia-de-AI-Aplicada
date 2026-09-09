package com.finetuningapi;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.function.BiFunction;

/**
 * Gerador do dataset em escala de treino real (305 brutos -> 300 dedup ->
 * 200 balanceados, 120 Amplitude Auto + 80 Amplitude Saúde Empresarial) --
 * porte de m3-dataset-scaling-tool.js (Modulo 3.2). Reusa o pipeline formal
 * do Modulo 2.2 via MinHashDedupBalancer (ver nota de adaptação naquela
 * classe: duplicado, não referenciado, por não haver módulo compartilhado
 * entre projetos Maven/Go independentes neste repo).
 */
public final class DatasetScaling {

    public record ExemploDataset(String instrucao, String entrada, Map<String, Object> saida, String caso, String fonte, String id) {
        /**
         * Base de shingling e' so `entrada` (nao instrucao+entrada) -- o mesmo
         * criterio de dataset-cleaning-balancing-tool.js (Modulo 2.2), que
         * compara so o documento de entrada dentro de um mesmo caso (a
         * instrucao e' identica pra todo exemplo do caso, entao concatenar so
         * infla a similaridade sem discriminar nada). A concatenacao
         * instrucao+entrada e' uma correcao especifica do extra Dolly
         * (DollyDatasetStarter), nao deste pipeline.
         */
        MinHashDedupBalancer.Exemplo paraDedup() {
            return new MinHashDedupBalancer.Exemplo(id, caso, fonte, entrada);
        }
    }

    private static final String[] NOMES = {
            "Marcos Vinicius Andrade Pereira", "Fernanda Costa Ribeiro", "Joaquim Pedro Salgado",
            "Beatriz Nogueira Lima", "Rafael Augusto Teixeira", "Camila dos Santos Farias",
            "Eduardo Henrique Barros", "Larissa Martins Cardoso", "Thiago Moreira Duarte",
            "Patricia Alves Monteiro", "Bruno Cesar Figueiredo", "Juliana Rocha Pimentel",
            "Gustavo Henrique Vasconcelos", "Renata Souza Albuquerque", "Diego Fernandes Castro",
            "Mariana Lopes Guimaraes", "Vinicius Almeida Correia", "Sabrina Ferreira Nunes",
            "Leonardo Batista Cavalcanti", "Priscila Andrade Melo", "Rodrigo Tavares Siqueira",
            "Amanda Cristina Peixoto", "Felipe Augusto Barbosa", "Carolina Machado Freitas",
            "Anderson Luiz Ramalho", "Vanessa Regina Coutinho", "Fabio Junior Aragao",
            "Debora Cristina Vieira", "Marcelo Souza Bittencourt", "Tatiane Pereira Godoy",
            "Alexandre Costa Miranda", "Cristiane Lopes Assuncao", "Fernando Braga Quintanilha",
            "Simone Rocha Vilaca", "Rogerio dos Santos Pena", "Michele Aparecida Fonseca",
            "Wagner Luiz Bessa", "Andreia Cristina Prado", "Cesar Augusto Nascimento",
            "Roberta Lima Sarmento", "Paulo Ricardo Andrade",
    };

    private static final String[] PLACAS = {
            "QJK-4F82", "RTL-9921", "MNB-3310", "PLW-7765", "ZXC-2298", "BVN-6641",
            "TYU-1183", "GHJ-5529", "FDS-8842", "LKM-3376", "OIU-9954", "CVB-1120",
            "ASD-6673", "WER-4481", "XSW-2290", "POI-7738", "HGF-3391", "MJU-6624",
            "NBV-1187", "KLO-5540", "ERT-8873", "YUI-2216", "CDE-9950", "VBN-4483",
            "AZS-6617", "QWE-1150", "DFG-7784", "RTY-3318", "FGH-8852", "TGB-2286",
            "YHN-5520", "UJM-9954", "IKM-4488", "OLP-1122", "WSX-6656", "EDC-1190",
            "RFV-5524", "TGB-9958", "YHN-3392", "UJM-7726", "ZAQ-1128", "XSW-6652", "CDE-3396",
    };

    private static final String[] PROCEDIMENTOS = {
            "consulta cardiologica", "exame de sangue completo", "fisioterapia ortopedica",
            "consulta ortopedica", "exame de imagem (ressonancia)", "consulta psiquiatrica",
            "sessao de fonoaudiologia", "exame oftalmologico", "consulta dermatologica",
            "exame de densitometria ossea", "consulta nutricional", "sessao de acupuntura",
            "consulta ginecologica", "exame de urina completo", "sessao de terapia ocupacional",
            "consulta endocrinologica", "exame de eletrocardiograma", "consulta neurologica",
            "sessao de pilates terapeutico", "exame de audiometria", "consulta pediatrica",
            "exame de mamografia", "sessao de psicoterapia", "consulta geriatrica",
            "consulta de clinica geral", "exame de colonoscopia", "sessao de fonoterapia",
            "consulta urologica", "exame de tomografia",
    };

    private static final String[] VALORES = {
            "3.210,50", "1.870,00", "5.640,00", "2.430,75", "890,00", "4.120,30",
            "1.250,00", "3.980,60", "2.760,00", "6.310,90", "1.540,00", "2.990,25",
            "3.450,00", "1.780,50", "4.560,00", "2.220,80", "5.120,00", "1.630,40",
            "3.870,00", "2.045,90", "4.780,60", "1.395,00", "6.020,50", "2.510,30",
            "3.660,00", "1.925,80", "4.310,00", "2.870,60", "5.480,00", "1.485,70",
            "3.120,00", "2.640,90", "4.950,00", "1.780,00", "3.390,60", "2.210,00",
            "5.870,00", "1.660,40", "4.120,00", "2.980,50", "3.780,90", "2.340,00", "4.910,60",
            "1.590,00", "3.260,40", "2.150,80", "4.430,00",
    };

    private static final String[] DATAS = {
            "12/03/2026", "02/04/2026", "18/05/2026", "25/03/2026", "09/04/2026", "30/04/2026",
            "14/03/2026", "21/05/2026", "05/04/2026", "11/05/2026", "28/03/2026", "16/04/2026",
            "03/06/2026", "19/06/2026", "07/06/2026", "24/06/2026", "01/07/2026", "15/07/2026",
            "22/07/2026", "29/07/2026", "06/02/2026", "13/02/2026", "20/02/2026", "27/02/2026",
            "04/02/2026", "10/06/2026", "17/03/2026", "26/04/2026", "02/05/2026", "08/06/2026",
            "23/06/2026",
    };

    private static final Map<String, BiFunction4> TEMPLATES_AUTO = new LinkedHashMap<>();
    private static final Map<String, BiFunction4> TEMPLATES_SAUDE = new LinkedHashMap<>();

    @FunctionalInterface
    private interface BiFunction4 {
        String aplicar(String nome, String campo2, String data, String valor);
    }

    static {
        TEMPLATES_AUTO.put("Oficina Estrela", (nome, placa, data, valor) ->
                "ATIVA ORCAMENTOS AUTOMOTIVOS OFICINA ESTRELA LTDA CNPJ 12.345.678/0001-90 Rua das Turbinas 450 Distrito Industrial "
                        + "Segurado: " + nome + " Placa do veiculo: " + placa + " Data do sinistro: " + data + " "
                        + "Descricao do servico: reparo de lataria e pintura no para-choque dianteiro Valor total do reparo: R$ " + valor);
        TEMPLATES_AUTO.put("Auto Center Silva", (nome, placa, data, valor) ->
                "AUTO CENTER SILVA - FUNILARIA E PINTURA - CNPJ 98.765.432/0001-11 Av. dos Mecanicos 220 "
                        + "Cliente/Segurado: " + nome + " Placa: " + placa + " Data do atendimento: " + data + " "
                        + "Servico executado: troca de para-lama e revisao de suspensao dianteira Valor: R$ " + valor);
        TEMPLATES_AUTO.put("Funilaria Rio Bonito", (nome, placa, data, valor) ->
                "FUNILARIA RIO BONITO ME CNPJ 45.111.222/0001-33 Rua Rio Bonito 88 "
                        + "Nome do segurado: " + nome + " Placa do veiculo: " + placa + " Data: " + data + " "
                        + "Orcamento: substituicao de parachoque traseiro e polimento Valor total: R$ " + valor);
        TEMPLATES_AUTO.put("Oficina Nova Aliança", (nome, placa, data, valor) ->
                "OFICINA NOVA ALIANCA LTDA CNPJ 22.333.444/0001-55 Estrada Velha 1200 "
                        + "Segurado: " + nome + " Placa do carro: " + placa + " Data do orcamento: " + data + " "
                        + "Descricao: reparo de amassado na porta dianteira Valor cobrado: R$ " + valor);
        TEMPLATES_AUTO.put("Mecânica Horizonte", (nome, placa, data, valor) ->
                "MECANICA HORIZONTE LTDA CNPJ 51.222.888/0001-19 Av. do Horizonte 640 "
                        + "Segurado: " + nome + " Placa do veiculo: " + placa + " Data do servico: " + data + " "
                        + "Descricao: alinhamento e balanceamento apos colisao lateral Valor total: R$ " + valor);
        TEMPLATES_AUTO.put("Auto Reparos União", (nome, placa, data, valor) ->
                "AUTO REPAROS UNIAO ME CNPJ 63.444.777/0001-28 Rua da Uniao 305 "
                        + "Nome do segurado: " + nome + " Placa: " + placa + " Data do atendimento: " + data + " "
                        + "Servico: troca de para-brisa trincado Valor cobrado: R$ " + valor);

        TEMPLATES_SAUDE.put("Clínica Vitalis", (nome, procedimento, data, valor) ->
                "CLINICA VITALIS SAUDE OCUPACIONAL CNPJ 33.222.111/0001-44 Av. Paulista 900 "
                        + "Paciente/Beneficiario: " + nome + " Procedimento: " + procedimento + " Data do atendimento: " + data + " "
                        + "Valor cobrado: R$ " + valor);
        TEMPLATES_SAUDE.put("Hospital Santa Clara", (nome, procedimento, data, valor) ->
                "HOSPITAL SANTA CLARA CNPJ 66.555.444/0001-22 Rua das Acacias 310 "
                        + "Beneficiario: " + nome + " Procedimento realizado: " + procedimento + " Data: " + data + " "
                        + "Valor total: R$ " + valor);
        TEMPLATES_SAUDE.put("Centro Médico Bem Estar", (nome, procedimento, data, valor) ->
                "CENTRO MEDICO BEM ESTAR CNPJ 77.888.999/0001-66 Rua da Saude 45 "
                        + "Nome do beneficiario: " + nome + " Procedimento: " + procedimento + " Data da consulta: " + data + " "
                        + "Valor cobrado: R$ " + valor);
        TEMPLATES_SAUDE.put("Clínica São Rafael", (nome, procedimento, data, valor) ->
                "CLINICA SAO RAFAEL CNPJ 84.111.222/0001-37 Rua Sao Rafael 512 "
                        + "Paciente/Beneficiario: " + nome + " Procedimento: " + procedimento + " Data do atendimento: " + data + " "
                        + "Valor total: R$ " + valor);
        TEMPLATES_SAUDE.put("Instituto Saúde Plena", (nome, procedimento, data, valor) ->
                "INSTITUTO SAUDE PLENA LTDA CNPJ 91.333.555/0001-08 Av. da Saude Plena 78 "
                        + "Beneficiario: " + nome + " Procedimento realizado: " + procedimento + " Data: " + data + " "
                        + "Valor cobrado: R$ " + valor);
    }

    private static final String[] PRENOMES = new String[NOMES.length];
    private static final String[] SOBRENOMES = new String[37];

    static {
        for (int i = 0; i < NOMES.length; i++) PRENOMES[i] = NOMES[i].split(" ")[0];
        for (int i = 0; i < 37; i++) {
            String[] partes = NOMES[i].split(" ");
            SOBRENOMES[i] = String.join(" ", java.util.Arrays.copyOfRange(partes, 1, partes.length));
        }
    }

    private static final Object[][] FONTES_AUTO = {
            {"Oficina Estrela", 60}, {"Auto Center Silva", 40}, {"Funilaria Rio Bonito", 30},
            {"Oficina Nova Aliança", 25}, {"Mecânica Horizonte", 15}, {"Auto Reparos União", 10},
    };

    private static final Object[][] FONTES_SAUDE = {
            {"Clínica Vitalis", 50}, {"Hospital Santa Clara", 30}, {"Centro Médico Bem Estar", 20},
            {"Clínica São Rafael", 12}, {"Instituto Saúde Plena", 8},
    };

    private DatasetScaling() {
    }

    private static double parseValorBr(String valor) {
        return Double.parseDouble(valor.replace(".", "").replace(",", "."));
    }

    public static ExemploDataset gerarExemplo(String caso, String fonte, int indice) {
        String nome = PRENOMES[indice % PRENOMES.length] + " " + SOBRENOMES[(indice * 7 + 3) % SOBRENOMES.length];
        String data = DATAS[(indice * 7 + 3) % DATAS.length];
        String valor = VALORES[(indice * 11 + 5) % VALORES.length];

        String entrada;
        Map<String, Object> saida;
        String instrucao;
        if (caso.equals("amplitude-auto")) {
            String placa = PLACAS[(indice * 13 + 2) % PLACAS.length];
            entrada = TEMPLATES_AUTO.get(fonte).aplicar(nome, placa, data, valor);
            saida = new LinkedHashMap<>();
            saida.put("segurado", nome);
            saida.put("placa", placa);
            saida.put("valor", parseValorBr(valor));
            instrucao = "Extraia segurado, placa e valor do orçamento de oficina abaixo.";
        } else {
            String procedimento = PROCEDIMENTOS[(indice * 5 + 1) % PROCEDIMENTOS.length];
            entrada = TEMPLATES_SAUDE.get(fonte).aplicar(nome, procedimento, data, valor);
            saida = new LinkedHashMap<>();
            saida.put("beneficiario", nome);
            saida.put("procedimento", procedimento);
            saida.put("valor", parseValorBr(valor));
            instrucao = "Extraia beneficiário, procedimento e valor do recibo médico abaixo.";
        }

        return new ExemploDataset(instrucao, entrada, saida, caso, fonte, caso + "-" + fonte + "-" + indice);
    }

    public static List<ExemploDataset> gerarDatasetBruto() {
        List<ExemploDataset> exemplos = new ArrayList<>();
        for (Object[] fn : FONTES_AUTO) {
            String fonte = (String) fn[0];
            int n = (int) fn[1];
            for (int i = 0; i < n; i++) exemplos.add(gerarExemplo("amplitude-auto", fonte, i));
        }
        for (Object[] fn : FONTES_SAUDE) {
            String fonte = (String) fn[0];
            int n = (int) fn[1];
            for (int i = 0; i < n; i++) exemplos.add(gerarExemplo("amplitude-saude-empresarial", fonte, i));
        }

        exemplos.add(reidentificar(buscarPorId(exemplos, "amplitude-auto-Oficina Estrela-0"),
                "amplitude-auto", "Oficina Estrela", "amplitude-auto-Oficina Estrela-0-reenviado"));
        exemplos.add(reidentificar(buscarPorId(exemplos, "amplitude-auto-Auto Center Silva-0"),
                "amplitude-auto", "Auto Center Silva", "amplitude-auto-Auto Center Silva-0-reenviado"));

        ExemploDataset base1 = buscarPorId(exemplos, "amplitude-auto-Oficina Estrela-5");
        exemplos.add(new ExemploDataset(base1.instrucao(),
                base1.entrada().replace("Placa do veiculo:", "P1aca do veicu1o:"), base1.saida(),
                "amplitude-auto", "Oficina Estrela", "amplitude-auto-Oficina Estrela-5-ruido-ocr"));

        exemplos.add(reidentificar(buscarPorId(exemplos, "amplitude-saude-empresarial-Clínica Vitalis-0"),
                "amplitude-saude-empresarial", "Clínica Vitalis", "amplitude-saude-empresarial-Clínica Vitalis-0-reenviado"));

        ExemploDataset base2 = buscarPorId(exemplos, "amplitude-saude-empresarial-Clínica Vitalis-10");
        exemplos.add(new ExemploDataset(base2.instrucao(),
                base2.entrada().replace("Valor cobrado:", "Va1or cobrad0:"), base2.saida(),
                "amplitude-saude-empresarial", "Clínica Vitalis", "amplitude-saude-empresarial-Clínica Vitalis-10-ruido-ocr"));

        return exemplos;
    }

    private static ExemploDataset buscarPorId(List<ExemploDataset> exemplos, String id) {
        return exemplos.stream().filter(e -> e.id().equals(id)).findFirst()
                .orElseThrow(() -> new IllegalStateException("id nao encontrado: " + id));
    }

    private static ExemploDataset reidentificar(ExemploDataset original, String caso, String fonte, String novoId) {
        return new ExemploDataset(original.instrucao(), original.entrada(), original.saida(), caso, fonte, novoId);
    }

    public record ResultadoPipeline(int original, int aposDedup, int duplicatasRemovidas, int total,
                                     List<ExemploDataset> exemplosFinal, Map<String, RelatorioCaso> relatorioPorCaso) {
    }

    public record RelatorioCaso(Map<String, Integer> contagensAntes, double entropiaAntes, double nEfetivoAntes,
                                 Map<String, Integer> contagensDepois, double entropiaDepois, double nEfetivoDepois) {
    }

    /** Equivalente a limparEBalancear do JS: dedup por caso (mantendo a primeira ocorrência), depois balanceia por temperatura. */
    public static ResultadoPipeline limparEBalancear(List<ExemploDataset> exemplos, Map<String, Integer> alvos) {
        List<MinHashDedupBalancer.Exemplo> paraDedup = exemplos.stream().map(ExemploDataset::paraDedup).toList();

        java.util.Set<Integer> remover = new java.util.HashSet<>();
        for (String caso : List.of("amplitude-auto", "amplitude-saude-empresarial")) {
            List<Integer> indicesDoCaso = new ArrayList<>();
            for (int i = 0; i < exemplos.size(); i++) if (exemplos.get(i).caso().equals(caso)) indicesDoCaso.add(i);
            List<MinHashDedupBalancer.Exemplo> doCaso = indicesDoCaso.stream().map(paraDedup::get).toList();
            var dedup = MinHashDedupBalancer.encontrarQuaseDuplicatasGenerico(doCaso, 5);
            for (var par : dedup.paresDuplicata()) remover.add(indicesDoCaso.get(par.j()));
        }

        List<ExemploDataset> aposDedup = new ArrayList<>();
        for (int i = 0; i < exemplos.size(); i++) if (!remover.contains(i)) aposDedup.add(exemplos.get(i));

        Map<String, RelatorioCaso> relatorio = new LinkedHashMap<>();
        List<ExemploDataset> atual = aposDedup;

        for (String caso : List.of("amplitude-auto", "amplitude-saude-empresarial")) {
            Map<String, Integer> contagensAntes = contarPorFonte(atual, caso);
            Map<String, Double> distAntes = MinHashDedupBalancer.distribuicaoDe(contagensAntes);
            int alvo = alvos.get(caso);
            Map<String, Integer> alocacao = MinHashDedupBalancer.alocarComCapacidade(contagensAntes, MinHashDedupBalancer.ALPHA_TEMPERATURA, alvo);

            Map<String, Integer> usados = new LinkedHashMap<>();
            List<ExemploDataset> selecionados = new ArrayList<>();
            List<ExemploDataset> outros = new ArrayList<>();
            for (ExemploDataset e : atual) {
                if (!e.caso().equals(caso)) {
                    outros.add(e);
                    continue;
                }
                int usadoAtual = usados.getOrDefault(e.fonte(), 0);
                if (usadoAtual < alocacao.getOrDefault(e.fonte(), 0)) {
                    selecionados.add(e);
                    usados.put(e.fonte(), usadoAtual + 1);
                }
            }
            List<ExemploDataset> novo = new ArrayList<>(outros);
            novo.addAll(selecionados);
            atual = novo;

            Map<String, Integer> contagensDepois = contarPorFonte(atual, caso);
            Map<String, Double> distDepois = MinHashDedupBalancer.distribuicaoDe(contagensDepois);
            relatorio.put(caso, new RelatorioCaso(
                    contagensAntes, MinHashDedupBalancer.entropiaShannon(distAntes), MinHashDedupBalancer.numeroEfetivoFontes(distAntes),
                    contagensDepois, MinHashDedupBalancer.entropiaShannon(distDepois), MinHashDedupBalancer.numeroEfetivoFontes(distDepois)));
        }

        return new ResultadoPipeline(exemplos.size(), aposDedup.size(), remover.size(), atual.size(), atual, relatorio);
    }

    private static Map<String, Integer> contarPorFonte(List<ExemploDataset> exemplos, String caso) {
        Map<String, Integer> contagem = new LinkedHashMap<>();
        for (ExemploDataset e : exemplos) if (e.caso().equals(caso)) contagem.merge(e.fonte(), 1, Integer::sum);
        return contagem;
    }
}
