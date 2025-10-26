package db

import (
	"context"
	"time"

	"gorm.io/gorm"

	"imgagent/api"
)

const (
	batchSize = 100

	DocumentStatusChapterReady = "chapterReady"
	DocumentStatusRoleReady    = "roleReady"
	DocumentStatusSceneReady   = "sceneReady"
	DocumentStatusImgReady     = "imgReady"
)

func (Role) TableName() string {
	return "roles"
}

// Document 文档表
type Document struct {
	ID        string    `gorm:"primaryKey;size:32;comment:'主键'"`
	Name      string    `gorm:"uniqueIndex:uk_name;size:128;comment:'文档名称'"`
	FileID    string    `gorm:"size:255;comment:'存储在阿里云百炼的 fileid'"`
	Summary   string    `gorm:"size:1000;comment:'小说摘要'"`
	Status    string    `gorm:"size:20;comment:'状态 indexing|ready'"`
	CreatedAt time.Time `gorm:"comment:'创建时间'"`
	UpdatedAt time.Time `gorm:"comment:'更新时间'"`
}

func (Document) TableName() string {
	return "documents"
}

// Chapter 章节表
type Chapter struct {
	ID         string    `gorm:"primaryKey;size:32;comment:'主键'"`
	Index      int       `gorm:"uniqueIndex:uk_document_index,priority:2;comment:'章节序号'"`
	DocumentID string    `gorm:"uniqueIndex:uk_document_index,priority:1;size:32;comment:'文档 id'"`
	Title      string    `gorm:"size:100;comment:'标题'"`
	Content    string    `gorm:"size:10000;comment:'章节内容'"`
	SceneIDs   []string  `gorm:"type:json;serializer:json;comment:'故事场景'"`
	CreatedAt  time.Time `gorm:"comment:'创建时间'"`
	UpdatedAt  time.Time `gorm:"comment:'更新时间'"`
}

func (Chapter) TableName() string {
	return "chapters"
}

// Scene 场景表
type Scene struct {
	ID         string    `gorm:"primaryKey;size:32;comment:'主键'"`
	ChapterID  string    `gorm:"index:idx_chapter_id;size:32;comment:'chapter id'"`
	DocumentID string    `gorm:"index:idx_document_id;size:32;comment:'文档 id'"`
	Index      int       `gorm:"comment:'场景序号'"`
	Content    string    `gorm:"size:1000;comment:'场景描述'"`
	ImageURL   string    `gorm:"size:500;comment:'场景图片url'"`
	VoiceURL   string    `gorm:"size:500;comment:'音频url'"`
	CreatedAt  time.Time `gorm:"comment:'创建时间'"`
	UpdatedAt  time.Time `gorm:"comment:'更新时间'"`
}

func (Scene) TableName() string {
	return "scenes"
}

// Role 任务角色表
type Role struct {
	ID         string    `gorm:"primaryKey;size:32;comment:'主键'"`
	DocumentID string    `gorm:"index:idx_role_document_id;size:32;comment:'文档 id'"`
	Name       string    `gorm:"size:50;comment:'角色名字'"`
	Gender     string    `gorm:"size:10;comment:'性别'"`
	Character  string    `gorm:"size:500;comment:'性格特点'"`
	Appearance string    `gorm:"size:500;comment:'外貌描述'"`
	CreatedAt  time.Time `gorm:"comment:'创建时间'"`
	UpdatedAt  time.Time `gorm:"comment:'更新时间'"`
}

// ===== Document DAO =====

