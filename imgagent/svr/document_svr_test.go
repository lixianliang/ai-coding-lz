package svr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"imgagent/api"
	"imgagent/bailian"
	"imgagent/db"
	"imgagent/pkg/logger"
	"imgagent/proto"
	"imgagent/storage"
)

func setupTestService(t *testing.T) (*Service, func()) {
	// 初始化日志
	logConf := logger.Config{
		Level: "debug",
	}
	_, err := logger.New(logConf)
	require.NoError(t, err)

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "imgagent-test-*")
	require.NoError(t, err)

	// 使用 SQLite 内存数据库进行测试
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移表结构
	err = gormDB.AutoMigrate(&db.Document{}, &db.Chapter{})
	require.NoError(t, err)

	database := &db.Database{}
	database.SetDB(gormDB)

	// 创建 bailian 客户端（如果环境变量设置了 API key）
	var bailianClient *bailian.Client
	if apiKey := os.Getenv("BAILIAN_API_KEY"); apiKey != "" {
		bailianConfig := bailian.Config{
			BaseURL:        "https://dashscope.aliyuncs.com",
			APIKey:         apiKey,
			RequestTimeout: 30,
			MaxRetries:     3,
		}
		var err error
		bailianClient, err = bailian.NewClient(bailianConfig)
		require.NoError(t, err)
	}

	// 创建测试 service
	service := &Service{
		conf: Config{
			APIVersion: "/v1",
			Temp:       tempDir,
			Storage:    storage.Config{},
		},
		db:            database,
		bailianClient: bailianClient,
	}

	// 返回清理函数
	cleanup := func() {
		os.RemoveAll(tempDir)
		sqlDB, _ := gormDB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	return service, cleanup
}

func TestDocumentCRUD(t *testing.T) {
	// 注意：此测试需要真实的 Bailian API key，设置环境变量 BAILIAN_API_KEY 来指定
	if os.Getenv("BAILIAN_API_KEY") == "" {
		t.Skip("跳过测试：需要设置环境变量 BAILIAN_API_KEY")
		return
	}

	service, cleanup := setupTestService(t)
	defer cleanup()

	router := service.RegisterRouter(os.Stdout)

	var createdDocID string
	var createdChapterID string

	t.Run("1. 创建文档 - 上传文件", func(t *testing.T) {
		// 查找测试文件
		testFile := "../books/骆驼祥子.txt"
		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			testFile = "../../books/骆驼祥子.txt"
		}
		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			t.Skip("测试文件不存在，跳过测试")
			return
		}

		// 创建 multipart form
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// 添加 name 字段
		err := writer.WriteField("name", "骆驼祥子")
		require.NoError(t, err)

		// 添加 file 字段
		file, err := os.Open(testFile)
		require.NoError(t, err)
		defer file.Close()

		part, err := writer.CreateFormFile("file", filepath.Base(testFile))
		require.NoError(t, err)
		_, err = io.Copy(part, file)
		require.NoError(t, err)

		err = writer.Close()
		require.NoError(t, err)

		// 发送请求
		req := httptest.NewRequest(http.MethodPost, "/v1/documents", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp proto.BaseResponse
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.Code)

		// 解析返回的文档数据
		docData, err := json.Marshal(resp.Data)
		require.NoError(t, err)
		var doc api.Document
		err = json.Unmarshal(docData, &doc)
		require.NoError(t, err)

		assert.NotEmpty(t, doc.ID)
		assert.Equal(t, "骆驼祥子", doc.Name)
		assert.Equal(t, "chapterReady", doc.Status)
		assert.NotEmpty(t, doc.CreatedAt)
		assert.NotEmpty(t, doc.UpdatedAt)

		createdDocID = doc.ID
		zap.S().Infof("创建文档成功，ID: %s", createdDocID)
	})

	t.Run("2. 获取文档详情", func(t *testing.T) {
		if createdDocID == "" {
			t.Skip("没有创建的文档 ID，跳过测试")
			return
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/documents/%s", createdDocID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp proto.BaseResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.Code)

		docData, err := json.Marshal(resp.Data)
		require.NoError(t, err)
		var doc api.Document
		err = json.Unmarshal(docData, &doc)
		require.NoError(t, err)

		assert.Equal(t, createdDocID, doc.ID)
		assert.Equal(t, "骆驼祥子", doc.Name)
		zap.S().Infof("获取文档详情成功: %s", doc.Name)
	})

	t.Run("3. 列出所有文档", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/documents", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp proto.BaseResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.Code)

		docData, err := json.Marshal(resp.Data)
		require.NoError(t, err)
		var result api.ListDocumentsResult
		err = json.Unmarshal(docData, &result)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(result.Documents), 1)
		zap.S().Infof("列出文档数量: %d", len(result.Documents))
	})

	t.Run("4. 更新文档名称", func(t *testing.T) {
		if createdDocID == "" {
			t.Skip("没有创建的文档 ID，跳过测试")
			return
		}

		updateArgs := api.UpdateDocumentArgs{
			Name: "骆驼祥子-更新版",
		}
		body, err := json.Marshal(updateArgs)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/documents/%s", createdDocID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp proto.BaseResponse
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.Code)

		docData, err := json.Marshal(resp.Data)
		require.NoError(t, err)
		var doc api.Document
		err = json.Unmarshal(docData, &doc)
		require.NoError(t, err)

		assert.Equal(t, "骆驼祥子-更新版", doc.Name)
		zap.S().Infof("更新文档名称成功: %s", doc.Name)
	})

	t.Run("5. 获取章节列表", func(t *testing.T) {
		if createdDocID == "" {
			t.Skip("没有创建的文档 ID，跳过测试")
			return
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/documents/%s/chapters", createdDocID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp proto.BaseResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.Code)

		chapterData, err := json.Marshal(resp.Data)
		require.NoError(t, err)
		var result api.ListChaptersResult
		err = json.Unmarshal(chapterData, &result)
		require.NoError(t, err)

		assert.Greater(t, len(result.Chapters), 0)
		if len(result.Chapters) > 0 {
			createdChapterID = result.Chapters[0].ID
			zap.S().Infof("章节列表数量: %d, 第一个章节ID: %s", len(result.Chapters), createdChapterID)
		}
	})

	t.Run("6. 获取章节详情", func(t *testing.T) {
		if createdDocID == "" || createdChapterID == "" {
			t.Skip("没有文档或章节 ID，跳过测试")
			return
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/documents/%s/chapters/%s", createdDocID, createdChapterID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp proto.BaseResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.Code)

		chapterData, err := json.Marshal(resp.Data)
		require.NoError(t, err)
		var chapter api.Chapter
		err = json.Unmarshal(chapterData, &chapter)
		require.NoError(t, err)

		assert.Equal(t, createdChapterID, chapter.ID)
		assert.Equal(t, createdDocID, chapter.DocumentID)
		assert.NotEmpty(t, chapter.Content)
		zap.S().Infof("获取章节详情成功, 内容长度: %d", len(chapter.Content))
	})

	t.Run("7. 更新章节内容", func(t *testing.T) {
		if createdDocID == "" || createdChapterID == "" {
			t.Skip("没有文档或章节 ID，跳过测试")
			return
		}

		updateArgs := api.UpdateChapterArgs{
			Content: "这是更新后的章节内容，用于测试。",
		}
		body, err := json.Marshal(updateArgs)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/documents/%s/chapters/%s", createdDocID, createdChapterID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp proto.BaseResponse
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.Code)

		chapterData, err := json.Marshal(resp.Data)
		require.NoError(t, err)
		var chapter api.Chapter
		err = json.Unmarshal(chapterData, &chapter)
		require.NoError(t, err)

		assert.Equal(t, updateArgs.Content, chapter.Content)
		zap.S().Infof("更新章节内容成功")
	})

	t.Run("8. 删除章节", func(t *testing.T) {
		if createdDocID == "" || createdChapterID == "" {
			t.Skip("没有文档或章节 ID，跳过测试")
			return
		}

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v1/documents/%s/chapters/%s", createdDocID, createdChapterID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp proto.BaseResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.Code)

		zap.S().Infof("删除章节成功")
	})

	t.Run("9. 删除文档", func(t *testing.T) {
		if createdDocID == "" {
			t.Skip("没有创建的文档 ID，跳过测试")
			return
		}

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v1/documents/%s", createdDocID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp proto.BaseResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.Code)

		zap.S().Infof("删除文档成功")

		// 验证文档已被删除
		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/documents/%s", createdDocID), nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp2 proto.BaseResponse
		err = json.Unmarshal(w.Body.Bytes(), &resp2)
		require.NoError(t, err)
		assert.Equal(t, ErrNoSuchDocumentCode, resp2.Code)
		zap.S().Infof("验证文档已删除")
	})
}

