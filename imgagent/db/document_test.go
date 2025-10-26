package db

import (
	"context"
	"testing"

	"imgagent/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) *Database {
	// 使用内存数据库（每次测试使用独立的数据库）
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)

	// AutoMigrate (SQLite 不需要表选项)
	err = db.AutoMigrate(&Document{}, &Chapter{}, &Scene{}, &Role{})
	require.NoError(t, err)

	return &Database{db: db}
}

func TestCreateDocument(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// 创建文档
	docID := MakeUUID()
	args := &api.CreateDocumentArgs{
		Name: "测试文档",
	}

	doc, err := db.CreateDocument(ctx, docID, "file-id-test", args)
	require.NoError(t, err)
	assert.Equal(t, docID, doc.ID)
	assert.Equal(t, "测试文档", doc.Name)
	assert.Equal(t, DocumentStatusChapterReady, doc.Status)
	assert.Equal(t, "file-id-test", doc.FileID)

	// 验证可以查询
	found, err := db.GetDocument(ctx, docID)
	require.NoError(t, err)
	assert.Equal(t, doc.Name, found.Name)
}

func TestCreateChapters(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	docID := MakeUUID()
	texts := []string{"第一章内容", "第二章内容", "第三章内容"}

	err := db.CreateChapters(ctx, docID, texts)
	require.NoError(t, err)

	// 查询章节
	chapters, err := db.ListChapters(ctx, docID)
	require.NoError(t, err)
	assert.Equal(t, 3, len(chapters))
	assert.Equal(t, 0, chapters[0].Index)
	assert.Equal(t, 1, chapters[1].Index)
	assert.Equal(t, 2, chapters[2].Index)
}

func TestCreateRoles(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	docID := MakeUUID()
	roles := []Role{
		{
			ID:         MakeUUID(),
			DocumentID: docID,
			Name:       "张三",
			Gender:     "男",
			Character:  "勇敢",
			Appearance: "高大",
		},
		{
			ID:         MakeUUID(),
			DocumentID: docID,
			Name:       "李四",
			Gender:     "女",
			Character:  "聪明",
			Appearance: "美丽",
		},
	}

	err := db.CreateRoles(ctx, roles)
	require.NoError(t, err)

	// 查询角色
	foundRoles, err := db.ListRolesByDocument(ctx, docID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(foundRoles))
}

func TestCreateScenes(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	docID := MakeUUID()
	chapterID := MakeUUID()

	scenes := []Scene{
		{
			ID:         MakeUUID(),
			ChapterID:  chapterID,
			DocumentID: docID,
			Index:      0,
			Content:    "场景1描述",
		},
		{
			ID:         MakeUUID(),
			ChapterID:  chapterID,
			DocumentID: docID,
			Index:      1,
			Content:    "场景2描述",
		},
	}

	err := db.CreateScenes(ctx, scenes)
	require.NoError(t, err)

	// 查询场景
	foundScenes, err := db.ListScenesByChapter(ctx, chapterID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(foundScenes))
}

func TestUpdateChapterSceneIDs(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	docID := MakeUUID()

	// 先创建章节
	err := db.CreateChapters(ctx, docID, []string{"测试内容"})
	require.NoError(t, err)

	chapters, err := db.ListChapters(ctx, docID)
	require.NoError(t, err)
	chapterID := chapters[0].ID

	// 更新场景ID列表
	sceneIDs := []string{"scene1", "scene2", "scene3"}
	err = db.UpdateChapterSceneIDs(ctx, chapterID, sceneIDs)
	require.NoError(t, err)

	// 验证
	updated, err := db.GetChapter(ctx, chapterID, docID)
	require.NoError(t, err)
	assert.Equal(t, 3, len(updated.SceneIDs))
}

func TestListChapterReadyDocuments(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// 创建不同状态的文档
	doc1 := MakeUUID()
	doc2 := MakeUUID()
	doc3 := MakeUUID()

	db.CreateDocument(ctx, doc1, "file-id-1", &api.CreateDocumentArgs{Name: "doc1"})
	db.CreateDocument(ctx, doc2, "file-id-2", &api.CreateDocumentArgs{Name: "doc2"})
	db.CreateDocument(ctx, doc3, "file-id-3", &api.CreateDocumentArgs{Name: "doc3"})

	// 更新 doc2 状态为 sceneReady
	db.UpdateDocumentStatus(ctx, doc2, DocumentStatusSceneReady)

	// 查询 chapterReady 的文档
	docs, err := db.ListChapterReadyDocuments(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(docs)) // doc1 和 doc3
}

