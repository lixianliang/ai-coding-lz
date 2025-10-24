package db

import (
	"context"

	"imgagent/api"
)

type IDataBase interface {
	UserToken(ctx context.Context, token string) (UserToken, error)
	User(ctx context.Context, uid int64) (User, error)
	GetAdminID(ctx context.Context) (int64, error)

	CreateDocument(ctx context.Context, datasetID, docID string, args *api.CreateDocumentArgs) (*Document, error)
	GetDocument(ctx context.Context, id string) (Document, error)
	GetDocumentWithName(ctx context.Context, datasetID string, name string) (Document, error)
	UpdateDocument(ctx context.Context, id string, args *api.UpdateDocumentArgs) error
	UpdateDocumentStatus(ctx context.Context, id string, status string) error
	DeleteDocument(ctx context.Context, id string) error
	ListDocuments(ctx context.Context, datasetID string) ([]Document, error)
	CountDocument(ctx context.Context, datasetID string) (int64, error)
	ListIndexingDocuments(ctx context.Context) ([]Document, error)

	CreateChapters(ctx context.Context, documentID string, texts []string) error
	GetChapter(ctx context.Context, id, documentID string) (Chapter, error)
	UpdateChapter(ctx context.Context, id string, args *api.UpdateChapterArgs) error
	DeleteChapter(ctx context.Context, id, documentID string) error
	DeleteAllChapter(ctx context.Context, documentID string) error
	ListChapters(ctx context.Context, documentID string) ([]Chapter, error)
}
