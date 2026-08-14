package com.embeddings.util;

import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

class TextPreviewTest {

    @Test
    void textShorterThanMaxLengthIsReturnedUnchanged() {
        String result = TextPreview.truncate("hello world", 300);

        assertThat(result).isEqualTo("hello world");
    }

    @Test
    void textExactlyAtMaxLengthIsReturnedUnchanged() {
        String text = "a".repeat(300);

        String result = TextPreview.truncate(text, 300);

        assertThat(result).isEqualTo(text);
        assertThat(result).doesNotContain("...");
    }

    @Test
    void textLongerThanMaxLengthIsTruncatedWithEllipsis() {
        String text = "a".repeat(400);

        String result = TextPreview.truncate(text, 300);

        assertThat(result).hasSize(303); // 300 chars + "..."
        assertThat(result).endsWith("...");
        assertThat(result).startsWith("a".repeat(300));
    }

    @Test
    void emptyTextIsReturnedUnchanged() {
        assertThat(TextPreview.truncate("", 300)).isEmpty();
    }

    @Test
    void nullTextReturnsNull() {
        assertThat(TextPreview.truncate(null, 300)).isNull();
    }
}
