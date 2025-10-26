package svr

import (
	"io"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"imgagent/bailian"
	"imgagent/db"
	"imgagent/pkg/dbutil"
	"imgagent/pkg/middleware"
	"imgagent/storage"
)

type Config struct {
	APIVersion     string         `json:"api_version"`
	Temp           string         `json:"temp"`
	Storage        storage.Config `json:"storage"`
	DB             dbutil.Config  `json:"db"`
	BailianConfig  bailian.Config `json:"-"` // 从外部传入
	DocumentConfig DocumentConfig `json:"-"` // 从外部传入
}

type EmbeddingConfig struct {
	URL    string `json:"url"`
	Model  string `json:"model"`
	APIKey string `json:"api_key"`
}

type Service struct {
	conf          Config
	db            db.IDataBase
	stg           *storage.Storage
	bailianClient *bailian.Client
	documentMgr   *DocumentMgr
}

func New(conf Config, bailianClient *bailian.Client) (*Service, error) {
	if conf.Temp == "" {
		conf.Temp = "./temp"
	}
	err := os.MkdirAll(conf.Temp, 0776)
	if err != nil {
		zap.S().Errorf("Failed to mkdir, err: %v", err)
		return nil, err
	}

	stg, err := storage.NewStorage(conf.Storage)
	if err != nil {
		zap.S().Errorf("Failed to new storage, err: %v", err)
		return nil, err
	}
	db, err := db.NewDatabase(conf.DB)
	if err != nil {
		zap.S().Errorf("Failed to new database, err: %v", err)
		return nil, err
	}

	// 创建文档管理器
	var docMgr *DocumentMgr
	if conf.DocumentConfig.Enable {
		confEx := DocumentConfigEx{
			config: conf.DocumentConfig,
			db:     db,
		}
		var err error
		docMgr, err = newDocumentMgr(confEx, bailianClient)
		if err != nil {
			zap.S().Errorf("Failed to new document manager, err: %v", err)
			return nil, err
		}
		// 启动文档管理器
		docMgr.Run()
		zap.S().Info("Document manager started")
	}

	return &Service{
		conf:          conf,
		db:            db,
		stg:           stg,
		bailianClient: bailianClient,
		documentMgr:   docMgr,
	}, nil
}

func (s *Service) RegisterRouter(writer io.Writer) *gin.Engine {
	router := middleware.NewRouter(writer)
	api := router.Group(s.conf.APIVersion)
	authGroup := api.Group("")
	// 暂不需要 auth
	authGroup.Use(s.NilAuth())

	// Document
	authGroup.POST("/documents", s.HandleCreateDocument)
	authGroup.GET("/documents/:document_id", s.HandleGetDocument)
	authGroup.PUT("/documents/:document_id", s.HandleUpdateDocument)
	authGroup.DELETE("/documents/:document_id", s.HandleDeleteDocument)
	authGroup.GET("/documents", s.HandleListDocuments)

	// Chapter
	authGroup.GET("/documents/:document_id/chapters/:id", s.HandleGetChapter)
	authGroup.PUT("/documents/:document_id/chapters/:id", s.HandleUpdateChapter)
	authGroup.DELETE("/documents/:document_id/chapters/:id", s.HandleDeleteChapter)
	authGroup.GET("/documents/:document_id/chapters", s.HandleListChapters)

	// Role
	authGroup.GET("/documents/:document_id/roles", s.HandleGetRoles)
	authGroup.PUT("/roles/:id", s.HandleUpdateRole)

	// Scene
	authGroup.GET("/documents/:document_id/scenes", s.HandleListScenesByDocument)
	authGroup.GET("/chapters/:chapter_id/scenes", s.HandleListScenesByChapter)
	authGroup.PUT("/scenes/:id", s.HandleUpdateScene)

	return router
}