func TestErrorCases(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	router := service.RegisterRouter(os.Stdout)

	t.Run("创建文档 - 缺少 name", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/v1/documents", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp proto.BaseResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("创建文档 - 缺少 file", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.WriteField("name", "测试文档")
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/v1/documents", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp proto.BaseResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("获取不存在的文档", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/documents/nonexistent", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp proto.BaseResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, ErrNoSuchDocumentCode, resp.Code)
	})

	t.Run("更新不存在的文档", func(t *testing.T) {
		updateArgs := api.UpdateDocumentArgs{
			Name: "新名称",
		}
		body, _ := json.Marshal(updateArgs)

		req := httptest.NewRequest(http.MethodPut, "/v1/documents/nonexistent", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp proto.BaseResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, ErrNoSuchDocumentCode, resp.Code)
	})
}

// TestCreateDocumentWithSampleFile 使用小文件测试创建文档
// 注意：此测试需要真实的 Bailian API key，设置环境变量 BAILIAN_API_KEY 来指定
func TestCreateDocumentWithSampleFile(t *testing.T) {
	// 检查是否有真实的 Bailian API key
	if os.Getenv("BAILIAN_API_KEY") == "" {
		t.Skip("跳过测试：需要设置环境变量 BAILIAN_API_KEY")
		return
	}
	service, cleanup := setupTestService(t)
	defer cleanup()

	router := service.RegisterRouter(os.Stdout)

	// 创建临时测试文件
	tempFile, err := os.CreateTemp("", "test-*.txt")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	testContent := "这是一个测试文件的内容。\n\n用于测试文档创建功能。\n\n包含多个段落。"
	_, err = tempFile.WriteString(testContent)
	require.NoError(t, err)
	tempFile.Close()

	// 创建 multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	err = writer.WriteField("name", "测试文档")
	require.NoError(t, err)

	file, err := os.Open(tempFile.Name())
	require.NoError(t, err)
	defer file.Close()

	part, err := writer.CreateFormFile("file", "test.txt")
	require.NoError(t, err)
	_, err = io.Copy(part, file)
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	// 发送请求
	req := httptest.NewRequest(http.MethodPost, "/v1/documents", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp proto.BaseResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.Code, "响应消息: %s", resp.Message)

	docData, err := json.Marshal(resp.Data)
	require.NoError(t, err)
	var doc api.Document
	err = json.Unmarshal(docData, &doc)
	require.NoError(t, err)

	assert.NotEmpty(t, doc.ID)
	assert.Equal(t, "测试文档", doc.Name)
	zap.S().Infof("使用临时文件创建文档成功，ID: %s", doc.ID)
}
