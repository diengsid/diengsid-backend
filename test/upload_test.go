package test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"id.diengs.backend/internal/model"

	"encoding/json"
)

// buildUploadRequest membuat multipart request dengan satu file pada field "file".
func buildUploadRequest(t *testing.T, fileName string, content []byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", fileName)
	assert.Nil(t, err)
	fw.Write(content)
	w.Close()

	r := httptest.NewRequest(http.MethodPost, "/api/upload", &buf)
	r.Header.Set("Content-Type", w.FormDataContentType())
	return r
}

// buildUploadsRequest membuat multipart request dengan beberapa file pada field "files".
func buildUploadsRequest(t *testing.T, files map[string][]byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for fileName, content := range files {
		fw, err := w.CreateFormFile("files", fileName)
		assert.Nil(t, err)
		fw.Write(content)
	}
	w.Close()

	r := httptest.NewRequest(http.MethodPost, "/api/uploads", &buf)
	r.Header.Set("Content-Type", w.FormDataContentType())
	return r
}

func smallFile() []byte { return bytes.Repeat([]byte("x"), 512*1024) }       // 512 KB
func largeFile() []byte { return bytes.Repeat([]byte("x"), 2*1024*1024+1) }  // 2 MB + 1 byte
func hugeFile() []byte  { return bytes.Repeat([]byte("x"), 10*1024*1024+1) } // 10 MB + 1 byte

// ─── POST /api/upload (single) ────────────────────────────────────────────────

func TestUpload_Success(t *testing.T) {
	req := buildUploadRequest(t, "foto.jpg", smallFile())
	resp, err := app.Test(req)

	assert.Nil(t, err)
	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[model.UploadResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, respBody.Success)
	assert.NotEmpty(t, respBody.Data.Url)
	assert.True(t, strings.Contains(respBody.Data.Url, "uploads/"))
	assert.True(t, strings.HasSuffix(respBody.Data.Url, ".jpg"))
}

func TestUpload_NoFile(t *testing.T) {
	// multipart form kosong tanpa field "file"
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := app.Test(req)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUpload_FileTooLarge(t *testing.T) {
	req := buildUploadRequest(t, "besar.jpg", largeFile())
	resp, err := app.Test(req)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUpload_NonMultipartRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/upload", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ─── POST /api/uploads (multiple) ────────────────────────────────────────────

func TestUploads_Success(t *testing.T) {
	req := buildUploadsRequest(t, map[string][]byte{
		"foto1.jpg": smallFile(),
		"foto2.png": smallFile(),
	})
	resp, err := app.Test(req)

	assert.Nil(t, err)
	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[model.UploadResponses]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, respBody.Success)
	assert.Len(t, respBody.Data.Urls, 2)
	for _, url := range respBody.Data.Urls {
		assert.True(t, strings.Contains(url, "uploads/"))
	}
}

func TestUploads_NoFiles(t *testing.T) {
	// multipart form kosong tanpa field "files"
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/uploads", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := app.Test(req)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// func TestUploads_AllFilesTooLarge(t *testing.T) {
// 	semua file > 10MB → di-skip semua → tidak ada hasil → 400
// 	req := buildUploadsRequest(t, map[string][]byte{
// 		"besar1.jpg": hugeFile(),
// 		"besar2.jpg": hugeFile(),
// 	})
// 	resp, _ := app.Test(req)

// 	assert.Nil(t, err)
// 	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
// }

// func TestUploads_SkipsOversizedFile(t *testing.T) {
// 	// satu valid + satu terlalu besar → hanya 1 URL dikembalikan
// 	req := buildUploadsRequest(t, map[string][]byte{
// 		"valid.jpg": smallFile(),
// 		"besar.jpg": hugeFile(),
// 	})
// 	resp, err := app.Test(req, 30000) // timeout 30s untuk file besar

// 	assert.Nil(t, err)
// 	body, _ := io.ReadAll(resp.Body)
// 	var respBody model.WebResponse[model.UploadResponses]
// 	json.Unmarshal(body, &respBody)

// 	assert.Equal(t, http.StatusOK, resp.StatusCode)
// 	assert.True(t, respBody.Success)
// 	assert.Len(t, respBody.Data.Urls, 1)
// }
