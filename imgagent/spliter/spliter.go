package spliter

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
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
