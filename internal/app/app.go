package app

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/jabuxas/abyss/assets"
	"github.com/jabuxas/abyss/internal/config"
)

type Application struct {
	Config           *config.Config
	Logger           *slog.Logger
	Templates        map[string]*template.Template
	JWTSecret        []byte
	lastUploadedFile string
	ChromaStyle      *chroma.Style
}

func NewApplication(cfg *config.Config) (*Application, error) {
	logger := slog.Default()

	if cfg.AuthUsername == "" {
		return nil, fmt.Errorf("AUTH_USERNAME is required")
	}
	if cfg.AuthPassword == "" {
		return nil, fmt.Errorf("AUTH_PASSWORD is required")
	}
	if cfg.UploadKey == "" {
		logger.Warn("UPLOAD_KEY is not set. API key uploads may be insecure or disabled.")
	}

	if _, err := os.Stat(cfg.FilesDir); os.IsNotExist(err) {
		logger.Info("Creating files directory", "path", cfg.FilesDir)
		if err := os.MkdirAll(cfg.FilesDir, 0750); err != nil {
			return nil, fmt.Errorf("failed to create files directory %s: %w", cfg.FilesDir, err)
		}
	}

	templates, err := loadTemplates(logger, "./assets/templates", assets.TemplatesFS, "templates")
	if err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	chromaStyle, err := loadChromaStyleFromXML(logger, "./assets/templates/colorscheme.xml", assets.TemplatesFS, "templates/colorscheme.xml")
	if err != nil {
		logger.Warn("Failed to load custom chroma style, falling back to default", "error", err)
		chromaStyle = styles.Fallback
	}

	app := &Application{
		Config:      cfg,
		Logger:      logger,
		Templates:   templates,
		JWTSecret:   []byte(cfg.UploadKey),
		ChromaStyle: chromaStyle,
	}

	logger.Info("Application initialized with effective settings",
		"url", cfg.AbyssURL, "filesDir", cfg.FilesDir, "port", cfg.Port,
	)
	return app, nil
}

func loadTemplates(logger *slog.Logger, diskPath string, fullEmbedFS fs.FS, embedDirToWalk string) (map[string]*template.Template, error) {
	templates := make(map[string]*template.Template)

	if _, err := os.Stat(diskPath); err == nil {
		logger.Info("Loading templates from disk", "path", diskPath)
		templateFiles, err := filepath.Glob(filepath.Join(diskPath, "*.html"))
		if err != nil {
			return nil, fmt.Errorf("error globbing disk templates: %w", err)
		}
		if len(templateFiles) == 0 {
			logger.Warn("No HTML files found in disk path", "path", diskPath)
			// if empty, we use embedded
		}
		for _, file := range templateFiles {
			name := filepath.Base(file)
			tmpl, err := template.New(name).ParseFiles(file)
			if err != nil {
				return nil, fmt.Errorf("error parsing disk template %s: %w", name, err)
			}
			templates[name] = tmpl
		}
		if len(templates) > 0 {
			logger.Debug("Loaded templates from disk", "count", len(templates))
			return templates, nil
		}
		logger.Warn("Disk path existed but no templates loaded, trying embedded", "path", diskPath)
	}

	logger.Info("Loading templates from embedded FS", "directory_in_fs", embedDirToWalk)

	templateDirFS, err := fs.Sub(fullEmbedFS, embedDirToWalk)
	if err != nil {
		return nil, fmt.Errorf("failed to create sub-FS for embedded templates at %s: %w", embedDirToWalk, err)
	}

	err = fs.WalkDir(templateDirFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error during walk at %s: %w", path, err)
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".html") {
			name := d.Name()
			tmpl, parseErr := template.New(name).ParseFS(templateDirFS, path)
			if parseErr != nil {
				return fmt.Errorf("parsing embedded template %s (path: %s): %w", name, path, parseErr)
			}
			templates[name] = tmpl
			logger.Debug("Loaded embedded template", "name", name)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking embedded templates dir %s: %w", embedDirToWalk, err)
	}
	if len(templates) == 0 {
		logger.Warn("No embedded templates loaded", "directory_in_fs", embedDirToWalk)
	} else {
		logger.Debug("Loaded templates from embed", "count", len(templates))
	}
	return templates, nil
}

func loadChromaStyleFromXML(logger *slog.Logger, diskFilePath string, embedFS fs.FS, pathInEmbedFS string) (*chroma.Style, error) {
	var data []byte
	var err error

	if _, statErr := os.Stat(diskFilePath); statErr == nil {
		logger.Info("Loading chroma style from disk", "path", diskFilePath)
		data, err = os.ReadFile(diskFilePath)
	} else {
		logger.Info("Loading chroma style from embedded FS", "path_in_fs", pathInEmbedFS)
		data, err = fs.ReadFile(embedFS, pathInEmbedFS)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read style XML: %w", err)
	}

	var styleXML struct {
		Entries []struct {
			Token string `xml:"type,attr"`
			Value string `xml:"style,attr"`
		} `xml:"entry"`
	}

	if err := xml.Unmarshal(data, &styleXML); err != nil {
		return nil, fmt.Errorf("invalid XML in chroma style: %w", err)
	}

	builder := chroma.NewStyleBuilder("abyss")
	for _, e := range styleXML.Entries {
		if e.Token == "" {
			logger.Warn("Empty token found in style XML, skipping")
			continue
		}
		tokenType, tokenErr := chroma.TokenTypeString(e.Token)
		if tokenErr != nil {
			logger.Warn("Invalid token type in chroma style XML", "token", e.Token, "error", tokenErr)
			continue
		}
		styleEntry, styleErr := chroma.ParseStyleEntry(e.Value)
		if styleErr != nil {
			logger.Warn("Invalid style value in chroma style XML", "token", e.Token, "value", e.Value, "error", styleErr)
			continue
		}
		builder.AddEntry(tokenType, styleEntry)
	}

	style, buildErr := builder.Build()
	if buildErr != nil {
		return nil, fmt.Errorf("chroma style build failed: %w", buildErr)
	}
	logger.Info("Custom chroma style loaded successfully.")
	return style, nil
}

func (app *Application) GetLastUploadedFile() string {
	return app.lastUploadedFile
}

func (app *Application) SetLastUploadedFile(filePath string) {
	app.lastUploadedFile = filePath
	app.Logger.Debug("Last uploaded file set", "path", filePath)
}
