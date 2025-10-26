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
	"imgagent/bailian"
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

	// 生成文档 ID
	docID := db.MakeUUID()

	// 保存临时文件用于分割
	tempFilename := s.conf.Temp + "/" + docID + "_temp." + ext
	err = c.SaveUploadedFile(file, tempFilename)
	if err != nil {
		log.Errorf("Failed to save temp file, err: %v", err)
		hutil.AbortError(c, hutil.ErrServerInternalCode, "save file failed")
		return
	}
	defer os.Remove(tempFilename) // 临时文件使用后删除

	// 分割章节
	chunkOverlap := 100
	texts, err := spliter.Split(ctx, tempFilename, spliter.Option{
		ChunkSize:    5000,
		ChunkOverlap: chunkOverlap,
		Separator:    "\n\n",
	})
	if err != nil {
		log.Errorf("Failed to split text, err: %v", err)
		hutil.AbortError(c, hutil.ErrServerInternalCode, "split text failed")
		return
	}

	err = s.db.CreateChapters(ctx, docID, texts)
	if err != nil {
		log.Errorf("Failed to create chapters, err: %v", err)
		hutil.AbortError(c, hutil.ErrServerInternalCode, "create chapters failed")
		return
	}

	// 上传文件到百炼
	log.Infof("Uploading file to Bailian, filename: %s", tempFilename)
	fileID, err := s.bailianClient.UploadFile(ctx, tempFilename)
	if err != nil {
		log.Errorf("Failed to upload file to Bailian, doc: %s, filename: %s, err: %v", docID, tempFilename, err)
		hutil.AbortError(c, hutil.ErrServerInternalCode, "upload file to Bailian failed")
		return
	}

	args := &api.CreateDocumentArgs{
		Name: name,
	}
	doc, err := s.db.CreateDocument(ctx, docID, fileID, args)
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
	log.Infof("List chapters, docID: %s", docID)
	chapters, err := s.db.ListChapters(ctx, docID)
	if err != nil {
		log.Errorf("list chapters failed, err: %v", err)
		hutil.AbortError(c, http.StatusBadRequest, "list chapters failed")
		return
	}

	result := &api.ListChaptersResult{}
	for _, seg := range chapters {
		result.Chapters = append(result.Chapters, makeChapter(&seg))
	}
	hutil.WriteData(c, result)
}

func makeDocument(d *db.Document) api.Document {
	return api.Document{
		ID:               d.ID,
		Name:             d.Name,
		FileID:           d.FileID,
		SummaryImageURL:  d.SummaryImageURL,
		Status:           d.Status,
		CreatedAt:        d.CreatedAt.Format(time.DateTime),
		UpdatedAt:        d.UpdatedAt.Format(time.DateTime),
	}
}

