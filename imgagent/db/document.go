package db

import (
	"context"
	"time"

	"gorm.io/gorm"

	"imgagent/api"
)

const (
	batchSize = 100

	DocumentStatusInited = "inited"
	DocumentStatusReady  = "ready"
)

// Document 文档表
type Document struct {
	ID        string    `gorm:"primaryKey;size:32;comment:'主键'"`
	Name      string    `gorm:"uniqueIndex:uk_name;size:128;comment:'文档名称'"`
	URL       string    `gorm:"size:255;comment:'存储URL'"`
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
	Index      int       `gorm:"index:uk_document_index,priority:2;comment:'章节序号'"`
	DocumentID string    `gorm:"index:uk_document_index,priority:1;size:32;comment:'文档 id'"`
	Title      string    `gorm:"size:100;comment:'标题'"`
	Content    string    `gorm:"size:10000;comment:'章节内容'"`
	Scene      string    `gorm:"size:1000;comment:'故事场景'"`
	CreatedAt  time.Time `gorm:"comment:'创建时间'"`
	UpdatedAt  time.Time `gorm:"comment:'更新时间'"`
}

func (Chapter) TableName() string {
	return "chapters"
}

// ===== Document DAO =====

func (db *Database) CreateDocument(ctx context.Context, docID string, args *api.CreateDocumentArgs) (*Document, error) {
	now := time.Now()
	doc := Document{
		ID:        docID,
		Name:      args.Name,
		Status:    DocumentStatusInited,
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

/*
func (db *Database) ListIndexingDocuments(ctx context.Context) ([]Document, error) {
	return gorm.G[Document](db.db).Where("status = ?", DocumentStatusInited).Find(ctx)
}*/

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
