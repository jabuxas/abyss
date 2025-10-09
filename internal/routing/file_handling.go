package routing

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jabuxas/abyss/internal/utils"
)

func indexHandler(c *gin.Context) {
	c.File("assets/static/index.html")
}

func deleteFileHandler(c *gin.Context) {
	filename := c.Param("file")
	filePath := filepath.Join(CFG.FilesDir, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.String(http.StatusNotFound, "file not found")
		return
	}

	err := os.Remove(filePath)
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to delete file")
		return
	}

	err = os.Remove(utils.JsonPathFromFilePath(filePath))
	if os.IsNotExist(err) {
		// metadata file doesn't exist, nothing to do
	} else if err != nil {
		log.Printf("failed to delete metadata file: %v", err)
	}

	c.String(http.StatusOK, "file deleted successfully")
}

func serveFileHandler(c *gin.Context) {
	filename := c.Param("file")
	filePath := filepath.Join(CFG.FilesDir, filename)

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		c.String(http.StatusNotFound, "file not found")
		return
	}

	meta, err := utils.ReadMetadata(filePath)
	if err != nil {
		c.String(http.StatusInternalServerError, "could not read file metadata")
		return
	}

	isProtected := len(meta.PasswordHash) > 0

	if !isProtected {
		servePublicFile(c, filename, filePath, fileInfo)
		return
	}

	sessionToken, err := c.Cookie(sessionCookieName)
	if err == nil && GetSession(sessionToken, filename) {
		servePublicFile(c, filename, filePath, fileInfo)
		return
	}

	switch c.Request.Method {
	case "GET":
		c.HTML(http.StatusUnauthorized, "passwordPrompt.html", nil)
	case "POST":
		password := c.PostForm("password")
		if utils.CheckPassword(password, meta.PasswordHash) {
			token, err := NewSession(filename)
			if err != nil {
				c.String(http.StatusInternalServerError, "failed to create session")
				return
			}
			c.SetCookie(sessionCookieName, token, int(sessionDuration.Seconds()), "/", "", false, true)
			c.Redirect(http.StatusFound, c.Request.URL.Path)
		} else {
			c.HTML(http.StatusUnauthorized, "passwordPrompt.html", gin.H{
				"Error": "invalid password. please, try again.",
			})
		}
	default:
		c.String(http.StatusMethodNotAllowed, "method not allowed")
	}
}

func servePublicFile(c *gin.Context, filename, filePath string, fileInfo os.FileInfo) {
	fileType := utils.DetectFileType(filename)
	if fileType == "unknown" {
		c.Redirect(http.StatusSeeOther, "/raw/"+filename)
		return
	}

	fileData := FileData{
		Name:       filename,
		Path:       "/raw/" + filename,
		Extension:  fileType,
		ModTimeStr: fileInfo.ModTime().Format("2006-01-02 15:04:05"),
	}

	if fileType == "text" {
		content, err := os.ReadFile(filePath)
		if err == nil {
			fileData.Content = string(content)

			highlighted, err := utils.HighlightCode(fileData.Content, filename)
			if err != nil {
				log.Printf("failed to highlight code: %v", err)
				fileData.Content = string(content)
			} else {
				fileData.HighlightedContent = highlighted
			}
		}
	}

	c.HTML(http.StatusOK, "fileDisplay.html", gin.H{"data": fileData})
}

func serveRawFileHandler(c *gin.Context) {
	filename := c.Param("file")
	filePath := filepath.Join(CFG.FilesDir, filename)

	if _, err := os.Stat(filePath); err != nil {
		c.String(http.StatusNotFound, "file not found")
		return
	}

	meta, err := utils.ReadMetadata(filePath)
	if err != nil {
		c.String(http.StatusInternalServerError, "could not read file metadata")
		return
	}

	if len(meta.PasswordHash) > 0 {
		sessionToken, err := c.Cookie(sessionCookieName)
		if err != nil || !GetSession(sessionToken, filename) {
			c.String(http.StatusUnauthorized, "unauthorized: password required")
			return
		}
	}

	c.File(filePath)
}

func uploadFileHandler(c *gin.Context) {
	if !isAuthorized(c) {
		c.String(http.StatusUnauthorized, "unauthorized")
		return
	}
	file, _ := c.FormFile("file")
	secretName := len(c.Request.Form["secret"]) > 0
	fileName := utils.HashedName(file.Filename, secretName)
	savePath := filepath.Join(CFG.FilesDir, fileName)

	c.SaveUploadedFile(file, savePath)

	var expiry *time.Time
	var passwordHash []byte
	var err error

	if len(c.Request.Form["expiration"]) > 0 {
		expiry, err = utils.ParseExpiration(c.Request.FormValue("expiration"))
		if err != nil {
			log.Println("failed to parse expiration:", err)
		}
	}

	if len(c.Request.Form["password"]) > 0 {
		passwordHash, err = utils.ParsePassword(c.Request.FormValue("password"))
		if err != nil {
			log.Println("failed to hash password:", err)
		}
	}

	err = utils.SaveMetadata(savePath, expiry, passwordHash)
	if err != nil {
		log.Println("failed to save metadata:", err)
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	fullURL := fmt.Sprintf("%s://%s/%s", scheme, c.Request.Host, fileName)

	c.String(http.StatusOK, fullURL+"\n")
}

func listFilesHandler(c *gin.Context) {
	dirEntries, err := os.ReadDir(CFG.FilesDir)
	if err != nil {
		if os.IsNotExist(err) {
			c.String(http.StatusNotFound, "directory not found")
			return
		}
		c.String(http.StatusInternalServerError, "failed to read directory")
		return
	}

	var fileInfos []FileData
	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			log.Println("failed to get file info for", entry.Name(), ":", err)
			continue
		}

		linkPath := filepath.Join("/" + entry.Name())
		fileInfos = append(fileInfos, FileData{
			Name:          entry.Name(),
			Path:          linkPath,
			ModTimeStr:    info.ModTime().UTC().Format("2006-01-02 15:04:05 UTC"),
			ModTime:       info.ModTime(),
			FormattedSize: utils.FormatFileSize(info.Size()),
			Extension:     filepath.Ext(entry.Name()),
		})
	}

	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].ModTime.After(fileInfos[j].ModTime)
	})

	c.HTML(http.StatusOK, "fileList.html", gin.H{
		"Files": fileInfos,
		"URL":   CFG.AbyssURL,
	})
}
