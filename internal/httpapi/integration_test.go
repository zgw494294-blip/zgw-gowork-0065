package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"maskreview/internal/service"
	"maskreview/internal/store"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	dir := t.TempDir()
	path := dir + "/store.json"
	st, err := store.NewFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	svc := service.NewService(st)
	handler := NewHandler(svc)
	hs := httptest.NewServer(handler)
	t.Cleanup(hs.Close)
	return hs
}

func TestAPIFullChain(t *testing.T) {
	hs := newTestServer(t)
	// 创建产品层级
	resp, err := http.Post(hs.URL+"/api/product-levels", "application/json", bytes.NewBufferString(`{"name":"L1"}`))
	if err != nil {
		t.Fatal(err)
	}
	var pl map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&pl)
	levelID := pl["id"].(string)

	// 创建版本
	body := `{"level_id":"` + levelID + `","version":1,"description":"v1"}`
	resp, err = http.Post(hs.URL+"/api/mask-versions", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	var mv map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&mv)
	mvID := mv["id"].(string)

	// 创建变更申请
	body = `{"mask_version_id":"` + mvID + `"}`
	resp, err = http.Post(hs.URL+"/api/change-requests", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	var cr map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&cr)
	crID := cr["id"].(string)

	// 送审
	resp, err = http.Post(hs.URL+"/api/change-requests/"+crID+"/submit", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}

	// 创建验证批次
	body = `{"change_request_id":"` + crID + `"}`
	resp, err = http.Post(hs.URL+"/api/verification-batches", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	var vb map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&vb)
	vbID := vb["id"].(string)

	// 提交验证结果
	body = `{"critical_passed":true,"note":"ok"}`
	resp, err = http.Post(hs.URL+"/api/verification-batches/"+vbID+"/result", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}

	// 审核通过
	body = `{"conclusion":"pass","reason":"good"}`
	resp, err = http.Post(hs.URL+"/api/change-requests/"+crID+"/audit", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}

	// 启用
	resp, err = http.Post(hs.URL+"/api/mask-versions/"+mvID+"/enable", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}

	// 确认状态
	resp, err = http.Get(hs.URL + "/api/mask-versions?level_id=" + levelID)
	if err != nil {
		t.Fatal(err)
	}
	var mvs []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&mvs)
	if len(mvs) != 1 || mvs[0]["status"] != "active" {
		t.Fatalf("unexpected mask versions: %v", mvs)
	}
}

func TestAPIBatchConstraint(t *testing.T) {
	hs := newTestServer(t)
	// 创建层级、版本、变更申请但不送审，直接创建批次
	resp, _ := http.Post(hs.URL+"/api/product-levels", "application/json", bytes.NewBufferString(`{"name":"L2"}`))
	var pl map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&pl)
	levelID := pl["id"].(string)

	body := `{"level_id":"` + levelID + `","version":1,"description":"v1"}`
	resp, _ = http.Post(hs.URL+"/api/mask-versions", "application/json", bytes.NewBufferString(body))
	var mv map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&mv)
	mvID := mv["id"].(string)

	body = `{"mask_version_id":"` + mvID + `"}`
	resp, _ = http.Post(hs.URL+"/api/change-requests", "application/json", bytes.NewBufferString(body))
	var cr map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&cr)
	crID := cr["id"].(string)

	body = `{"change_request_id":"` + crID + `"}`
	resp, err := http.Post(hs.URL+"/api/verification-batches", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
