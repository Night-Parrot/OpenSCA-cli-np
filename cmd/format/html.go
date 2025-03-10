package format

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"io"

	"github.com/Night-Parrot/OpenSCA-cli-np/v3/cmd/detail"
	"github.com/Night-Parrot/OpenSCA-cli-np/v3/opensca/logs"
)

//go:embed html_tpl
var index []byte

// html组件字段
type htmlDep struct {
	*detail.DepDetailGraph
	SecId    int         `json:"security_level_id,omitempty"`
	Statis   map[int]int `json:"vuln_statis"`
	Children any         `json:"children,omitempty"`
}

// html统计信息
type htmlStatis struct {
	Component map[int]int `json:"component"`
	Vuln      map[int]int `json:"vuln"`
}

func Html(report Report, out string) {

	deps := []htmlDep{}
	statis := htmlStatis{
		Component: map[int]int{},
		Vuln:      map[int]int{},
	}
	vulnMap := map[string]int{}

	// 遍历所有组件
	report.DepDetailGraph.ForEach(func(n *detail.DepDetailGraph) bool {

		// 组件风险
		secid := 5
		// 不同风险等级的漏洞数
		vuln_statis := map[int]int{}
		for _, v := range n.Vulnerabilities {
			vulnMap[v.Id] = v.SecurityLevelId
			vuln_statis[v.SecurityLevelId]++
			if secid > v.SecurityLevelId {
				secid = v.SecurityLevelId
			}
		}

		if n.Name != "" {
			statis.Component[secid]++
			deps = append(deps, htmlDep{
				DepDetailGraph: n,
				Children:       nil,
				SecId:          secid,
				Statis:         vuln_statis,
			})
		}

		return true
	})

	// 统计漏洞风险
	for _, secid := range vulnMap {
		statis.Vuln[secid]++
	}

	// report依赖信息临时置空用于生成html报告
	graph := report.DepDetailGraph
	report.DepDetailGraph = nil
	defer func() { report.DepDetailGraph = graph }()

	// 生成html报告需要的json数据
	if data, err := json.Marshal(struct {
		TaskInfo   TaskInfo   `json:"task_info"`
		Statis     htmlStatis `json:"statis"`
		Components []htmlDep  `json:"components"`
	}{
		TaskInfo:   report.TaskInfo,
		Statis:     statis,
		Components: deps,
	}); err != nil {
		logs.Warn(err)
	} else {
		outWrite(out, func(w io.Writer) error {
			_, err := w.Write(bytes.Replace(index, []byte(`"此处填充json数据"`), data, 1))
			return err
		})
		return
	}
}
