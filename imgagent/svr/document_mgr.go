package svr

import (
	"context"
	"fmt"
	"time"

	"imgagent/bailian"
	"imgagent/db"
	"imgagent/pkg/logger"
)

type DocumentConfigEx struct {
	config DocumentConfig

	db db.IDataBase
}

type DocumentConfig struct {
	Enable                     bool `json:"enable"`
	HandleRoleIntervalSecs     int  `json:"handle_role_interval_secs"`
	HandleSceneIntervalSecs    int  `json:"handle_scene_interval_secs"`
	HandleImageGenIntervalSecs int  `json:"handle_image_gen_interval_secs"`
}

type DocumentMgr struct {
	DocumentConfigEx

	close         chan bool
	db            db.IDataBase
	bailianClient *bailian.Client
}

func newDocumentMgr(confEx DocumentConfigEx, bailianClient *bailian.Client) (*DocumentMgr, error) {
	// 设置默认值
	if confEx.config.HandleRoleIntervalSecs == 0 {
		confEx.config.HandleRoleIntervalSecs = 30
	}
	if confEx.config.HandleSceneIntervalSecs == 0 {
		confEx.config.HandleSceneIntervalSecs = 30
	}
	if confEx.config.HandleImageGenIntervalSecs == 0 {
		confEx.config.HandleImageGenIntervalSecs = 30
	}

	return &DocumentMgr{
		DocumentConfigEx: confEx,
		db:               confEx.db,
		bailianClient:    bailianClient,
		close:            make(chan bool),
	}, nil
}

func (m *DocumentMgr) Run() {
	go m.loopHandleDocumentRoleTasks()
	go m.loopHandleDocumentScenceTasks()
	go m.loopHandleImageGenTasks()
}

func (m *DocumentMgr) loopHandleDocumentRoleTasks() {
	ticker := time.NewTicker(time.Second * time.Duration(m.config.HandleRoleIntervalSecs))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := logger.NewContext(fmt.Sprintf("HandleDocumentRoleTasks-%d", time.Now().Unix()))
			m.HandleDocumentRoleTasks(ctx)
		case <-m.close:
			return
		}
	}
}

func (m *DocumentMgr) loopHandleDocumentScenceTasks() {
	ticker := time.NewTicker(time.Second * time.Duration(m.config.HandleSceneIntervalSecs))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := logger.NewContext(fmt.Sprintf("HandleDocumentScenceTasks-%d", time.Now().Unix()))
			m.HandleDocumentScenceTasks(ctx)
		case <-m.close:
			return
		}
	}
}

func (m *DocumentMgr) loopHandleImageGenTasks() {
	ticker := time.NewTicker(time.Second * time.Duration(m.config.HandleImageGenIntervalSecs))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := logger.NewContext(fmt.Sprintf("HandleImageGenTasks-%d", time.Now().Unix()))
			m.HandleImageGenTasks(ctx)
		case <-m.close:
			return
		}
	}
}

func (m *DocumentMgr) HandleDocumentRoleTasks(ctx context.Context) {
	log := logger.FromContext(ctx)

	docs, err := m.db.ListChapterReadyDocuments(ctx)
	if err != nil {
		log.Errorf("Failed to list chapterReady documents, err: %v", err)
		return
	}

	for _, doc := range docs {
		err = m.HandleDocumentRole(ctx, doc)
		if err != nil {
			log.Errorf("Failed to handle document role, doc: %v, err: %v", doc, err)
			continue
		}
		err = m.db.UpdateDocumentStatus(ctx, doc.ID, db.DocumentStatusRoleReady)
		if err != nil {
			log.Errorf("Failed to update document status, err: %v", err)
			continue
		}
	}
}