func makeChapter(d *db.Chapter) api.Chapter {
	return api.Chapter{
		ID:         d.ID,
		DocumentID: d.DocumentID,
		Index:      d.Index,
		Title:      d.Title,
		Content:    d.Content,
		SceneIDs:   d.SceneIDs,
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

// HandleGetRoles 获取文档的角色列表
func (s *Service) HandleGetRoles(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromGinContext(c)

	docID := c.Param("document_id")
	if docID == "" {
		hutil.AbortError(c, http.StatusBadRequest, "invalid doc id")
		return
	}

	log.Infof("Get roles, docID: %s", docID)
	roles, err := s.db.ListRolesByDocument(ctx, docID)
	if err != nil {
		log.Errorf("Failed to list roles, err: %v", err)
		hutil.AbortError(c, http.StatusInternalServerError, "list roles failed")
		return
	}

	result := &api.ListRolesResult{}
	for _, role := range roles {
		result.Roles = append(result.Roles, makeRole(&role))
	}
	hutil.WriteData(c, result)
}

// HandleListScenesByDocument 获取文档的所有场景
func (s *Service) HandleListScenesByDocument(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromGinContext(c)

	docID := c.Param("document_id")
	if docID == "" {
		hutil.AbortError(c, http.StatusBadRequest, "invalid doc id")
		return
	}

	log.Infof("List scenes by document, docID: %s", docID)
	scenes, err := s.db.ListScenesByDocument(ctx, docID)
	if err != nil {
		log.Errorf("Failed to list scenes, err: %v", err)
		hutil.AbortError(c, http.StatusInternalServerError, "list scenes failed")
		return
	}

	result := &api.ListScenesResult{}
	for _, scene := range scenes {
		result.Scenes = append(result.Scenes, makeScene(&scene))
	}
	hutil.WriteData(c, result)
}

// HandleListScenesByChapter 获取章节的场景列表
func (s *Service) HandleListScenesByChapter(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromGinContext(c)

	chapterID := c.Param("chapter_id")
	if chapterID == "" {
		hutil.AbortError(c, http.StatusBadRequest, "invalid chapter id")
		return
	}

	log.Infof("List scenes by chapter, chapterID: %s", chapterID)
	scenes, err := s.db.ListScenesByChapter(ctx, chapterID)
	if err != nil {
		log.Errorf("Failed to list scenes, err: %v", err)
		hutil.AbortError(c, http.StatusInternalServerError, "list scenes failed")
		return
	}

	result := &api.ListScenesResult{}
	for _, scene := range scenes {
		result.Scenes = append(result.Scenes, makeScene(&scene))
	}
	hutil.WriteData(c, result)
}

func makeRole(r *db.Role) api.Role {
	return api.Role{
		ID:         r.ID,
		DocumentID: r.DocumentID,
		Name:       r.Name,
		Gender:     r.Gender,
		Character:  r.Character,
		Appearance: r.Appearance,
		CreatedAt:  r.CreatedAt.Format(time.DateTime),
		UpdatedAt:  r.UpdatedAt.Format(time.DateTime),
	}
}

func makeScene(s *db.Scene) api.Scene {
	return api.Scene{
		ID:         s.ID,
		ChapterID:  s.ChapterID,
		DocumentID: s.DocumentID,
		Index:      s.Index,
		Content:    s.Content,
		ImageURL:   s.ImageURL,
		VoiceURL:   s.VoiceURL,
		CreatedAt:  s.CreatedAt.Format(time.DateTime),
		UpdatedAt:  s.UpdatedAt.Format(time.DateTime),
	}
}

// HandleUpdateRole 更新角色信息
func (s *Service) HandleUpdateRole(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromGinContext(c)

	roleID := c.Param("id")
	if roleID == "" {
		hutil.AbortError(c, http.StatusBadRequest, "invalid role id")
		return
	}

	var args api.UpdateRoleArgs
	if err := c.ShouldBindJSON(&args); err != nil {
		log.Errorf("Invalid request body, err: %v", err)
		hutil.AbortError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	log.Infof("Update role, roleID: %s", roleID)
	err := s.db.UpdateRole(ctx, roleID, &args)
	if err != nil {
		log.Errorf("Failed to update role, err: %v", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			hutil.AbortError(c, http.StatusNotFound, "role not found")
		} else {
			hutil.AbortError(c, http.StatusInternalServerError, "update role failed")
		}
		return
	}

	role, err := s.db.GetRole(ctx, roleID)
	if err != nil {
		log.Errorf("Failed to get role, err: %v", err)
		hutil.AbortError(c, http.StatusInternalServerError, "get role failed")
		return
	}

	hutil.WriteData(c, makeRole(&role))
}

// HandleUpdateScene 更新场景内容，立即重新生成图片和语音
func (s *Service) HandleUpdateScene(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromGinContext(c)

	sceneID := c.Param("id")
	if sceneID == "" {
		hutil.AbortError(c, http.StatusBadRequest, "invalid scene id")
		return
	}

	var args api.UpdateSceneArgs
	if err := c.ShouldBindJSON(&args); err != nil {
		log.Errorf("Invalid request body, err: %v", err)
		hutil.AbortError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	// 1. 获取场景信息
	scene, err := s.db.GetScene(ctx, sceneID)
	if err != nil {
		log.Errorf("Failed to get scene, err: %v", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			hutil.AbortError(c, http.StatusNotFound, "scene not found")
		} else {
			hutil.AbortError(c, http.StatusInternalServerError, "get scene failed")
		}
		return
	}

	// 2. 更新场景内容
	log.Infof("Update scene content, sceneID: %s", sceneID)
	err = s.db.UpdateScene(ctx, sceneID, &args)
	if err != nil {
		log.Errorf("Failed to update scene, err: %v", err)
		hutil.AbortError(c, http.StatusInternalServerError, "update scene failed")
		return
	}

	// 3. 获取文档信息（需要摘要和角色信息）
	doc, err := s.db.GetDocument(ctx, scene.DocumentID)
	if err != nil {
		log.Errorf("Failed to get document, err: %v", err)
		hutil.AbortError(c, http.StatusInternalServerError, "get document failed")
		return
	}

	// 4. 获取角色信息
	dbRoles, err := s.db.ListRolesByDocument(ctx, doc.ID)
	if err != nil {
		log.Errorf("Failed to list roles, err: %v", err)
		hutil.AbortError(c, http.StatusInternalServerError, "list roles failed")
		return
	}

	// 转换为 bailian.RoleInfo
	roles := make([]bailian.RoleInfo, 0, len(dbRoles))
	for _, r := range dbRoles {
		roles = append(roles, bailian.RoleInfo{
			Name:       r.Name,
			Gender:     r.Gender,
			Character:  r.Character,
			Appearance: r.Appearance,
		})
	}

	// 5. 生成图片
	log.Infof("Generating image for scene, sceneID: %s", sceneID)
	imageURL, err := s.bailianClient.GenerateImage(ctx, args.Content, doc.Summary, roles)
	if err != nil {
		log.Errorf("Failed to generate image, scene: %s, err: %v", sceneID, err)
		hutil.AbortError(c, http.StatusInternalServerError, "generate image failed")
		return
	}

	// 更新图片 URL
	err = s.db.UpdateSceneImageURL(ctx, sceneID, imageURL)
	if err != nil {
		log.Errorf("Failed to update scene imageURL, err: %v", err)
		hutil.AbortError(c, http.StatusInternalServerError, "update image failed")
		return
	}

	log.Infof("Image generated for scene: %s, URL: %s", sceneID, imageURL)

	// 6. 生成语音
	log.Infof("Generating TTS for scene, sceneID: %s", sceneID)
	voiceURL, err := s.bailianClient.GenerateTTS(ctx, args.Content)
	if err != nil {
		log.Errorf("Failed to generate TTS, scene: %s, err: %v", sceneID, err)
		hutil.AbortError(c, http.StatusInternalServerError, "generate voice failed")
		return
	}

	// 更新语音 URL
	err = s.db.UpdateSceneVoiceURL(ctx, sceneID, voiceURL)
	if err != nil {
		log.Errorf("Failed to update scene voiceURL, err: %v", err)
		hutil.AbortError(c, http.StatusInternalServerError, "update voice failed")
		return
	}

	log.Infof("Voice generated for scene: %s, URL: %s", sceneID, voiceURL)

	// 7. 返回更新后的场景
	scene, err = s.db.GetScene(ctx, sceneID)
	if err != nil {
		log.Errorf("Failed to get scene, err: %v", err)
		hutil.AbortError(c, http.StatusInternalServerError, "get scene failed")
		return
	}

	log.Infof("Scene updated and regenerated, sceneID: %s", sceneID)
	hutil.WriteData(c, makeScene(&scene))
}
