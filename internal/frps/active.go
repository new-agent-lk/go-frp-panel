package frps

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/xxl6097/glog/pkg/z"
	"github.com/xxl6097/go-frp-panel/pkg/comm"
	"github.com/xxl6097/go-frp-panel/pkg/utils"
	"github.com/xxl6097/go-service/pkg/github"
	model3 "github.com/xxl6097/go-service/pkg/github/model"
)

// activeState 记录当前激活的 acfrpc 版本与上一版，上一版用于回滚。
// shortcut: 激活只做标注与回滚，不参与下发；等确实需要「不选版本也能生成客户端」再接到 client_gen
type activeState struct {
	Version  string `json:"version"`
	Previous string `json:"previous"`
}

func activeFile(dir string) string {
	return filepath.Join(dir, "active.json")
}

func readActive(dir string) activeState {
	var s activeState
	data, err := os.ReadFile(activeFile(dir))
	if err != nil {
		return s
	}
	_ = json.Unmarshal(data, &s)
	return s
}

func writeActive(dir string, s activeState) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err := utils.MakeDir(dir); err != nil {
		return err
	}
	return os.WriteFile(activeFile(dir), data, 0o644)
}

func (this *frps) clientsDir() string {
	return filepath.Join(this.binDir, "clients")
}

// clientTree 把三处来源合成一棵版本树：本地目录、版本清单、GitHub latest 兜底。
func (this *frps) clientTree() []utils.Option {
	dir := this.clientsDir()
	entries := utils.ListNames(dir)
	entries = append(entries, clientManifestUrls(dir)...)
	entries = append(entries, github.Api().GetDownloadUrls(func(version string, assets *model3.Assets) bool {
		return assets != nil &&
			strings.HasPrefix(assets.Name, "acfrpc") &&
			!strings.Contains(assets.Name, ".patch")
	})...)
	return utils.ToTree(dir, entries)
}

func (this *frps) apiClientGet(w http.ResponseWriter, r *http.Request) {
	res, f := comm.Response(r)
	defer f(w)
	tree := this.clientTree()
	active := readActive(this.clientsDir())
	for i := range tree {
		tree[i].Active = tree[i].Value == active.Version
	}
	res.Data = tree
}

func (this *frps) apiClientActiveGet(w http.ResponseWriter, r *http.Request) {
	res, f := comm.Response(r)
	defer f(w)
	res.Data = readActive(this.clientsDir())
}

// nextActive 算出切换后的状态：version 为空表示回滚到上一版。
func nextActive(cur activeState, version string) (activeState, error) {
	if version == "" {
		if cur.Previous == "" {
			return cur, fmt.Errorf("没有可回滚的版本")
		}
		return activeState{Version: cur.Previous, Previous: cur.Version}, nil
	}
	if version == cur.Version {
		return cur, fmt.Errorf("%s 已是当前版本", version)
	}
	return activeState{Version: version, Previous: cur.Version}, nil
}

// apiClientActiveSet 切换激活版本；version 传空表示回滚到上一版。
func (this *frps) apiClientActiveSet(w http.ResponseWriter, r *http.Request) {
	res, f := comm.Response(r)
	defer f(w)
	body, err := utils.GetDataByJson[struct {
		Version string `json:"version"`
	}](r)
	if err != nil {
		res.Err(err)
		return
	}
	if body == nil {
		res.Error("body is nil")
		return
	}
	dir := this.clientsDir()
	if body.Version != "" {
		found := false
		for _, o := range this.clientTree() {
			if o.Value == body.Version {
				found = true
				break
			}
		}
		if !found {
			res.Error(fmt.Sprintf("版本 %s 不存在", body.Version))
			return
		}
	}
	next, err := nextActive(readActive(dir), body.Version)
	if err != nil {
		res.Error(err.Error())
		return
	}
	if err := writeActive(dir, next); err != nil {
		z.Error(err)
		res.Err(err)
		return
	}
	res.Sucess("已切换", next)
}
