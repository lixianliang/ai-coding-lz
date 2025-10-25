package db

import (
	"context"

	"imgagent/api"
)

type IDataBase interface {
	UserToken(ctx context.Context, token string) (UserToken, error)
	User(ctx context.Context, uid int64) (User, error)
	GetAdminID(ctx context.Context) (int64, error)

	// Document
	CreateDocument(ctx context.Context, docID string, args *api.CreateDocumentArgs) (*Document, error)
	GetDocument(ctx context.Context, id string) (Document, error)
	GetDocumentWithName(ctx context.Context, name string) (Document, error)
	UpdateDocument(ctx context.Context, id string, args *api.UpdateDocumentArgs) error
	UpdateDocumentStatus(ctx context.Context, id string, status string) error
	UpdateDocumentFileID(ctx context.Context, id string, fileID string) error
	DeleteDocument(ctx context.Context, id string) error
	ListDocuments(ctx context.Context) ([]Document, error)
	ListChapterReadyDocuments(ctx context.Context) ([]Document, error)
	ListSceneReadyDocuments(ctx context.Context) ([]Document, error)

	// Chapter
	CreateChapters(ctx context.Context, documentID string, texts []string) error
	GetChapter(ctx context.Context, id, documentID string) (Chapter, error)
	UpdateChapter(ctx context.Context, id string, args *api.UpdateChapterArgs) error
	UpdateChapterSceneIDs(ctx context.Context, chapterID string, sceneIDs []string) error
	DeleteChapter(ctx context.Context, id, documentID string) error
	DeleteAllChapter(ctx context.Context, documentID string) error
	ListChapters(ctx context.Context, documentID string) ([]Chapter, error)

	// Scene
	CreateScenes(ctx context.Context, scenes []Scene) error
	GetScene(ctx context.Context, id string) (Scene, error)
	ListScenesByChapter(ctx context.Context, chapterID string) ([]Scene, error)
	ListScenesByDocument(ctx context.Context, documentID string) ([]Scene, error)
	ListPendingImageScenes(ctx context.Context, documentID string) ([]Scene, error)
	UpdateSceneImageURL(ctx context.Context, sceneID string, imageURL string) error
	DeleteScenesByChapter(ctx context.Context, chapterID string) error
	DeleteScenesByDocument(ctx context.Context, documentID string) error

	// Role
	CreateRoles(ctx context.Context, roles []Role) error
	GetRole(ctx context.Context, id string) (Role, error)
	ListRolesByDocument(ctx context.Context, documentID string) ([]Role, error)
	DeleteRolesByDocument(ctx context.Context, documentID string) error
}
