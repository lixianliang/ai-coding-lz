package spliter

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	worddoc "baliance.com/gooxml/document"
	"github.com/ledongthuc/pdf"
	"github.com/tmc/langchaingo/textsplitter"

	"imgagent/pkg/logger"
)

type Option struct {
	ChunkSize    int
	ChunkOverlap int
	Separator    string
}

func Split(ctx context.Context, filename string, opt Option) ([]string, error) {
	var content string

	start := time.Now()
	log := logger.FromContext(ctx)
	separators := []string{"\n\n", "\n", " ", ""}
	if opt.Separator == "\n" {
		separators = []string{"\n", " ", ""}
	}

	ext := filepath.Ext(filename)
	switch ext {
	case ".txt", ".md":
		bytes, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		content = string(bytes)
	case ".doc", ".docx":
		d, err := worddoc.Open(filename)
		if err != nil {
			return nil, err
		}
		for _, para := range d.Paragraphs() {
			for _, run := range para.Runs() {
				content += run.Text()
			}
			content += "\n"
		}
	case ".pdf":
		f, r, err := pdf.Open(filename)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		var buf bytes.Buffer
		// 获取 pdf 文本数据
		pt, err := r.GetPlainText()
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(&buf, pt); err != nil {
			return nil, err
		}
		content = buf.String()
	default:
		return nil, errors.New("unknown file ext")
	}
	if content == "" {
		return nil, errors.New("empty content")
	}

	// 3. 创建文本分割器
	var err error
	var texts []string
	var splitter textsplitter.TextSplitter
	if ext == ".md" {
		mdSparators := []string{"#", "##", "###", "####"}
		mdSparators = append(mdSparators, separators...)
		splitter = textsplitter.NewMarkdownTextSplitter(
			textsplitter.WithChunkSize(opt.ChunkSize),
			textsplitter.WithChunkOverlap(opt.ChunkOverlap),
			textsplitter.WithSeparators(mdSparators),
		)
		texts, err = splitter.SplitText(content)
		if err != nil {
			return nil, err
		}
	} else {
		splitter = textsplitter.NewRecursiveCharacter(
			textsplitter.WithChunkSize(opt.ChunkSize),
			textsplitter.WithChunkOverlap(opt.ChunkOverlap),
			textsplitter.WithSeparators(separators),
		)
		// 使用 SplitText 方法分割文本内容
		texts, err = splitText(ctx, splitter, content, opt.Separator, opt.ChunkSize)
		if err != nil {
			return nil, err
		}
	}

	// 数据清洗
	for i, text := range texts {
		// 去掉空白符号
		text = strings.TrimSpace(text)
		// 替换中间换行符
		texts[i] = strings.ReplaceAll(text, "\n", ",")
		log.Debugf("Splite content, i: %d, len: %d,  %s", i, len(texts[i]), texts[i][:min(48, len(texts[i]))])
	}
	log.Infof("Split costMS: %d", time.Since(start).Milliseconds())
	return texts, nil
}

func splitText(ctx context.Context, splitter textsplitter.TextSplitter, content string, separator string, chunkSize int) ([]string, error) {
	log := logger.FromContext(ctx)

	// 优先按章节分割
	chapterChunks := splitByChapters(ctx, content)
	if len(chapterChunks) > 1 {
		log.Infof("按章节分割成功，共 %d 个章节", len(chapterChunks))
		return chapterChunks, nil
	}

	// 如果章节分割失败，使用传统方式分割
	log.Infof("章节分割失败，使用传统方式分割")
	finalChunks := make([]string, 0)
	sep := "\n\n"
	if separator == "\n" {
		sep = separator
	}
	splits := strings.Split(content, sep)
	// 若 \n\n 未分割文本则尝试 \r\n\r\n (windows 换行符)
	if sep == "\n\n" && len(splits) == 1 {
		splits = strings.Split(content, "\r\n\r\n")
	}
	for _, split := range splits {
		// 去掉空白符号
		split = strings.TrimSpace(split)
		if split == "" {
			continue
		}
		if utf8.RuneCountInString(split) > chunkSize {
			texts, err := splitter.SplitText(split)
			if err != nil {
				finalChunks = append(finalChunks, split)
				log.Warnf("Failed to split text, err: %v", err)
			} else {
				finalChunks = append(finalChunks, texts...)
			}
			continue
		}

		finalChunks = append(finalChunks, split)
	}
	return finalChunks, nil
}

// splitByChapters 按章节分割文本
func splitByChapters(ctx context.Context, content string) []string {
	log := logger.FromContext(ctx)

	// 定义章节匹配的正则表达式
	chapterPatterns := []string{
		// 第X章、第X回、第X节
		`(?i)(第[一二三四五六七八九十百千万\d]+[章节回节])`,
		// 第X章 标题
		`(?i)(第[一二三四五六七八九十百千万\d]+章\s*[^\n]*)`,
		// 第X回 标题
		`(?i)(第[一二三四五六七八九十百千万\d]+回\s*[^\n]*)`,
		// 第X节 标题
		`(?i)(第[一二三四五六七八九十百千万\d]+节\s*[^\n]*)`,
		// 数字章节
		`(?i)(第\d+[章节回节])`,
		// 纯数字章节
		`(?i)(第\d+章\s*[^\n]*)`,
		// 英文章节
		`(?i)(Chapter\s+\d+)`,
		// 罗马数字章节
		`(?i)(第[IVX]+[章节回节])`,
	}

	var chapterRegex *regexp.Regexp
	var bestPattern string

	// 尝试不同的章节模式，找到匹配最多的
	maxMatches := 0
	for _, pattern := range chapterPatterns {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		matches := regex.FindAllString(content, -1)
		if len(matches) > maxMatches {
			maxMatches = len(matches)
			chapterRegex = regex
			bestPattern = pattern
		}
	}

	// 如果找到章节模式且匹配数量大于1，进行分割
	if chapterRegex != nil && maxMatches > 1 {
		log.Infof("找到章节模式: %s，匹配到 %d 个章节", bestPattern, maxMatches)

		// 打印所有匹配的章节标题
		matches := chapterRegex.FindAllString(content, -1)
		for i, match := range matches {
			log.Infof("章节 %d: %s", i+1, strings.TrimSpace(match))
		}

		// 按章节分割 - 简单直接的方法
		// 找到所有章节标题的位置
		indices := chapterRegex.FindAllStringIndex(content, -1)
		var result []string

		log.Infof("找到 %d 个章节标题位置", len(indices))
		for i, idx := range indices {
			log.Infof("章节 %d 位置: %d-%d, 内容: %s", i+1, idx[0], idx[1], content[idx[0]:idx[1]])
		}

		if len(indices) == 0 {
			return []string{content}
		}

		// 从第一个章节标题开始分割
		start := indices[0][0]
		for i := 0; i < len(indices); i++ {
			var end int
			if i+1 < len(indices) {
				end = indices[i+1][0] // 下一个章节标题开始位置
			} else {
				end = len(content) // 最后一个章节到结尾
			}

			chapter := strings.TrimSpace(content[start:end])
			log.Infof("章节 %d: 位置 %d-%d, 内容: %s", i+1, start, end, chapter[:min(50, len(chapter))])
			if chapter != "" {
				result = append(result, chapter)
			}
			start = end
		}

		return result
	}

	log.Infof("未找到有效的章节模式，匹配数量: %d", maxMatches)
	return []string{content}
}
