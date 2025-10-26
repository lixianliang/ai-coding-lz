package bailian

import "context"

// BailianInterface 定义百炼客户端的接口
type BailianInterface interface {
	UploadFile(ctx context.Context, filename string) (string, error)
	ExtractSummary(ctx context.Context, fileID string) (string, error)
	ExtractRoles(ctx context.Context, fileID string, summary string) ([]RoleInfo, error)
	GenerateScenes(ctx context.Context, content string) ([]string, error)
	GenerateImage(ctx context.Context, prompt string, summary string, roles []RoleInfo) (string, error)
	GenerateCoverImage(ctx context.Context, summary string) (string, error)
	GenerateTTS(ctx context.Context, text string) (string, error)
}
