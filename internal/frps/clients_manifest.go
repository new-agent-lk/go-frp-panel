package frps

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/xxl6097/glog/pkg/z"
	"github.com/xxl6097/go-frp-panel/pkg"
	"github.com/xxl6097/go-service/pkg/github"
)

// clientAsset 是客户端版本清单里的一条。
// 清单随 release 一起发布：go-service 只会请求 /releases/latest，拿不到历史版本，
// 所以「有哪些版本」由清单自己说明。
type clientAsset struct {
	Version         string `json:"version"`
	Platform        string `json:"platform"`
	File            string `json:"file"`
	Sha256          string `json:"sha256"`
	MinPanelVersion string `json:"minPanelVersion"`
}

// clientManifestUrls 返回清单里所有客户端二进制的下载地址。
// 先读 <exeDir>/clients/index.json（离线部署可直接放一份），没有再拉 latest release 上的同名资产。
func clientManifestUrls(dir string) []string {
	data, err := os.ReadFile(filepath.Join(dir, "index.json"))
	if err != nil {
		url := fmt.Sprintf("https://github.com/%s/%s/releases/latest/download/index.json", pkg.GithubUser, pkg.GithubRepo)
		data = fetchWithProxy(url)
	}
	if len(data) == 0 {
		return nil
	}
	var assets []clientAsset
	if err := json.Unmarshal(data, &assets); err != nil {
		z.Error(fmt.Sprintf("客户端清单解析失败: %v", err))
		return nil
	}
	urls := make([]string, 0, len(assets))
	for _, a := range assets {
		if a.Version == "" || a.File == "" {
			continue
		}
		urls = append(urls, fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s",
			pkg.GithubUser, pkg.GithubRepo, a.Version, a.File))
	}
	return urls
}

// fetchWithProxy 依次试直连与各 GitHub 代理，返回第一份成功的内容。
func fetchWithProxy(url string) []byte {
	client := &http.Client{Timeout: 10 * time.Second}
	for _, u := range github.Api().GetProxyUrls(url) {
		resp, err := client.Get(u)
		if err != nil {
			continue
		}
		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err == nil && resp.StatusCode == http.StatusOK {
			return data
		}
	}
	return nil
}
