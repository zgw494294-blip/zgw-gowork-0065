package audit

import (
	"os"
	"sync"
	"time"
)

// 简单的审计日志写入，供非API能力使用
type FileAudit struct {
	mu   sync.Mutex
	path string
}

func NewFileAudit(path string) *FileAudit {
	return &FileAudit{path: path}
}

func (a *FileAudit) Log(entry string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	f, err := os.OpenFile(a.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(time.Now().Format(time.RFC3339) + " " + entry + "\n")
	return err
}
