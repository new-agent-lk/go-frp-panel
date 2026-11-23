package utils

import "testing"

func TestToTreeGroupsByVersion(t *testing.T) {
	entries := []string{
		"https://x/acfrpc_v0.4.15_linux_amd64",
		"https://x/acfrpc_v0.4.16_linux_amd64",
		"https://x/acfrpc_v0.4.16_windows_amd64.exe",
		"https://x/acfrpc_v0.4.16_linux_amd64", // 本地与清单重复，应被去重
		"github_proxy.json",                    // 段数不足，忽略
	}
	got := ToTree("", entries)

	if len(got) != 2 {
		t.Fatalf("期望 2 个版本分组，得到 %d: %+v", len(got), got)
	}
	byVersion := make(map[string][]string, len(got))
	for _, o := range got {
		for _, c := range o.Children {
			byVersion[o.Value] = append(byVersion[o.Value], c.Label)
		}
	}
	if n := len(byVersion["acfrpc_v0.4.16"]); n != 2 {
		t.Errorf("v0.4.16 期望 2 个平台叶子，得到 %d: %v", n, byVersion["acfrpc_v0.4.16"])
	}
	if n := len(byVersion["acfrpc_v0.4.15"]); n != 1 {
		t.Errorf("v0.4.15 期望 1 个平台叶子，得到 %d: %v", n, byVersion["acfrpc_v0.4.15"])
	}
	for _, o := range got {
		for _, c := range o.Children {
			if c.Label == "" || c.FilePath == "" {
				t.Errorf("叶子缺少 label 或 filePath: %+v", c)
			}
		}
	}
}
