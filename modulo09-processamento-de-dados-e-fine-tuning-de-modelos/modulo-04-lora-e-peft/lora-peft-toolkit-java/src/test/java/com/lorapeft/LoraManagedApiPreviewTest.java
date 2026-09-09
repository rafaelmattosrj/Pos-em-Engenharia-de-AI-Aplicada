package com.lorapeft;

import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

class LoraManagedApiPreviewTest {

    @Test
    void usaOEndpointRealDeFineTuningDaTogetherAi() {
        var req = LoraManagedApiPreview.montarRequisicaoLoraGerenciada(null, "file-abc123");
        assertThat(req.url()).isEqualTo("https://api.together.ai/v1/fine-tunes");
        assertThat(req.method()).isEqualTo("POST");
    }

    @Test
    void exigeTrainingFileId() {
        assertThatThrownBy(() -> LoraManagedApiPreview.montarRequisicaoLoraGerenciada(null, null))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("trainingFileId é obrigatório");
    }

    @Test
    void reaproveitaRank8Scale20Dropout0DoTreinoRealDeM42() {
        var req = LoraManagedApiPreview.montarRequisicaoLoraGerenciada(null, "file-abc123");
        var trainingType = req.body().getAsJsonObject("training_type");
        assertThat(trainingType.get("lora_r").getAsInt()).isEqualTo(8);
        assertThat(trainingType.get("lora_alpha").getAsDouble()).isEqualTo(20.0);
        assertThat(trainingType.get("lora_dropout").getAsDouble()).isEqualTo(0.0);
    }

    @Test
    void corpoDaRequisicaoUsaOFormatoDocumentado() {
        var req = LoraManagedApiPreview.montarRequisicaoLoraGerenciada(null, "file-abc123");
        var trainingType = req.body().getAsJsonObject("training_type");
        assertThat(trainingType.get("type").getAsString()).isEqualTo("Lora");
        assertThat(req.body().has("training_file")).isTrue();
        assertThat(req.body().has("model")).isTrue();
    }

    @Test
    void semApiKeyOHeaderAuthorizationUsaPlaceholder() {
        var req = LoraManagedApiPreview.montarRequisicaoLoraGerenciada(null, "file-abc123");
        assertThat(req.headers().get("Authorization")).isEqualTo("Bearer <TOGETHER_API_KEY>");
    }

    @Test
    void comApiKeyOHeaderAuthorizationCarregaAChaveDeVerdade() {
        var req = LoraManagedApiPreview.montarRequisicaoLoraGerenciada("sk-real-123", "file-abc123");
        assertThat(req.headers().get("Authorization")).isEqualTo("Bearer sk-real-123");
    }
}
