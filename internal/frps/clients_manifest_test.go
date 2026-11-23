package frps

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClientManifestUrlsFromLocalFile(t *testing.T) {
	dir := t.TempDir()
	data := `[{"version":"v0.4.15","platform":"linux_amd64","file":"acfrpc_v0.4.15_linux_amd64","sha256":"x"},
{"version":"v0.4.16","platform":"linux_amd64","file":"acfrpc_v0.4.16_linux_amd64","sha256":"y"},
{"version":"v0.4.16","platform":"linux_amd64.patch","file":"acfrpc_v0.4.16_linux_amd64.patch","sha256":"z"},
{"version":"","file":"acfrpc_broken"}]`
	if err := os.WriteFile(filepath.Join(dir, "index.json"), []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	urls := clientManifestUrls(dir)
	if len(urls) != 2 {
		t.Fatalf("期望 2 条下载地址（空 version 与 .patch 增量包应被跳过），得到 %d: %v", len(urls), urls)
	}
	for _, u := range urls {
		if strings.Contains(u, ".patch") {
			t.Errorf(".patch 增量包不该出现在下载地址里: %s", u)
		}
	}
	if !strings.HasSuffix(urls[0], "/releases/download/v0.4.15/acfrpc_v0.4.15_linux_amd64") {
		t.Errorf("下载地址拼接不对: %s", urls[0])
	}
}
