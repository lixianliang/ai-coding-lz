package svr

import (
	"context"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	"gorm.io/gorm"

	"imgagent/api"
	"imgagent/db"
	hutil "imgagent/httputil"
	"imgagent/pkg/logger"
	"imgagent/spliter"
)

const (
	ErrNoSuchDocumentCode   = 612
	ErrExistingDocumentCode = 614
	ErrNoSuchDocument       = "no such document"
	ErrExistingDocument     = "existing document"
)

func (s *Service) HandleCreateDocument(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromGinContext(c)

	name := c.PostForm("name")
	if name == "" {
		hutil.AbortError(c, http.StatusBadRequest, "name is required")
		return
	}
	if len(name) > 50 {
		hutil.AbortError(c, http.StatusBadRequest, "name exceeds maximum length of 50")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		log.Errorf("Failed to get file, err: %v", err)
		hutil.AbortError(c, http.StatusBadRequest, "file is required")
		return
	}

	log.Infof("Create document, name: %s, file: %s", name, file.Filename)

	_, err = s.db.GetDocumentWithName(ctx, name)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Errorf("Failed to get document, err: %v", err)
			hutil.AbortError(c, hutil.ErrServerInternalCode, "get document failed")
			return
		}
	} else {
		log.Warnf("Document existing")
		hutil.AbortError(c, ErrExistingDocumentCode, ErrExistingDocument)
		return
	}

	index := strings.LastIndex(file.Filename, ".")
	if index == -1 {
		hutil.AbortError(c, http.StatusBadRequest, "file has no extension")
		return
	}
	ext := file.Filename[index+1:]
	id := uuid.New()
	uid := hex.EncodeToString(id[:])
	filename := s.conf.Temp + "/" + uid + "." + ext

	err = c.SaveUploadedFile(file, filename)
	if err != nil {
		log.Errorf("Failed to save file, err: %v", err)
		hutil.AbortError(c, hutil.ErrServerInternalCode, "save file failed")
		return
	}
	defer os.Remove(filename)

	chunkOverlap := 100
	texts, err := spliter.Split(ctx, filename, spliter.Option{
		ChunkSize:    2000,
		ChunkOverlap: chunkOverlap,
		Separator:    "\n\n",
	})
	if err != nil {
		log.Errorf("Failed to split text, err: %v", err)
		hutil.AbortError(c, hutil.ErrServerInternalCode, "split text failed")
		return
	}

	docID := db.MakeUUID()
	err = s.db.CreateChapters(ctx, docID, texts)
	if err != nil {
		log.Errorf("Failed to create Chapters, err: %v", err)
		hutil.AbortError(c, hutil.ErrServerInternalCode, "create Chapters failed")
		return
	}

	args := &api.CreateDocumentArgs{
		Name: name,
	}
	doc, err := s.db.CreateDocument(ctx, docID, args)
	if err != nil {
		log.Errorf("Failed to create document, err: %v", err)
		documentErr(c, err, "create document failed")
		return
	}

	hutil.WriteData(c, makeDocument(doc))
}

func (s *Service) HandleGetDocument(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromGinContext(c)
	//ui := GetUserInfo(c)

	docID := c.Param("document_id")
	if docID == "" {
		hutil.AbortError(c, http.StatusBadRequest, "invalid doc id")
		return
	}

	log.Infof("Get document, docID: %s", docID)
	doc, err := s.db.GetDocument(ctx, docID)
	if err != nil {
		log.Errorf("get document failed, id: %s, err: %v", docID, err)
		documentErr(c, err, "get document failed")
		return
	}
	hutil.WriteData(c, makeDocument(&doc))
}

