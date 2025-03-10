package sca

import (
	"context"

	"github.com/Night-Parrot/OpenSCA-cli-np/v3/opensca/model"
	"github.com/Night-Parrot/OpenSCA-cli-np/v3/opensca/sca/erlang"
	"github.com/Night-Parrot/OpenSCA-cli-np/v3/opensca/sca/golang"
	"github.com/Night-Parrot/OpenSCA-cli-np/v3/opensca/sca/groovy"
	"github.com/Night-Parrot/OpenSCA-cli-np/v3/opensca/sca/java"
	"github.com/Night-Parrot/OpenSCA-cli-np/v3/opensca/sca/javascript"
	"github.com/Night-Parrot/OpenSCA-cli-np/v3/opensca/sca/php"
	"github.com/Night-Parrot/OpenSCA-cli-np/v3/opensca/sca/python"
	"github.com/Night-Parrot/OpenSCA-cli-np/v3/opensca/sca/ruby"
	"github.com/Night-Parrot/OpenSCA-cli-np/v3/opensca/sca/rust"
	"github.com/Night-Parrot/OpenSCA-cli-np/v3/opensca/sca/sbom"
)

type Sca interface {
	Language() model.Language
	Filter(relpath string) bool
	Sca(ctx context.Context, parent *model.File, files []*model.File, call model.ResCallback)
}

var AllSca = []Sca{
	python.Sca{},
	javascript.Sca{},
	golang.Sca{},
	ruby.Sca{},
	rust.Sca{},
	erlang.Sca{},
	php.Sca{},
	java.Sca{},
	groovy.Sca{},
	sbom.Sca{},
}
