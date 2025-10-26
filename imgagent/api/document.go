package api

type CreateDocumentArgs struct {
	Name string `json:"name" binding:"required,max=50"`
}

type UpdateDocumentArgs struct {
	Name string `json:"name" binding:"required,max=50"`
}

type Document struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	FileID          string `json:"file_id"`
	SummaryImageURL string `json:"summary_image_url"`
	Status          string `json:"status"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type ListDocumentsResult struct {
	Documents []Document `json:"documents"`
}

type Chapter struct {
	ID         string   `json:"id"`
	Index      int      `json:"index"`
	DocumentID string   `json:"document_id"`
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	SceneIDs   []string `json:"scene_ids"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
}

type UpdateChapterArgs struct {
	Content string `json:"content" binding:"required,max=4000"`
}

type ListChaptersResult struct {
	Chapters []Chapter `json:"chapters"`
}

type Records struct {
	ID      string  `json:"id"`
	Content string  `json:"content"`
	Score   float32 `json:"score"`
}
