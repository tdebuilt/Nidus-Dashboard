package models

import "time"

// Category represents a dashboard category for organizing widgets.
type Category struct {
	ID        int64     `json:"id" example:"1"`
	Name      string    `json:"name" example:"Home"`
	Slug      string    `json:"slug" example:"home"`
	Icon      string    `json:"icon" example:"home"`
	SortOrder int       `json:"sort_order" example:"0"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateCategoryRequest is the payload for POST /api/categories.
type CreateCategoryRequest struct {
	Name string `json:"name" example:"Media"`
	Icon string `json:"icon" example:"tv"`
}

// UpdateCategoryRequest is the payload for PUT /api/categories/{id}.
type UpdateCategoryRequest struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

// ReorderRequest is the payload for PUT /api/categories/reorder.
type ReorderRequest struct {
	IDs []int64 `json:"ids"`
}

// Widget represents a dashboard widget within a category.
type Widget struct {
	ID         int64     `json:"id" example:"1"`
	CategoryID int64     `json:"category_id" example:"1"`
	Type       string    `json:"type" example:"docker"`
	Title      string    `json:"title" example:"My Docker"`
	Config     string    `json:"config" example:"{}"`
	Collapsed  bool      `json:"collapsed" example:"false"`
	PosX       int       `json:"pos_x" example:"0"`
	PosY       int       `json:"pos_y" example:"0"`
	Width      int       `json:"width" example:"4"`
	Height     int       `json:"height" example:"0"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ToggleCollapseRequest is the payload for PATCH /api/widgets/{id}/toggle-collapse.
type ToggleCollapseRequest struct {
	Collapsed bool `json:"collapsed"`
}

// CreateWidgetRequest is the payload for POST /api/categories/{id}/widgets.
type CreateWidgetRequest struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Config string `json:"config"`
	PosX   int    `json:"pos_x"`
	PosY   int    `json:"pos_y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// UpdateWidgetRequest is the payload for PUT /api/widgets/{id}.
type UpdateWidgetRequest struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Config string `json:"config"`
}

// WidgetLayout represents one widget's position and size for layout saving.
type WidgetLayout struct {
	ID     int64 `json:"id"`
	PosX   int   `json:"pos_x"`
	PosY   int   `json:"pos_y"`
	Width  int   `json:"width"`
	Height int   `json:"height"`
}

// SaveLayoutRequest is the payload for PUT /api/widgets/layout.
type SaveLayoutRequest struct {
	Widgets []WidgetLayout `json:"widgets"`
}