func (db *Database) CreateDocument(ctx context.Context, docID, fileID string, args *api.CreateDocumentArgs) (*Document, error) {
	now := time.Now()
	doc := Document{
		ID:        docID,
		FileID:    fileID,
		Name:      args.Name,
		Status:    DocumentStatusChapterReady,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := gorm.G[Document](db.db).Create(ctx, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func (db *Database) GetDocument(ctx context.Context, id string) (Document, error) {
	return gorm.G[Document](db.db).Where("id = ?", id).Take(ctx)
}

func (db *Database) GetDocumentWithName(ctx context.Context, name string) (Document, error) {
	return gorm.G[Document](db.db).Where("name = ?", name).Take(ctx)
}

func (db *Database) UpdateDocument(ctx context.Context, id string, args *api.UpdateDocumentArgs) error {
	now := time.Now()
	doc := Document{
		Name:      args.Name,
		UpdatedAt: now,
	}
	rowsAffected, err := gorm.G[Document](db.db).Where("id = ?", id).Updates(ctx, doc)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (db *Database) UpdateDocumentStatus(ctx context.Context, id string, status string) error {
	rowsAffected, err := gorm.G[Document](db.db).Where("id = ?", id).Update(ctx, "status", status)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (db *Database) DeleteDocument(ctx context.Context, id string) error {
	_, err := gorm.G[Document](db.db).Where("id = ?", id).Delete(ctx)
	return err
}

func (db *Database) ListDocuments(ctx context.Context) ([]Document, error) {
	return gorm.G[Document](db.db).Order("updated_at DESC").Find(ctx)
}

func (db *Database) UpdateDocumentFileID(ctx context.Context, id string, fileID string) error {
	rowsAffected, err := gorm.G[Document](db.db).Where("id = ?", id).Update(ctx, "file_id", fileID)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (db *Database) UpdateDocumentSummary(ctx context.Context, id string, summary string) error {
	rowsAffected, err := gorm.G[Document](db.db).Where("id = ?", id).Update(ctx, "summary", summary)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (db *Database) ListChapterReadyDocuments(ctx context.Context) ([]Document, error) {
	return gorm.G[Document](db.db).Where("status = ?", DocumentStatusChapterReady).Order("created_at ASC").Find(ctx)
}

func (db *Database) ListRoleReadyDocuments(ctx context.Context) ([]Document, error) {
	return gorm.G[Document](db.db).Where("status = ?", DocumentStatusRoleReady).Order("created_at ASC").Find(ctx)
}

func (db *Database) ListSceneReadyDocuments(ctx context.Context) ([]Document, error) {
	return gorm.G[Document](db.db).Where("status = ?", DocumentStatusSceneReady).Order("created_at ASC").Find(ctx)
}

// ===== Chapter DAO =====

func (db *Database) CreateChapters(ctx context.Context, documentID string, texts []string) error {
	var Chapters []Chapter

	now := time.Now()
	for i, text := range texts {
		Chapters = append(Chapters, Chapter{
			ID:         MakeUUID(),
			Index:      i,
			DocumentID: documentID,
			Content:    text,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}
	return gorm.G[Chapter](db.db).CreateInBatches(ctx, &Chapters, batchSize)
}

func (db *Database) GetChapter(ctx context.Context, id, documentID string) (Chapter, error) {
	return gorm.G[Chapter](db.db).Where("id = ? AND document_id = ?", id, documentID).Take(ctx)
}

func (db *Database) UpdateChapter(ctx context.Context, id string, args *api.UpdateChapterArgs) error {
	now := time.Now()
	seg := Chapter{
		Content:   args.Content,
		UpdatedAt: now,
	}
	rowsAffected, err := gorm.G[Chapter](db.db).Where("id = ?", id).Updates(ctx, seg)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return err
}

func (db *Database) DeleteChapter(ctx context.Context, id, documentID string) error {
	_, err := gorm.G[Chapter](db.db).Where("id = ? AND document_id = ?", id, documentID).Delete(ctx)
	return err
}

func (db *Database) DeleteAllChapter(ctx context.Context, documentID string) error {
	_, err := gorm.G[Chapter](db.db).Where("document_id = ?", documentID).Delete(ctx)
	return err
}

func (db *Database) ListChapters(ctx context.Context, documentID string) ([]Chapter, error) {
	return gorm.G[Chapter](db.db).Where("document_id = ?", documentID).Order("`index` ASC").Find(ctx)
}

func (db *Database) UpdateChapterSceneIDs(ctx context.Context, chapterID string, sceneIDs []string) error {
	// GORM 使用 JSON tag 会自动序列化 []string
	chapter := Chapter{
		SceneIDs:  sceneIDs,
		UpdatedAt: time.Now(),
	}
	result := db.db.WithContext(ctx).Model(&Chapter{}).Where("id = ?", chapterID).Updates(chapter)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ===== Scene DAO =====

func (db *Database) CreateScenes(ctx context.Context, scenes []Scene) error {
	if len(scenes) == 0 {
		return nil
	}
	return gorm.G[Scene](db.db).CreateInBatches(ctx, &scenes, batchSize)
}

func (db *Database) GetScene(ctx context.Context, id string) (Scene, error) {
	return gorm.G[Scene](db.db).Where("id = ?", id).Take(ctx)
}

func (db *Database) ListScenesByChapter(ctx context.Context, chapterID string) ([]Scene, error) {
	return gorm.G[Scene](db.db).Where("chapter_id = ?", chapterID).Order("`index` ASC").Find(ctx)
}

func (db *Database) ListScenesByDocument(ctx context.Context, documentID string) ([]Scene, error) {
	return gorm.G[Scene](db.db).Where("document_id = ?", documentID).Order("`index` ASC").Find(ctx)
}

func (db *Database) ListPendingImageScenes(ctx context.Context, documentID string) ([]Scene, error) {
	return gorm.G[Scene](db.db).Where("document_id = ? AND (image_url = ? OR image_url IS NULL)", documentID, "").Order("`index` ASC").Find(ctx)
}

func (db *Database) UpdateSceneImageURL(ctx context.Context, sceneID string, imageURL string) error {
	result := db.db.WithContext(ctx).Model(&Scene{}).Where("id = ?", sceneID).Updates(map[string]interface{}{
		"image_url":  imageURL,
		"updated_at": time.Now(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (db *Database) UpdateSceneVoiceURL(ctx context.Context, sceneID string, voiceURL string) error {
	result := db.db.WithContext(ctx).Model(&Scene{}).Where("id = ?", sceneID).Updates(map[string]interface{}{
		"voice_url":  voiceURL,
		"updated_at": time.Now(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (db *Database) DeleteScenesByChapter(ctx context.Context, chapterID string) error {
	_, err := gorm.G[Scene](db.db).Where("chapter_id = ?", chapterID).Delete(ctx)
	return err
}

func (db *Database) DeleteScenesByDocument(ctx context.Context, documentID string) error {
	_, err := gorm.G[Scene](db.db).Where("document_id = ?", documentID).Delete(ctx)
	return err
}

// ===== Role DAO =====

func (db *Database) CreateRoles(ctx context.Context, roles []Role) error {
	if len(roles) == 0 {
		return nil
	}
	return gorm.G[Role](db.db).CreateInBatches(ctx, &roles, batchSize)
}

func (db *Database) GetRole(ctx context.Context, id string) (Role, error) {
	return gorm.G[Role](db.db).Where("id = ?", id).Take(ctx)
}

func (db *Database) ListRolesByDocument(ctx context.Context, documentID string) ([]Role, error) {
	return gorm.G[Role](db.db).Where("document_id = ?", documentID).Order("created_at ASC").Find(ctx)
}

func (db *Database) DeleteRolesByDocument(ctx context.Context, documentID string) error {
	_, err := gorm.G[Role](db.db).Where("document_id = ?", documentID).Delete(ctx)
	return err
}
