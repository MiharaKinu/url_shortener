package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	initConfig()
	// 初始化数据库
	urlDB = NewURLDatabase()
	// 清理旧数据
	urlDB.db.Exec("DELETE FROM url_mapping")

	r := gin.Default()
	r.POST("/short", shortenURL)
	r.POST("/decode", decodeURL)
	r.GET("/:shortID", redirectURL)
	return r
}

func TestShortenURL(t *testing.T) {
	r := setupTestRouter()
	defer urlDB.Close() // 测试结束后关闭数据库连接

	// 测试用例1：正常情况
	body := map[string]string{
		"url": "https://example.com/test",
	}
	jsonData, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/short", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	// 解析返回的JSON响应
	var response struct {
		Code string `json:"code"`
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "200", response.Code)

	// 测试用例2：测试重定向功能
	fullURL := response.Data.URL
	shortID := fullURL[strings.LastIndex(fullURL, "/")+1:]
	w2redirect := httptest.NewRecorder()
	req2redirect, _ := http.NewRequest("GET", "/"+shortID, nil)
	r.ServeHTTP(w2redirect, req2redirect)
	assert.Equal(t, 302, w2redirect.Code)
	assert.Equal(t, "https://example.com/test", w2redirect.Header().Get("Location")) // 检查重定向地址是否正确

	// 测试用例2.1：解码刚刚生成的短URL
	decodeBody := map[string]string{
		"url": response.Data.URL,
	}
	jsonDecodeData, _ := json.Marshal(decodeBody)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/decode", bytes.NewBuffer(jsonDecodeData))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)

	assert.Equal(t, 200, w2.Code)
	var decodeResponse struct {
		Code string `json:"code"`
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	err = json.Unmarshal(w2.Body.Bytes(), &decodeResponse)
	assert.NoError(t, err)
	assert.Equal(t, "200", decodeResponse.Code)
	assert.Equal(t, "https://example.com/test", decodeResponse.Data.URL)

	// 测试用例3：尝试缩短不在allowDomain列表中的URL
	invalidBody := map[string]string{
		"url": "https://invalid-domain.com/test",
	}
	jsonInvalidData, _ := json.Marshal(invalidBody)

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("POST", "/short", bytes.NewBuffer(jsonInvalidData))
	req3.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w3, req3)

	assert.Equal(t, 200, w3.Code)
	var invalidResponse struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}
	err = json.Unmarshal(w3.Body.Bytes(), &invalidResponse)
	assert.NoError(t, err)
	assert.Equal(t, "-1", invalidResponse.Code)
	assert.Contains(t, invalidResponse.Msg, "not included in allowDomain")
}
