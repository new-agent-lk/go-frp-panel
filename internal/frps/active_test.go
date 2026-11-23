package frps

import (
	"path/filepath"
	"testing"
)

func TestActiveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	if got := readActive(dir); got.Version != "" || got.Previous != "" {
		t.Fatalf("没有文件时应是空状态，得到 %+v", got)
	}
	want := activeState{Version: "acfrpc_v0.4.16", Previous: "acfrpc_v0.4.15"}
	if err := writeActive(dir, want); err != nil {
		t.Fatal(err)
	}
	if got := readActive(dir); got != want {
		t.Errorf("读回不一致：%+v", got)
	}
	if _, err := filepath.Abs(activeFile(dir)); err != nil {
		t.Fatal(err)
	}
}

func TestNextActive(t *testing.T) {
	cur := activeState{Version: "v16", Previous: "v15"}

	next, err := nextActive(cur, "v17")
	if err != nil || next.Version != "v17" || next.Previous != "v16" {
		t.Errorf("切到新版本：%+v %v", next, err)
	}

	next, err = nextActive(cur, "")
	if err != nil || next.Version != "v15" || next.Previous != "v16" {
		t.Errorf("回滚应换回上一版：%+v %v", next, err)
	}

	if _, err = nextActive(cur, "v16"); err == nil {
		t.Error("切到当前版本应报错")
	}
	if _, err = nextActive(activeState{}, ""); err == nil {
		t.Error("没有上一版时回滚应报错")
	}
}
