package api

type CreateDocumentArgs struct {
	Name string `json:"name" binding:"required,max=50"`
	URL  string `json:"url" binding:"required,max=200"`
}

type UpdateDocumentArgs struct {
	Name string `json:"name" binding:"required,max=50"`
}

type Document struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ListDocumentsResult struct {
	Documents []Document `json:"documents"`
}

type Chapter struct {
	ID         string `json:"id"`
	Index      int    `json:"index"`
	DocumentID string `json:"document_id"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type UpdateChapterArgs struct {
	Content string `json:"content" binding:"required,max=4000"`
}

type ListChaptersResult struct {
	Chapters []Chapter `json:"Chapters"`
}

type Records struct {
	ID      string  `json:"id"`
	Content string  `json:"content"`
	Score   float32 `json:"score"`
}
