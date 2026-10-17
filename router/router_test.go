package router

import (
	"encoding/json"
	"host-agent/conf"
	"net/http"
	"net/http/httptest"
	"testing"
)

func doRequest(t *testing.T, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	New().ServeHTTP(w, httptest.NewRequest(method, path, nil))
	return w
}

func decryptResponse(t *testing.T, w *httptest.ResponseRecorder) conf.Response {
	t.Helper()
	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际 %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/octet-stream" {
		t.Fatalf("期望 Content-Type application/octet-stream，实际 %q", ct)
	}
	plain, err := conf.DecryptAESGCM(w.Body.Bytes())
	if err != nil {
		t.Fatalf("解密响应失败: %v", err)
	}
	var resp conf.Response
	if err := json.Unmarshal(plain, &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	return resp
}

func TestGetSystemInfo(t *testing.T) {
	resp := decryptResponse(t, doRequest(t, http.MethodGet, "/get-system-info"))
	if resp.Status != 0 {
		t.Fatalf("期望 status 0，实际 %d（%s）", resp.Status, resp.Message)
	}
	if len(resp.Data) == 0 {
		t.Fatal("期望返回系统信息数据")
	}
}

func TestRestartContainer(t *testing.T) {
	resp := decryptResponse(t, doRequest(t, http.MethodPost, "/restart/containter/not-exist-container"))
	// 本地通常无同名容器或无 docker 守护进程，应返回业务错误而非成功
	if resp.Status != -1 {
		t.Fatalf("期望 status -1，实际 %d", resp.Status)
	}
	if resp.Message == "" {
		t.Fatal("期望返回错误信息")
	}
}

func TestMethodNotAllowed(t *testing.T) {
	cases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/restart/containter/foo"},
		{http.MethodPost, "/get-system-info"},
		{http.MethodPost, "/get-system-trends"},
	}
	for _, c := range cases {
		if w := doRequest(t, c.method, c.path); w.Code == http.StatusOK {
			t.Errorf("%s %s 应被拒绝，实际 200", c.method, c.path)
		}
	}
}