func TestUpdateSceneImageURL(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	docID := MakeUUID()
	chapterID := MakeUUID()

	// 创建场景
	scenes := []Scene{
		{
			ID:         MakeUUID(),
			ChapterID:  chapterID,
			DocumentID: docID,
			Index:      0,
			Content:    "测试场景",
		},
	}
	err := db.CreateScenes(ctx, scenes)
	require.NoError(t, err)

	// 更新图片URL
	imageURL := "https://example.com/image.png"
	err = db.UpdateSceneImageURL(ctx, scenes[0].ID, imageURL)
	require.NoError(t, err)

	// 验证
	scene, err := db.GetScene(ctx, scenes[0].ID)
	require.NoError(t, err)
	assert.Equal(t, imageURL, scene.ImageURL)
}

func TestListPendingImageScenes(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	docID := MakeUUID()
	chapterID := MakeUUID()

	// 创建场景
	scenes := []Scene{
		{
			ID:         MakeUUID(),
			ChapterID:  chapterID,
			DocumentID: docID,
			Index:      0,
			Content:    "场景1",
			ImageURL:   "", // 没有图片
		},
		{
			ID:         MakeUUID(),
			ChapterID:  chapterID,
			DocumentID: docID,
			Index:      1,
			Content:    "场景2",
			ImageURL:   "https://example.com/img.png", // 有图片
		},
	}
	err := db.CreateScenes(ctx, scenes)
	require.NoError(t, err)

	// 查询待生成图片的场景
	pendingScenes, err := db.ListPendingImageScenes(ctx, docID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(pendingScenes)) // 只有场景1需要生成图片
}

func TestListScenesByDocument(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	docID := MakeUUID()
	chapterID1 := MakeUUID()
	chapterID2 := MakeUUID()

	// 创建不同章节的场景
	scenes := []Scene{
		{ID: MakeUUID(), ChapterID: chapterID1, DocumentID: docID, Index: 0, Content: "场景1"},
		{ID: MakeUUID(), ChapterID: chapterID1, DocumentID: docID, Index: 1, Content: "场景2"},
		{ID: MakeUUID(), ChapterID: chapterID2, DocumentID: docID, Index: 0, Content: "场景3"},
	}
	err := db.CreateScenes(ctx, scenes)
	require.NoError(t, err)

	// 查询文档的所有场景
	allScenes, err := db.ListScenesByDocument(ctx, docID)
	require.NoError(t, err)
	assert.Equal(t, 3, len(allScenes))
}

func TestUpdateDocumentFileID(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// 创建文档
	docID := MakeUUID()
	args := &api.CreateDocumentArgs{Name: "测试文档"}
	_, err := db.CreateDocument(ctx, docID, "file-id-init", args)
	require.NoError(t, err)

	// 更新 FileID
	fileID := "file-123456"
	err = db.UpdateDocumentFileID(ctx, docID, fileID)
	require.NoError(t, err)

	// 验证
	doc, err := db.GetDocument(ctx, docID)
	require.NoError(t, err)
	assert.Equal(t, fileID, doc.FileID)
}

func TestListRoleReadyDocuments(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// 创建不同状态的文档
	doc1 := MakeUUID()
	doc2 := MakeUUID()
	doc3 := MakeUUID()
	doc4 := MakeUUID()

	db.CreateDocument(ctx, doc1, "file-id-1", &api.CreateDocumentArgs{Name: "doc1"})
	db.CreateDocument(ctx, doc2, "file-id-2", &api.CreateDocumentArgs{Name: "doc2"})
	db.CreateDocument(ctx, doc3, "file-id-3", &api.CreateDocumentArgs{Name: "doc3"})
	db.CreateDocument(ctx, doc4, "file-id-4", &api.CreateDocumentArgs{Name: "doc4"})

	// 设置状态
	db.UpdateDocumentStatus(ctx, doc1, DocumentStatusChapterReady)
	db.UpdateDocumentStatus(ctx, doc2, DocumentStatusRoleReady)
	db.UpdateDocumentStatus(ctx, doc3, DocumentStatusSceneReady)
	db.UpdateDocumentStatus(ctx, doc4, DocumentStatusImgReady)

	// 查询 roleReady 的文档
	docs, err := db.ListRoleReadyDocuments(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, len(docs))
	assert.Equal(t, doc2, docs[0].ID)
}

