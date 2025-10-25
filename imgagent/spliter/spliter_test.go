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

// TestSplitBooks_ChapterDetection 测试小说章节检测功能（使用 mock 文件）
func TestSplitBooks_ChapterDetection(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Mock 小说内容测试用例
	testCases := []struct {
		name     string
		content  string
		expected int // 期望的章节数量
	}{
		{
			name: "金庸武侠小说格式",
			content: `第一章 华山论剑
华山之巅，剑气纵横。令狐冲手持长剑，面对强敌。

第二章 独孤九剑
独孤九剑，破尽天下武功。令狐冲领悟剑意。

第三章 笑傲江湖
江湖路远，笑傲人生。令狐冲终成一代大侠。`,
			expected: 3,
		},
		{
			name: "古典小说回目格式",
			content: `第一回 桃园三结义
话说天下大势，分久必合，合久必分。

第二回 张翼德怒鞭督邮
张飞怒鞭督邮，刘备三兄弟投奔公孙瓒。

第三回 议温明董卓叱丁原
董卓进京，废立皇帝，专权朝政。`,
			expected: 3,
		},
		{
			name: "现代小说章节格式",
			content: `第1章 开始
这是一个现代故事的开头。

第2章 发展
故事逐渐展开，人物关系复杂化。

第3章 高潮
冲突达到顶点，情节紧张刺激。

第4章 结局
故事走向尾声，人物命运尘埃落定。`,
			expected: 4,
		},
		{
			name: "混合格式",
			content: `第一章 序章
这是序章内容。

第二回 正传开始
正传内容开始。

第3节 细节描述
详细的描述内容。

第四章 总结
总结章节内容。`,
			expected: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建临时文件
			dir := t.TempDir()
			filePath := writeTempFile(t, dir, "test.txt", tc.content)

			// 测试章节分割
			opts := Option{
				ChunkSize:    1000, // 较大的块大小，优先按章节分割
				ChunkOverlap: 100,
				Separator:    "\n\n",
			}

			chunks, err := Split(ctx, filePath, opts)
			require.NoError(t, err, "分割文件失败: %s", tc.name)
			require.NotEmpty(t, chunks, "分割结果为空: %s", tc.name)

			t.Logf("文件 %s 分割结果: %d 个块", tc.name, len(chunks))

			// 检查每个块的内容
			for i, chunk := range chunks {
				require.NotEmpty(t, strings.TrimSpace(chunk), "第 %d 个块为空", i+1)
				require.False(t, strings.Contains(chunk, "\n"), "块 %d 包含未处理的换行符", i+1)

				// 打印前几个块的内容预览
				if i < 3 {
					preview := chunk
					if len(preview) > 100 {
						preview = preview[:100] + "..."
					}
					t.Logf("块 %d 预览: %s", i+1, preview)
				}
			}

			// 验证章节数量
			require.Equal(t, tc.expected, len(chunks), "章节数量不匹配")
		})
	}
}

// TestSplitBooks_SpecificNovels 测试特定小说类型的章节识别（使用 mock 文件）
func TestSplitBooks_SpecificNovels(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testCases := []struct {
		name        string
		description string
		content     string
		expected    int
	}{
		{
			name:        "金庸武侠小说",
			description: "金庸武侠小说，应该有章节结构",
			content: `第一章 华山论剑
华山之巅，剑气纵横。令狐冲手持长剑，面对强敌。

第二章 独孤九剑
独孤九剑，破尽天下武功。令狐冲领悟剑意。

第三章 笑傲江湖
江湖路远，笑傲人生。令狐冲终成一代大侠。`,
			expected: 3,
		},
		{
			name:        "古典小说",
			description: "古典小说，应该有回目结构",
			content: `第一回 桃园三结义
话说天下大势，分久必合，合久必分。

第二回 张翼德怒鞭督邮
张飞怒鞭督邮，刘备三兄弟投奔公孙瓒。

第三回 议温明董卓叱丁原
董卓进京，废立皇帝，专权朝政。`,
			expected: 3,
		},
		{
			name:        "现代小说",
			description: "现代小说，可能有章节结构",
			content: `第1章 开始
这是一个现代故事的开头。

第2章 发展
故事逐渐展开，人物关系复杂化。

第3章 高潮
冲突达到顶点，情节紧张刺激。`,
			expected: 3,
		},
		{
			name:        "历史小说",
			description: "历史小说，应该有章节结构",
			content: `第一章 明朝建立
朱元璋建立明朝，结束元朝统治。

第二章 永乐盛世
朱棣迁都北京，开创永乐盛世。

第三章 土木堡之变
明英宗被俘，明朝国力开始衰落。`,
			expected: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建临时文件
			dir := t.TempDir()
			filePath := writeTempFile(t, dir, "test.txt", tc.content)

			opts := Option{
				ChunkSize:    2000,
				ChunkOverlap: 200,
				Separator:    "\n\n",
			}

			chunks, err := Split(ctx, filePath, opts)
			require.NoError(t, err)
			require.NotEmpty(t, chunks)

			t.Logf("%s (%s): 分割为 %d 个块", tc.name, tc.description, len(chunks))

			// 检查是否有章节标题的迹象
			chapterIndicators := 0
			for i, chunk := range chunks {
				// 检查是否包含章节关键词
				chunkLower := strings.ToLower(chunk)
				if strings.Contains(chunkLower, "第") &&
					(strings.Contains(chunkLower, "章") ||
						strings.Contains(chunkLower, "回") ||
						strings.Contains(chunkLower, "节")) {
					chapterIndicators++
					if i < 5 { // 只打印前5个可能的章节
						preview := strings.TrimSpace(chunk)
						if len(preview) > 200 {
							preview = preview[:200] + "..."
						}
						t.Logf("可能的章节标题 (块 %d): %s", i+1, preview)
					}
				}
			}

			t.Logf("检测到 %d 个可能的章节标题", chapterIndicators)
			require.Equal(t, tc.expected, len(chunks), "章节数量不匹配")
		})
	}
}

// TestSplitByChapters_Unit 单独测试章节分割函数
func TestSplitByChapters_Unit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testCases := []struct {
		name     string
		content  string
		expected int // 期望的章节数量
	}{
		{
			name:     "标准章节格式",
			content:  `第一章 开始这是内容。第二章 发展这是内容。第三章 结束这是内容。`,
			expected: 3,
		},
		{
			name: "数字章节格式",
			content: `第1章 标题1
内容1

第2章 标题2  
内容2`,
			expected: 2,
		},
		{
			name: "回目格式",
			content: `第一回 开篇
内容

第二回 发展
内容`,
			expected: 2,
		},
		{
			name: "无章节结构",
			content: `这是一段普通的文本。
没有章节结构。
只有段落。`,
			expected: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("测试内容: %s", tc.content)
			chunks := splitByChapters(ctx, tc.content)
			t.Logf("实际分割结果: %d 个章节，期望: %d 个", len(chunks), tc.expected)

			for i, chunk := range chunks {
				require.NotEmpty(t, strings.TrimSpace(chunk), "章节 %d 为空", i+1)
				chunkPreview := strings.TrimSpace(chunk)
				if len(chunkPreview) > 100 {
					chunkPreview = chunkPreview[:100] + "..."
				}
				t.Logf("章节 %d: %s", i+1, chunkPreview)
			}

			require.Equal(t, tc.expected, len(chunks), "章节数量不匹配")
		})
	}
}
