package spliter

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/textsplitter"
)

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp file failed: %v", err)
	}
	return path
}

func TestSplitTXT_Basic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	dir := t.TempDir()
	content := "Hello world\nThis is a simple test.\nLine3"
	file := writeTempFile(t, dir, "sample.txt", content)

	opts := Option{ChunkSize: 32, ChunkOverlap: 4, Separator: "\n"}
	chunks, err := Split(ctx, file, opts)
	require.NoError(t, err)
	require.NotEmpty(t, chunks)
	for _, c := range chunks {
		require.False(t, strings.Contains(c, "\n"), "chunk should not contain raw newline after cleaning: %q", c)
		require.NotEmpty(t, strings.TrimSpace(c), "chunk should not be empty after trim")
	}
}

func TestSplitMD_Headings(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	dir := t.TempDir()
	md := "# Title\n\n## Section\ncontent line 1\ncontent line 2\n\n### Sub\nmore content"
	file := writeTempFile(t, dir, "doc.md", md)

	// Use small chunk to encourage splitting by headings/separators
	opts := Option{ChunkSize: 40, ChunkOverlap: 0, Separator: "\n"}
	chunks, err := Split(ctx, file, opts)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(chunks), 2, "expected multiple chunks for markdown")
}

func TestSplitText_SeparatorChoiceAndChunking(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Case 1: spaces only, explicitly set separator to space
	content1 := "alpha beta gamma delta epsilon zeta eta theta iota"
	splitter1 := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(10),
		textsplitter.WithChunkOverlap(0),
		textsplitter.WithSeparators([]string{" ", ""}),
	)
	chunks, err := splitText(ctx, splitter1, content1, " ", 10)
	require.NoError(t, err)
	require.NotEmpty(t, chunks)
	for _, c := range chunks {
		require.False(t, strings.Contains(c, "\n"), "unexpected newline in chunk: %q", c)
		utfLen := len([]rune(c))
		require.Greater(t, utfLen, 0, "chunk should not be empty")
		require.LessOrEqual(t, utfLen, 10, "chunk size out of bound: %q (len=%d)", c, utfLen)
	}

	// Case 2: contains newlines, set separator to "\n"
	content2 := "row1\nrow2\nrow3 with long content to be split further"
	splitter2 := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(8),
		textsplitter.WithChunkOverlap(0),
		textsplitter.WithSeparators([]string{"\n", " ", ""}),
	)
	chunks, err = splitText(ctx, splitter2, content2, "\n", 8)
	require.NoError(t, err)
	require.NotEmpty(t, chunks)
	for _, c := range chunks {
		utfLen := len([]rune(c))
		require.Greater(t, utfLen, 0)
		require.LessOrEqual(t, utfLen, 8)
	}

	// Case 3: no explicit separator provided (defaults to "\n\n"), content has none
	// Should rely on splitter when chunkSize exceeded
	content3 := "ABCDEFGHIJKL" // 12 runes, chunkSize 5 => expect multiple chunks
	splitter3 := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(5),
		textsplitter.WithChunkOverlap(0),
		textsplitter.WithSeparators([]string{""}),
	)
	chunks, err = splitText(ctx, splitter3, content3, "", 5)
	require.NoError(t, err)
	require.NotEmpty(t, chunks)
	for _, c := range chunks {
		utfLen := len([]rune(c))
		require.Greater(t, utfLen, 0)
		require.LessOrEqual(t, utfLen, 5)
	}

	// Case 4: mixture with double newlines, choose "\n\n" as separator via default
	content4 := "  part1  \n\n  part2  \n  \npart3  "
	splitter4 := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(20),
		textsplitter.WithChunkOverlap(0),
		textsplitter.WithSeparators([]string{"\n\n", "\n", " ", ""}),
	)
	chunks, err = splitText(ctx, splitter4, content4, "", 20)
	require.NoError(t, err)
	require.NotEmpty(t, chunks)
	for _, c := range chunks {
		require.NotEmpty(t, strings.TrimSpace(c))
	}
}
