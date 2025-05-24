package handler

import (
	"html/template"

	"github.com/jabuxas/abyss/internal/app"
)

// Handler struct holds dependencies for HTTP handlers, embedding app.Application.
type Handler struct {
	App *app.Application
}

// NewHandler creates a new Handler.
func NewHandler(application *app.Application) *Handler {
	return &Handler{App: application}
}

// FileInfo holds metadata for a single file.
type FileInfo struct {
	Name          string
	Path          string
	Size          int64
	FormattedSize string
	Type          string        // e.g., "image", "video", "text", "pdf"
	Content       template.HTML // For pre-rendered content like syntax highlighted code
	TimeUploaded  string
}

// TemplateData is the data structure passed to HTML templates.
type TemplateData struct {
	Files      []FileInfo
	URL        string // base URL of the application
	PageTitle  string
	ActivePath string
	SingleFile *FileInfo
}
