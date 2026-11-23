package utils

import (
	"github.com/xxl6097/go-service/pkg/utils"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Node struct {
	Label    string `json:"label"`
	Value    string `json:"value"`
	FilePath string `json:"filePath"`
}
type Option struct {
	Label    string `json:"label"`
	Value    string `json:"value"`
	Children []Node `json:"children"`
	Active   bool   `json:"active,omitempty"`
}

func SplitLastTwoByUnderscore(s string) []string {
	// 过滤空元素
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_'
	})
	if len(parts) < 2 {
		return []string{}
	}
	return parts[len(parts)-2:]
}

func CleanExt(name string) string {
	filename := path.Base(name) // 获取文件名"app.log"
	nameOnly := filename[:len(filename)-len(path.Ext(filename))]
	return nameOnly
}

func ToTree(dir string, entries []string) []Option {
	maps := make(map[string][]Node)
	for _, uri := range entries {
		var name string
		var fileUri string
		if utils.IsURL(uri) {
			name = path.Base(uri) // 输出：report.pdf
			fileUri = uri
		} else {
			name = uri
			fileUri = filepath.Join(dir, name)
		}
		// 分组键取「名_版本」、叶子取最后两段（os_arch），同一客户端的多个版本才各成一组；
		// 以前用最后两段当分组键，版本号会被丢掉，多版本撞成同名叶子。
		parts := strings.FieldsFunc(name, func(r rune) bool {
			return r == '_'
		})
		if len(parts) < 3 {
			continue
		}
		group := strings.Join(parts[:len(parts)-2], "_")
		// 只在最后一段去扩展名：CleanExt 认的是最后一个点，作用在整名上会把 "v0.4.15" 当扩展名切掉
		leaf := CleanExt(strings.Join(parts[len(parts)-2:], "_"))
		nodeArray := maps[group]
		dup := false
		for _, n := range nodeArray {
			if n.Label == leaf {
				dup = true
				break
			}
		}
		if dup {
			continue // 本地目录与远端清单会给出同一个文件，按叶子名去重
		}
		maps[group] = append(nodeArray, Node{
			Label:    leaf,
			Value:    leaf,
			FilePath: fileUri,
		})
	}

	var options []Option
	for k, v := range maps {
		options = append(options, Option{
			Label:    ToUpperFirst(k),
			Value:    k,
			Children: v,
		})
	}
	return options
}

func ListNames(dir string) []string {
	files := make([]string, 0)
	entries, _ := os.ReadDir(dir) // 读取当前目录
	for _, entry := range entries {
		files = append(files, entry.Name())
	}
	return files
}

func GetNodes(dir string) []Option {
	return ToTree(dir, ListNames(dir))
}

func GetNodes1(dir string) []Option {
	maps := make(map[string][]Node)
	entries, _ := os.ReadDir(dir) // 读取当前目录
	for _, entry := range entries {
		name := entry.Name()
		result := SplitLastTwoByUnderscore(name)
		//fmt.Printf("%-30s => %v\n", name, result)
		if len(result) == 2 {
			nodeArray := maps[result[0]]
			if nodeArray == nil {
				nodeArray = make([]Node, 0)
			}
			nameOnly := CleanExt(result[1])

			nodeArray = append(nodeArray, Node{
				Label:    nameOnly,
				Value:    nameOnly,
				FilePath: filepath.Join(dir, name),
			})
			maps[result[0]] = nodeArray
		}
	}
	var options []Option
	for k, v := range maps {
		options = append(options, Option{
			Label:    ToUpperFirst(k),
			Value:    k,
			Children: v,
		})
	}
	return options
}
