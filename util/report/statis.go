package report

import (
	"fmt"

	"github.com/xmirrorsecurity/opensca-cli/util/args"
	"github.com/xmirrorsecurity/opensca-cli/util/model"
)

// Statis 统计概览信息
func Statis(depRoot *model.DepTree, taskInfo TaskInfo) string {
	if taskInfo.Error != nil || taskInfo.ErrorString != "" {
		return ""
	}
	coms := map[int]int{
		0: 0, 1: 0, 2: 0, 3: 0, 4: 0, 5: 0,
	}
	vuls := map[int]int{
		0: 0, 1: 0, 2: 0, 3: 0, 4: 0,
	}
	vset := map[string]bool{}
	q := []*model.DepTree{depRoot}
	depSet := map[string]bool{}
	for len(q) > 0 {
		n := q[0]
		if _, ok := depSet[n.Dependency.String()]; !(args.Config.Dedup && ok || n.Name == "") {
			depSet[n.Dependency.String()] = true
			risk := 5
			for _, v := range n.Vulnerabilities {
				if !vset[v.Id] {
					vset[v.Id] = true
					vuls[v.SecurityLevelId]++
					vuls[0]++
				}
				if v.SecurityLevelId < risk {
					risk = v.SecurityLevelId
				}
			}
			coms[risk]++
			coms[0]++
		}
		q = append(q[1:], n.Children...)
	}
	return fmt.Sprintf("\n\nComplete!"+
		"\nComponents:%d C:%d H:%d M:%d L:%d"+
		"\nVulnerabilities:%d C:%d H:%d M:%d L:%d",
		coms[0], coms[1], coms[2], coms[3], coms[4],
		vuls[0], vuls[1], vuls[2], vuls[3], vuls[4],
	)
}