func TestListSceneReadyDocuments(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// 创建不同状态的文档
	doc1 := MakeUUID()
	doc2 := MakeUUID()
	doc3 := MakeUUID()

	db.CreateDocument(ctx, doc1, "file-id-1", &api.CreateDocumentArgs{Name: "doc1"})
	db.CreateDocument(ctx, doc2, "file-id-2", &api.CreateDocumentArgs{Name: "doc2"})
	db.CreateDocument(ctx, doc3, "file-id-3", &api.CreateDocumentArgs{Name: "doc3"})

	// 设置状态
	db.UpdateDocumentStatus(ctx, doc1, DocumentStatusChapterReady)
	db.UpdateDocumentStatus(ctx, doc2, DocumentStatusSceneReady)
	db.UpdateDocumentStatus(ctx, doc3, DocumentStatusImgReady)

	// 查询 sceneReady 的文档
	docs, err := db.ListSceneReadyDocuments(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, len(docs))
	assert.Equal(t, doc2, docs[0].ID)
}

func TestDeleteRolesByDocument(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	docID := MakeUUID()

	// 创建角色
	roles := []Role{
		{ID: MakeUUID(), DocumentID: docID, Name: "角色1"},
		{ID: MakeUUID(), DocumentID: docID, Name: "角色2"},
	}
	err := db.CreateRoles(ctx, roles)
	require.NoError(t, err)

	// 删除角色的角色
	err = db.DeleteRolesByDocument(ctx, docID)
	require.NoError(t, err)

	// 验证
	foundRoles, err := db.ListRolesByDocument(ctx, docID)
	require.NoError(t, err)
	assert.Equal(t, 0, len(foundRoles))
}

func TestDeleteScenesByDocument(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	docID := MakeUUID()
	chapterID := MakeUUID()

	// 创建场景
	scenes := []Scene{
		{ID: MakeUUID(), ChapterID: chapterID, DocumentID: docID, Index: 0, Content: "场景1"},
		{ID: MakeUUID(), ChapterID: chapterID, DocumentID: docID, Index: 1, Content: "场景2"},
	}
	err := db.CreateScenes(ctx, scenes)
	require.NoError(t, err)

	// 删除文档的场景
	err = db.DeleteScenesByDocument(ctx, docID)
	require.NoError(t, err)

	// 验证
	foundScenes, err := db.ListScenesByDocument(ctx, docID)
	require.NoError(t, err)
	assert.Equal(t, 0, len(foundScenes))
}

func TestFullFlow(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// 1. 创建文档
	docID := MakeUUID()
	args := &api.CreateDocumentArgs{Name: "完整流程测试"}
	doc, err := db.CreateDocument(ctx, docID, "file-id-full", args)
	require.NoError(t, err)
	assert.Equal(t, DocumentStatusChapterReady, doc.Status)

	// 2. 创建章节
	err = db.CreateChapters(ctx, docID, []string{"第一章", "第二章"})
	require.NoError(t, err)

	chapters, err := db.ListChapters(ctx, docID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(chapters))

	// 3. 创建角色
	roles := []Role{
		{ID: MakeUUID(), DocumentID: docID, Name: "主角", Gender: "男", Character: "勇敢"},
	}
	err = db.CreateRoles(ctx, roles)
	require.NoError(t, err)

	foundRoles, err := db.ListRolesByDocument(ctx, docID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(foundRoles))

	// 4. 创建场景
	scenes := []Scene{
		{ID: MakeUUID(), ChapterID: chapters[0].ID, DocumentID: docID, Index: 0, Content: "场景描述"},
	}
	err = db.CreateScenes(ctx, scenes)
	require.NoError(t, err)

	// 5. 更新章节的场景IDs
	sceneIDs := []string{scenes[0].ID}
	err = db.UpdateChapterSceneIDs(ctx, chapters[0].ID, sceneIDs)
	require.NoError(t, err)

	// 6. 更新状态
	err = db.UpdateDocumentStatus(ctx, docID, DocumentStatusSceneReady)
	require.NoError(t, err)

	// 7. 生成图片
	err = db.UpdateSceneImageURL(ctx, scenes[0].ID, "https://example.com/img.png")
	require.NoError(t, err)

	err = db.UpdateDocumentStatus(ctx, docID, DocumentStatusImgReady)
	require.NoError(t, err)

	// 验证最终状态
	finalDoc, err := db.GetDocument(ctx, docID)
	require.NoError(t, err)
	assert.Equal(t, DocumentStatusImgReady, finalDoc.Status)
}
