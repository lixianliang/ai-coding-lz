package bailian

import "context"

// BailianInterface 定义百炼客户端的接口
type BailianInterface interface {
	UploadFile(ctx context.Context, filename string) (string, error)
	ExtractRoles(ctx context.Context, fileID string) ([]RoleInfo, error)
	GenerateScenes(ctx context.Context, content string) ([]string, error)
	GenerateImage(ctx context.Context, prompt string) (string, error)
}
