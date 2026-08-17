package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"maskreview/internal/domain"
)

func TestFileStorePersistAndRecover(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	st, err := NewFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	snap := st.GetSnapshot()
	pl := &domain.ProductLevel{ID: "pl1", Name: "Layer1"}
	mv := &domain.MaskVersion{ID: "mv1", LevelID: "pl1", Version: 1, Status: domain.StatusDraft, CreatedAt: time.Now()}
	snap.ProductLevels = append(snap.ProductLevels, pl)
	snap.MaskVersions = append(snap.MaskVersions, mv)
	if err := st.SaveSnapshot(snap); err != nil {
		t.Fatal(err)
	}

	// 重新打开
	st2, err := NewFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	recovered := st2.GetSnapshot()
	if len(recovered.ProductLevels) != 1 || recovered.ProductLevels[0].ID != "pl1" {
		t.Fatalf("unexpected product levels: %+v", recovered.ProductLevels)
	}
	if len(recovered.MaskVersions) != 1 || recovered.MaskVersions[0].ID != "mv1" {
		t.Fatalf("unexpected mask versions: %+v", recovered.MaskVersions)
	}
}

func TestAtomicWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	st, _ := NewFileStore(path)
	snap := st.GetSnapshot()
	if err := st.SaveSnapshot(snap); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("temp file should not exist, got %v", err)
	}
}