func (m *DocumentMgr) HandleDocumentRole(ctx context.Context, doc db.Document) error {
	log := logger.FromContext(ctx)
	log.Infof("Handling document role extraction, docID: %s", doc.ID)

	// 1. 先提取摘要
	if doc.Summary == "" {
		log.Infof("Extracting summary, docID: %s", doc.ID)
		summary, err := m.bailianClient.ExtractSummary(ctx, doc.FileID)
		if err != nil {
			log.Errorf("Failed to extract summary, doc: %s, err: %v", doc.ID, err)
			return err
		}

		if summary == "" {
			log.Warnf("Empty summary extracted for doc: %s", doc.ID)
		}

		err = m.db.UpdateDocumentSummary(ctx, doc.ID, summary)
		if err != nil {
			log.Errorf("Failed to update document summary, doc: %s, err: %v", doc.ID, err)
			return err
		}

		doc.Summary = summary
		log.Infof("Summary extracted and saved for doc: %s, length: %d", doc.ID, len(summary))

		// 生成封面图片
		if summary != "" {
			log.Infof("Generating cover image for doc: %s", doc.ID)
			coverImageURL, err := m.bailianClient.GenerateCoverImage(ctx, summary)
			if err != nil {
				log.Errorf("Failed to generate cover image, doc: %s, err: %v", doc.ID, err)
				// 封面生成失败不影响后续流程，记录日志后继续
			} else {
				err = m.db.UpdateDocumentSummaryImageURL(ctx, doc.ID, coverImageURL)
				if err != nil {
					log.Errorf("Failed to update document summary image URL, doc: %s, err: %v", doc.ID, err)
					// 更新失败不影响后续流程
				} else {
					log.Infof("Cover image generated and saved for doc: %s, URL: %s", doc.ID, coverImageURL)
				}
			}
		}
	} else {
		log.Infof("Summary already exists for doc: %s", doc.ID)
	}

	// 2. 检查是否已有角色
	existingRoles, err := m.db.ListRolesByDocument(ctx, doc.ID)
	if err != nil {
		log.Errorf("Failed to list existing roles, doc: %s, err: %v", doc.ID, err)
		return err
	}

	if len(existingRoles) > 0 {
		log.Infof("Roles already exist for doc: %s, count: %d", doc.ID, len(existingRoles))
		return nil
	}

	// 3. 提取角色（传入摘要以获得更好的结果）
	log.Infof("Extracting roles, docID: %s", doc.ID)
	roles, err := m.bailianClient.ExtractRoles(ctx, doc.FileID, doc.Summary)
	if err != nil {
		log.Errorf("Failed to extract roles, doc: %s, err: %v", doc.ID, err)
		return err
	}

	// 角色不允许为空
	if len(roles) == 0 {
		log.Errorf("No roles extracted for doc: %s", doc.ID)
		return fmt.Errorf("no roles extracted")
	}

	// 4. 保存角色到数据库
	dbRoles := make([]db.Role, 0, len(roles))
	now := time.Now()
	for _, r := range roles {
		dbRoles = append(dbRoles, db.Role{
			ID:         db.MakeUUID(),
			DocumentID: doc.ID,
			Name:       r.Name,
			Gender:     r.Gender,
			Character:  r.Character,
			Appearance: r.Appearance,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}

	err = m.db.CreateRoles(ctx, dbRoles)
	if err != nil {
		log.Errorf("Failed to create roles, doc: %s, err: %v", doc.ID, err)
		return err
	}

	log.Infof("Created %d roles for doc: %s", len(dbRoles), doc.ID)
	return nil
}

func (m *DocumentMgr) HandleDocumentScenceTasks(ctx context.Context) {
	log := logger.FromContext(ctx)

	docs, err := m.db.ListRoleReadyDocuments(ctx)
	if err != nil {
		log.Errorf("Failed to list roleReady documents, err: %v", err)
		return
	}

	for _, doc := range docs {
		err = m.HandleDocumentScence(ctx, doc)
		if err != nil {
			log.Errorf("Failed to handle document scene, doc: %v, err: %v", doc, err)
			continue
		}
		err = m.db.UpdateDocumentStatus(ctx, doc.ID, db.DocumentStatusSceneReady)
		if err != nil {
			log.Errorf("Failed to update document status, err: %v", err)
			continue
		}
	}
}

func (m *DocumentMgr) HandleDocumentScence(ctx context.Context, doc db.Document) error {
	log := logger.FromContext(ctx)
	log.Infof("Handling document scene extraction, docID: %s", doc.ID)

	// 1. 获取所有章节
	chapters, err := m.db.ListChapters(ctx, doc.ID)
	if err != nil {
		log.Errorf("Failed to list chapters, doc: %s, err: %v", doc.ID, err)
		return err
	}

	if len(chapters) == 0 {
		log.Warnf("No chapters found for doc: %s", doc.ID)
		return nil
	}

	// 2. 为每个章节生成场景
	sceneIndex := 0
	for _, chapter := range chapters {
		log.Infof("Generating scenes for chapter, chapterID: %s, index: %d", chapter.ID, chapter.Index)

		scenes, err := m.bailianClient.GenerateScenes(ctx, chapter.Content)
		if err != nil {
			log.Errorf("Failed to generate scenes, chapter: %s, err: %v", chapter.ID, err)
			return err
		}

		log.Infof("Generated %d scenes for chapter: %s", len(scenes), chapter.ID)

		// 保存场景到数据库
		if len(scenes) > 0 {
			dbScenes := make([]db.Scene, 0, len(scenes))
			sceneIDs := make([]string, 0, len(scenes))
			now := time.Now()

			for _, sceneContent := range scenes {
				sceneID := db.MakeUUID()
				sceneIDs = append(sceneIDs, sceneID)
				dbScenes = append(dbScenes, db.Scene{
					ID:         sceneID,
					ChapterID:  chapter.ID,
					DocumentID: doc.ID,
					Index:      sceneIndex,
					Content:    sceneContent,
					CreatedAt:  now,
					UpdatedAt:  now,
				})
				sceneIndex++
			}

			err = m.db.CreateScenes(ctx, dbScenes)
			if err != nil {
				log.Errorf("Failed to create scenes, chapter: %s, err: %v", chapter.ID, err)
				return err
			}

			// 更新 Chapter 的 SceneIDs
			err = m.db.UpdateChapterSceneIDs(ctx, chapter.ID, sceneIDs)
			if err != nil {
				log.Errorf("Failed to update chapter sceneIDs, chapter: %s, err: %v", chapter.ID, err)
				return err
			}
		}
	}

	log.Infof("Scene extraction completed for doc: %s", doc.ID)
	return nil
}

// HandleImageGenTasks 处理图片生成任务
func (m *DocumentMgr) HandleImageGenTasks(ctx context.Context) {
	log := logger.FromContext(ctx)

	// 查询 sceneReady 状态的文档
	docs, err := m.db.ListSceneReadyDocuments(ctx)
	if err != nil {
		log.Errorf("Failed to list sceneReady documents, err: %v", err)
		return
	}

	// 逐个处理文档
	for _, doc := range docs {
		err = m.HandleDocumentImageGen(ctx, doc)
		if err != nil {
			log.Errorf("Failed to handle document image gen, doc: %s, err: %v", doc.ID, err)
			continue // 失败保持状态，下次继续处理
		}

		// 更新文档状态为 imgReady
		err = m.db.UpdateDocumentStatus(ctx, doc.ID, db.DocumentStatusImgReady)
		if err != nil {
			log.Errorf("Failed to update document status, doc: %s, err: %v", doc.ID, err)
			continue
		}

		log.Infof("Image generation completed for doc: %s", doc.ID)
	}
}

// HandleDocumentImageGen 处理单个文档的图片生成
func (m *DocumentMgr) HandleDocumentImageGen(ctx context.Context, doc db.Document) error {
	log := logger.FromContext(ctx)
	log.Infof("Handling document image generation, docID: %s", doc.ID)

	// 1. 获取文档的角色信息
	dbRoles, err := m.db.ListRolesByDocument(ctx, doc.ID)
	if err != nil {
		log.Errorf("Failed to list roles, doc: %s, err: %v", doc.ID, err)
		return err
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

	// 2. 获取所有未生成图片的场景
	scenes, err := m.db.ListPendingImageScenes(ctx, doc.ID)
	if err != nil {
		log.Errorf("Failed to list pending image scenes, doc: %s, err: %v", doc.ID, err)
		return err
	}

	if len(scenes) == 0 {
		log.Infof("No pending image scenes for doc: %s", doc.ID)
		return nil
	}

	log.Infof("Found %d pending image scenes for doc: %s", len(scenes), doc.ID)

	// 3. 为每个场景生成图片和语音（包含摘要和角色信息）
	for _, scene := range scenes {
		log.Infof("Generating image and voice for scene, sceneID: %s, content: %s", scene.ID, scene.Content)

		imageURL, err := m.bailianClient.GenerateImage(ctx, scene.Content, doc.Summary, roles)
		if err != nil {
			log.Errorf("Failed to generate image, scene: %s, err: %v", scene.ID, err)
			return err // 失败则整个文档重试
		}

		// 更新场景图片 URL
		err = m.db.UpdateSceneImageURL(ctx, scene.ID, imageURL)
		if err != nil {
			log.Errorf("Failed to update scene imageURL, scene: %s, err: %v", scene.ID, err)
			return err
		}

		log.Infof("Image generated for scene: %s, URL: %s", scene.ID, imageURL)

		// 生成语音
		voiceURL, err := m.bailianClient.GenerateTTS(ctx, scene.Content)
		if err != nil {
			log.Errorf("Failed to generate TTS, scene: %s, err: %v", scene.ID, err)
			return err
		}

		// 更新场景语音 URL
		err = m.db.UpdateSceneVoiceURL(ctx, scene.ID, voiceURL)
		if err != nil {
			log.Errorf("Failed to update scene voiceURL, scene: %s, err: %v", scene.ID, err)
			return err
		}

		log.Infof("Voice generated for scene: %s, URL: %s", scene.ID, voiceURL)
	}

	log.Infof("All images generated for doc: %s", doc.ID)
	return nil
}
