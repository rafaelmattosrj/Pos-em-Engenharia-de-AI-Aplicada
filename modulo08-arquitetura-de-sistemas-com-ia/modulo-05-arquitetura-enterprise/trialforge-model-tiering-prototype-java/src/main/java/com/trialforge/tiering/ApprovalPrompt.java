package com.trialforge.tiering;

import java.io.IOException;

/**
 * Approval Gate (Módulo 4.4): pede aprovação humana para um rascunho antes de
 * virar oficial. Extraída como interface para permitir testar
 * {@link CascadeGateway} sem esperar entrada real de stdin durante `mvn test`.
 */
public interface ApprovalPrompt {
    boolean approve(String rascunho) throws IOException;
}