func (s *Service) HandleUpdateDocument(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromGinContext(c)

	docID := c.Param("document_id")
	if docID == "" {
		hutil.AbortError(c, http.StatusBadRequest, "invalid doc id")
		return
	}
	var args api.UpdateDocumentArgs
	if err := c.ShouldBindJSON(&args); err != nil {
		log.Errorf("Invalid request body, err: %v", err)
		hutil.AbortError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	log.Infof("Update document, docID: %s", docID)
	if err := s.db.UpdateDocument(ctx, docID, &args); err != nil {
		log.Errorf("Failed update document failed, id: %s, err: %v", docID, err)
		documentErr(c, err, "update document failed")
		return
	}
	doc, err := s.db.GetDocument(ctx, docID)
	if err != nil {
		log.Errorf("get document failed, id: %s, err: %v", docID, err)
		documentErr(c, err, "get document failed")
		return
	}
	hutil.WriteData(c, makeDocument(&doc))
}

func (s *Service) HandleDeleteDocument(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromGinContext(c)
	// ui := GetUserInfo(c)

	docID := c.Param("document_id")
	if docID == "" {
		hutil.AbortError(c, http.StatusBadRequest, "invalid doc id")
		return
	}

	log.Infof("Delete document, docID: %s", docID)
	// 删除对应的 Chapter
	err := s.db.DeleteAllChapter(ctx, docID)
	if err != nil {
		log.Errorf("Failed to delete document Chapter, err: %v", err)
		hutil.AbortError(c, hutil.ErrServerInternalCode, "delete document Chapter failed")
	}
	err = s.db.DeleteDocument(ctx, docID)
	if err != nil {
		log.Errorf("Failed to delete document, err: %v", err)
		hutil.AbortError(c, hutil.ErrServerInternalCode, "delete document failed")
		return
	}
	hutil.WriteData(c, nil)
}

func (s *Service) HandleListDocuments(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromGinContext(c)
	// ui := GetUserInfo(c)

	log.Infof("List documents")
	docs, err := s.db.ListDocuments(ctx)
	if err != nil {
		log.Errorf("Failed to list documents, err: %v", err)
		hutil.AbortError(c, hutil.ErrServerInternalCode, "list documents failed")
		return
	}

	ret := &api.ListDocumentsResult{}
	for _, d := range docs {
		ret.Documents = append(ret.Documents, makeDocument(&d))
	}
	hutil.WriteData(c, ret)
}

func (s *Service) HandleGetChapter(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromGinContext(c)

	docID := c.Param("document_id")
	id := c.Param("id")
	if docID == "" {
		hutil.AbortError(c, http.StatusBadRequest, "invalid doc id")
		return
	}
	if id == "" {
		hutil.AbortError(c, http.StatusBadRequest, "invalid id")
		return
	}

	log.Infof("Get Chapter, docID: %s, id: %s", docID, id)
	Chapter, err := s.db.GetChapter(ctx, id, docID)
	if err != nil {
		log.Errorf("Failed to get Chapter, err: %v", err)
		hutil.AbortError(c, http.StatusInternalServerError, "get Chapter failed")
		return
	}

	hutil.WriteData(c, makeChapter(&Chapter))
}

func (s *Service) HandleUpdateChapter(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromGinContext(c)

	docID := c.Param("document_id")
	id := c.Param("id")
	if docID == "" {
		hutil.AbortError(c, http.StatusBadRequest, "invalid doc id")
		return
	}
	if id == "" {
		hutil.AbortError(c, http.StatusBadRequest, "invalid id")
		return
	}

	var args api.UpdateChapterArgs
	if err := c.ShouldBindJSON(&args); err != nil {
		log.Errorf("Invalid request body, err: %v", err)
		hutil.AbortError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	log.Infof("Update Chapter, docID: %s, id: %s", docID, id)
	err := s.db.UpdateChapter(ctx, id, &args)
	if err != nil {
		log.Errorf("Failed to update db Chapter, err: %v", err)
		hutil.AbortError(c, http.StatusInternalServerError, "update Chapter failed")
		return
	}
	Chapter, err := s.db.GetChapter(ctx, id, docID)
	if err != nil {
		log.Errorf("Failed to get Chapter, err: %v", err)
		hutil.AbortError(c, http.StatusInternalServerError, "get Chapter failed")
		return
	}

	hutil.WriteData(c, makeChapter(&Chapter))
}

func (s *Service) HandleDeleteChapter(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromGinContext(c)

	docID := c.Param("document_id")
	id := c.Param("id")
	if docID == "" {
		hutil.AbortError(c, http.StatusBadRequest, "invalid doc id")
		return
	}
	if id == "" {
		hutil.AbortError(c, http.StatusBadRequest, "invalid id")
		return
	}

	log.Infof("Delete Chapter, docID: %s, id: %s", docID, id)
	err := s.db.DeleteChapter(ctx, id, docID)
	if err != nil {
		log.Errorf("Failed to delete db Chapter, err: %v", err)
		hutil.AbortError(c, http.StatusInternalServerError, "delete Chapter failed")
		return
	}

	hutil.WriteData(c, nil)
}

func (s *Service) HandleListChapters(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromGinContext(c)

	docID := c.Param("document_id")
	if docID == "" {
		hutil.AbortError(c, http.StatusBadRequest, "invalid doc id")
		return
	}

	// todo： 后续需要考虑分页
	log.Infof("List Chapters, docID: %s", docID)
	Chapters, err := s.db.ListChapters(ctx, docID)
	if err != nil {
		log.Errorf("list Chapters failed, err: %v", err)
		hutil.AbortError(c, http.StatusBadRequest, "list Chapters failed")
		return
	}

	result := &api.ListChaptersResult{}
	for _, seg := range Chapters {
		result.Chapters = append(result.Chapters, makeChapter(&seg))
	}
	hutil.WriteData(c, result)
}

func makeDocument(d *db.Document) api.Document {
	return api.Document{
		ID:        d.ID,
		Name:      d.Name,
		Status:    d.Status,
		CreatedAt: d.CreatedAt.Format(time.DateTime),
		UpdatedAt: d.UpdatedAt.Format(time.DateTime),
	}
}

func makeChapter(d *db.Chapter) api.Chapter {
	return api.Chapter{
		ID:         d.ID,
		DocumentID: d.DocumentID,
		Index:      d.Index,
		Content:    d.Content,
		CreatedAt:  d.CreatedAt.Format(time.DateTime),
		UpdatedAt:  d.UpdatedAt.Format(time.DateTime),
	}
}

func documentErr(c *gin.Context, err error, errMsg string) {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		hutil.AbortError(c, ErrExistingDocumentCode, ErrExistingDocument)
		return
	}
	// sqlite for test
	if sqliteErr, ok := err.(sqlite3.Error); ok {
		if sqliteErr.Code == 19 && sqliteErr.ExtendedCode == 2067 {
			hutil.AbortError(c, ErrExistingDocumentCode, ErrExistingDocument)
			return
		}
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		hutil.AbortError(c, ErrNoSuchDocumentCode, ErrNoSuchDocument)
	} else {
		hutil.AbortError(c, hutil.ErrServerInternalCode, errMsg)
	}
}

func (s *Service) downloadFile(ctx context.Context, textURL string) (string, error) {
	log := logger.FromContext(ctx)

	url, err := url.ParseRequestURI(textURL)
	if err != nil {
		return "", err
	}
	index := strings.LastIndex(url.Path, ".")
	if index == -1 {
		return "", errors.New("unknown ext")
	}
	ext := url.Path[index+1:]
	id := uuid.New()
	uid := hex.EncodeToString(id[:])
	filename := s.conf.Temp + "/" + uid + "." + ext
	resp, err := http.Get(textURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warnf("Failed to get %s, code: %d", textURL, resp.StatusCode)
		return "", errors.New("unexpected status code")
	}

	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	n, err := io.Copy(file, resp.Body)
	if err != nil {
		os.Remove(filename)
		return "", err
	}

	log.Infof("Download url %s, filename: %s, n: %d", textURL, filename, n)
	return filename, nil
}
