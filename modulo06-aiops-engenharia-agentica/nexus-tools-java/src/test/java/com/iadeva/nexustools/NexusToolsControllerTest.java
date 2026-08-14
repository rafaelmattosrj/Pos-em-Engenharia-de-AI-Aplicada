package com.iadeva.nexustools;

import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;

import java.io.File;

import static org.hamcrest.Matchers.containsString;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

/**
 * Testes de integração HTTP — smoke test confirmando que os endpoints estão
 * corretamente conectados aos services.
 */
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
@AutoConfigureMockMvc
class NexusToolsControllerTest {

    @Autowired
    private MockMvc mockMvc;

    @Test
    void postCheckComplianceRules_retornaRegrasDePolitica() throws Exception {
        mockMvc.perform(post("/tools/check-compliance-rules")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"query\": \"nomenclatura de buckets\"}"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.result", containsString("nexus-")));
    }

    @Test
    void postSuggestFix_retornaRemediacaoParaOOMKilled() throws Exception {
        mockMvc.perform(post("/tools/suggest-fix")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"issueType\": \"OOMKilled\"}"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.result", containsString("resources.limits.memory")));
    }

    @Test
    void postExecuteTerraform_bloqueiaAcaoDestrutivaSemSenha() throws Exception {
        mockMvc.perform(post("/tools/execute-terraform")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"command\": \"destroy producao\"}"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.result", containsString("BLOCKED")));

        new File("test-output.tf").delete();
    }

    @Test
    void postWriteFile_salvaConteudoNoDisco() throws Exception {
        mockMvc.perform(post("/tools/write-file")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"content\": \"resource \\\"aws_s3_bucket\\\" \\\"x\\\" {}\", \"filename\": \"controller-test-output.tf\"}"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.result", containsString("saved successfully")));

        new File("controller-test-output.tf").delete();
    }
}
