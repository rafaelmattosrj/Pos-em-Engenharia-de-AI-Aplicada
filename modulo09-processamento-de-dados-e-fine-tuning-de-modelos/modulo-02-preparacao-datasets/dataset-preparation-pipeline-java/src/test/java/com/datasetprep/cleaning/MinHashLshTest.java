package com.datasetprep.cleaning;

import org.junit.jupiter.api.Test;

import java.util.ArrayList;
import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

class MinHashLshTest {

    private Exemplo achar(List<Exemplo> dataset, String id) {
        return dataset.stream().filter(e -> e.metadata().id().equals(id)).findFirst().orElseThrow();
    }

    @Test
    void assinaturaMinHashDeTextoIdenticoProduzSimilaridadeEstimada1() {
        List<Exemplo> dataset = SampleDataset.gerarDatasetSimulado();
        var coeficientes = MinHashLsh.gerarCoeficientesHash(MinHashLsh.MINHASH_K, MinHashLsh.MINHASH_SEMENTE);
        Exemplo e = achar(dataset, "amplitude-auto-Oficina Estrela-0");
        long[] sig = MinHashLsh.assinaturaMinHash(MinHashLsh.shingles(e.entrada(), 5), coeficientes);
        assertThat(MinHashLsh.similaridadeMinHashEstimada(sig, sig)).isEqualTo(1.0);
    }

    @Test
    void minHashAproximaJaccardExatoParaDuplicataExata() {
        List<Exemplo> dataset = SampleDataset.gerarDatasetSimulado();
        var coeficientes = MinHashLsh.gerarCoeficientesHash(MinHashLsh.MINHASH_K, MinHashLsh.MINHASH_SEMENTE);
        Exemplo a = achar(dataset, "amplitude-auto-Oficina Estrela-0");
        Exemplo b = achar(dataset, "amplitude-auto-Oficina Estrela-0-reenviado");
        long[] sigA = MinHashLsh.assinaturaMinHash(MinHashLsh.shingles(a.entrada(), 5), coeficientes);
        long[] sigB = MinHashLsh.assinaturaMinHash(MinHashLsh.shingles(b.entrada(), 5), coeficientes);
        double estimado = MinHashLsh.similaridadeMinHashEstimada(sigA, sigB);
        double exato = MinHashLsh.similaridadeJaccardExata(a.entrada(), b.entrada());
        assertThat(Math.abs(estimado - exato)).isLessThan(0.1);
    }

    @Test
    void minHashAproximaJaccardExatoParaDuplicataComRuidoDeOcr() {
        List<Exemplo> dataset = SampleDataset.gerarDatasetSimulado();
        var coeficientes = MinHashLsh.gerarCoeficientesHash(MinHashLsh.MINHASH_K, MinHashLsh.MINHASH_SEMENTE);
        Exemplo a = achar(dataset, "amplitude-auto-Oficina Estrela-2");
        Exemplo b = achar(dataset, "amplitude-auto-Oficina Estrela-2-ruido-ocr");
        long[] sigA = MinHashLsh.assinaturaMinHash(MinHashLsh.shingles(a.entrada(), 5), coeficientes);
        long[] sigB = MinHashLsh.assinaturaMinHash(MinHashLsh.shingles(b.entrada(), 5), coeficientes);
        double estimado = MinHashLsh.similaridadeMinHashEstimada(sigA, sigB);
        double exato = MinHashLsh.similaridadeJaccardExata(a.entrada(), b.entrada());
        assertThat(Math.abs(estimado - exato)).isLessThan(0.15);
    }

    @Test
    void lshEncontraExatamenteOs3ParesDeQuaseDuplicataPlantados() {
        List<Exemplo> dataset = SampleDataset.gerarDatasetSimulado();
        var r = MinHashLsh.encontrarQuaseDuplicatasMinHashLSH(dataset);
        assertThat(r.paresDuplicata()).hasSize(3);
    }

    @Test
    void lshReduzOsNumeroDeComparacoesEmPeloMenos80Porcento() {
        List<Exemplo> dataset = SampleDataset.gerarDatasetSimulado();
        var r = MinHashLsh.encontrarQuaseDuplicatasMinHashLSH(dataset);
        double reducao = 1 - (double) r.totalCandidatosLSH() / r.totalParesForcaBruta();
        assertThat(reducao).isGreaterThanOrEqualTo(0.8);
    }

    @Test
    void nenhumParNaoDuplicataEConfirmadoComoDuplicata() {
        List<Exemplo> dataset = SampleDataset.gerarDatasetSimulado();
        var r = MinHashLsh.encontrarQuaseDuplicatasMinHashLSH(dataset);
        List<String> idsDuplicata = new ArrayList<>();
        for (var par : r.paresDuplicata()) {
            idsDuplicata.add(dataset.get(par.i()).metadata().id());
            idsDuplicata.add(dataset.get(par.j()).metadata().id());
        }
        assertThat(idsDuplicata).containsExactlyInAnyOrder(
                "amplitude-auto-Oficina Estrela-0", "amplitude-auto-Oficina Estrela-0-reenviado",
                "amplitude-auto-Oficina Estrela-2", "amplitude-auto-Oficina Estrela-2-ruido-ocr",
                "amplitude-saude-empresarial-Clinica Vitalis-0", "amplitude-saude-empresarial-Clinica Vitalis-0-reenviado");
    }

    @Test
    void remocaoDeQuaseDuplicataTiraExatamente3Exemplos() {
        List<Exemplo> dataset = SampleDataset.gerarDatasetSimulado();
        var r = MinHashLsh.removerQuaseDuplicatas(dataset);
        assertThat(r.removidos()).isEqualTo(3);
    }
}
