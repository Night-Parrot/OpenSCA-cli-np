package main

import (
	"context"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/Night-Parrot/OpenSCA-cli-np/v3/opensca/common"
	"github.com/Night-Parrot/OpenSCA-cli-np/v3/opensca/logs"
	"github.com/Night-Parrot/OpenSCA-cli-np/v3/opensca/model"
	"github.com/Night-Parrot/OpenSCA-cli-np/v3/opensca/sca/javascript"
)

func init() {

	// register npm repository origin
	javascript.RegisterNpmOrigin(func(name, version string) *javascript.PackageJson {
		var pkgjson *javascript.PackageJson
		common.DownloadUrlFromRepos(name, func(repo common.RepoConfig, r io.Reader) { pkgjson = javascript.ReadNpmJson(r, version) },
			common.RepoConfig{Url: "https://r.cnpmjs.org/"},
		)
		return pkgjson
	})
}

func main() {

	projectDir := "../../test/javascript/5"

	sca := javascript.Sca{}

	var files []*model.File
	filepath.WalkDir(projectDir, func(path string, d fs.DirEntry, err error) error {
		if !sca.Filter(path) {
			return nil
		}
		file := model.NewFile(path, strings.TrimPrefix(path, projectDir))
		files = append(files, file)
		return nil
	})

	sca.Sca(context.TODO(), nil, files, func(file *model.File, root ...*model.DepGraph) {
		for _, dep := range root {
			dep.Build(false, sca.Language())
			logs.Infof("file %s:\n%s", file.Relpath(), dep.Tree(false, false))
		}
	})
}
