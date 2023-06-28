/*
 * @Descripation: 查找漏洞
 * @Date: 2021-11-16 21:05:05
 */

package vuln

import (
	"errors"
	"util/args"
	"util/model"
)

// SearchVuln 查找漏洞
func SearchVuln(root *model.DepTree) (err error) {
	queue := model.NewQueue()
	queue.Push(root)
	deps := []*model.DepTree{}
	for !queue.Empty() {
		node := queue.Pop().(*model.DepTree)
		deps = append(deps, node)
		for _, child := range node.Children {
			queue.Push(child)
		}
	}
	localVulns := [][]*model.Vuln{}
	serverVulns := [][]*model.Vuln{}
	ds := make([]model.Dependency, len(deps))
	for i, d := range deps {
		ds[i] = d.Dependency
	}
	if args.Config.VulnDB != "" {
		localVulns = GetLocalVulns(ds)
	}
	if args.Config.Url != "" && args.Config.Token != "" {
		serverVulns, err = GetServerVuln(ds)
	} else if args.Config.VulnDB == "" && args.Config.Url == "" && args.Config.Token != "" {
		err = errors.New("url is null")
	} else if args.Config.VulnDB == "" && args.Config.Url != "" && args.Config.Token == "" {
		err = errors.New("token is null")
	}
	for i, dep := range deps {
		// 合并本地和云端库搜索的漏洞
		exist := map[string]struct{}{}
		if len(localVulns) != 0 {
			for _, vuln := range localVulns[i] {
				if vuln.Id == "" {
					continue
				}
				if _, ok := exist[vuln.Id]; !ok {
					exist[vuln.Id] = struct{}{}
					dep.Vulnerabilities = append(dep.Vulnerabilities, vuln)
				}
			}
		}
		if len(serverVulns) != 0 {
			for _, vuln := range serverVulns[i] {
				if vuln.Id == "" {
					// 约定没有id时按照许可证处理
					dep.AddLicense(vuln.Name)
					continue
				}
				if _, ok := exist[vuln.Id]; !ok {
					exist[vuln.Id] = struct{}{}
					dep.Vulnerabilities = append(dep.Vulnerabilities, vuln)
				}
			}
		}
	}
	return
}
