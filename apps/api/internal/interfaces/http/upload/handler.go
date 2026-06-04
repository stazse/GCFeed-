package interfaceshttpupload

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// 允许上传的文件类型
var allowedKinds = map[string]bool{
	"video":  true,
	"cover":  true,
	"avatar": true,
}

// Handler 处理文件上传。
type Handler struct {
	baseDir string // 上传文件存放的根目录，比如 "./uploads"
}

func New(baseDir string) *Handler {
	return &Handler{baseDir: baseDir}
}

// Create 上传文件：POST /api/uploads
func (h *Handler) Create(c *gin.Context) {
	// 第一步：获取上传类型（video / cover / avatar）
	kind := strings.TrimSpace(c.PostForm("kind"))
	if kind == "" {
		kind = "file" // 默认类型
	}
	if !allowedKinds[kind] && kind != "file" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid kind"})
		return
	}

	// 第二步：获取文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// 第三步：确保目录存在
	dir := filepath.Join(h.baseDir, kind)
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create directory"})
		return
	}

	// 第四步：保存文件
	// 用原始文件名保存（生产环境应该用随机名防冲突，这里简化）
	dst := filepath.Join(dir, file.Filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	// 第五步：返回文件访问路径
	url := fmt.Sprintf("/uploads/%s/%s", kind, file.Filename)
	c.JSON(http.StatusCreated, gin.H{
		"url":      url,
		"kind":     kind,
		"filename": file.Filename,
		"size":     file.Size,
	})
}

// 下划线消除编译警告
var _ = io.Discard