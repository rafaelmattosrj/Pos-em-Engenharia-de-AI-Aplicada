package com.opspilot.llm;

import com.opspilot.domain.Exceptions;
import com.opspilot.tools.Tool;

import java.util.List;

/**
 * Porte de OpsResilientChatModel / composeResilientRunnable (agents/model.ts):
 * retry no primario -> fallback pro backup (se configurado) -> lanca
 * ModelUnavailableException se ambos falharem apos as tentativas.
 */
public class ResilientChatModel implements ChatModel {

    public static final int RETRY_ATTEMPTS = 2;

    private final ChatModel primary;
    private final ChatModel backup;

    public ResilientChatModel(ChatModel primary, ChatModel backup) {
        this.primary = primary;
        this.backup = backup;
    }

    @Override
    public ModelResponse invoke(List<ChatMessage> messages, List<Tool> tools) {
        RuntimeException primaryError = null;
        for (int attempt = 0; attempt < RETRY_ATTEMPTS; attempt++) {
            try {
                return primary.invoke(messages, tools);
            } catch (RuntimeException e) {
                primaryError = e;
            }
        }

        if (backup == null) {
            throw new Exceptions.ModelUnavailableException(messageOf(primaryError));
        }

        RuntimeException backupError = null;
        for (int attempt = 0; attempt < RETRY_ATTEMPTS; attempt++) {
            try {
                return backup.invoke(messages, tools);
            } catch (RuntimeException e) {
                backupError = e;
            }
        }
        throw new Exceptions.ModelUnavailableException(messageOf(backupError));
    }

    private static String messageOf(RuntimeException error) {
        return error != null ? error.getMessage() : "All configured language models failed";
    }
}
