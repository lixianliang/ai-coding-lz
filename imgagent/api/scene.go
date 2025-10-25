package api

// Role 角色信息
type Role struct {
	ID         string `json:"id"`
	DocumentID string `json:"document_id"`
	Name       string `json:"name"`
	Gender     string `json:"gender"`
	Character  string `json:"character"`
	Appearance string `json:"appearance"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// Scene 场景信息
type Scene struct {
	ID         string `json:"id"`
	ChapterID  string `json:"chapter_id"`
	DocumentID string `json:"document_id"`
	Index      int    `json:"index"`
	Content    string `json:"content"`
	ImageURL   string `json:"image_url"`
	VoiceURL   string `json:"voice_url"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// ListRolesResult 角色列表响应
type ListRolesResult struct {
	Roles []Role `json:"roles"`
}

// ListScenesResult 场景列表响应
type ListScenesResult struct {
	Scenes []Scene `json:"scenes"`
}
