package sbom

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"strings"

	"github.com/Night-Parrot/OpenSCA-cli-np/v3/opensca/model"
)

func ParseSpdx(f *model.File) *model.DepGraph {

	depIdMap := map[string]*model.DepGraph{}
	_dep := model.NewDepGraphMap(func(s ...string) string {
		return s[0]
	}, func(s ...string) *model.DepGraph {
		return &model.DepGraph{
			Vendor:   s[1],
			Name:     s[2],
			Version:  s[3],
			Language: model.Language(s[4]),
		}
	}).LoadOrStore

	// 记录spdx中的tag信息
	tags := map[string]string{}
	checkAndSet := func(k, v string) {
		if _, ok := tags[k]; ok {
			depIdMap[tags["id"]] = _dep(tags["id"], tags["group"], tags["name"], tags["version"], tags["language"])
			tags = map[string]string{}
		}
		tags[k] = strings.TrimSpace(v)
	}
	// 记录relationship
	relation := map[string][]string{}

	start := false
	f.ReadLine(func(line string) {
		i := strings.Index(line, ":")
		if strings.HasPrefix(line, "#") || i == -1 {
			return
		}
		if strings.HasPrefix(line, "PackageName:") {
			start = true
		}
		if !start {
			return
		}
		k := strings.TrimSpace(line[:i])
		v := strings.TrimSpace(line[i+1:])
		switch k {
		case "PackageName":
			checkAndSet("name", v)
		case "PackageVersion":
			checkAndSet("version", v)
		case "PackageSupplier":
			checkAndSet("group", strings.TrimPrefix(v, "Organization:"))
		case "SPDXID":
			checkAndSet("id", v)
		case "ExternalRef":
			words := strings.Fields(v)
			if len(words) >= 3 && words[len(words)-2] == "purl" {
				_, _, _, language := model.ParsePurl(words[len(words)-1])
				checkAndSet("language", string(language))
			}
		case "Relationship":
			ids := strings.Split(v, "DEPENDS_ON")
			if len(ids) == 2 {
				parent := strings.TrimSpace(ids[0])
				child := strings.TrimSpace(ids[1])
				relation[parent] = append(relation[parent], child)
			}
		}
	})
	depIdMap[tags["id"]] = _dep(tags["id"], tags["group"], tags["name"], tags["version"], tags["language"])

	if len(depIdMap) == 0 {
		return nil
	}

	for parent, children := range relation {
		for _, child := range children {
			depIdMap[parent].AppendChild(depIdMap[child])
		}
	}

	var roots []*model.DepGraph
	for _, dep := range depIdMap {
		if len(dep.Parents) == 0 && dep.Name != "" {
			roots = append(roots, dep)
		}
	}

	if len(roots) == 1 {
		return roots[0]
	}

	root := &model.DepGraph{Path: f.Relpath()}
	for _, r := range roots {
		root.AppendChild(r)
	}
	return root
}

func ParseSpdxJson(f *model.File) *model.DepGraph {
	doc := &model.SpdxDocument{}
	f.OpenReader(func(reader io.Reader) {
		json.NewDecoder(reader).Decode(&doc)
	})
	return parseSpdxDoc(f, doc)
}

func ParseSpdxXml(f *model.File) *model.DepGraph {
	doc := &model.SpdxDocument{}
	f.OpenReader(func(reader io.Reader) {
		xml.NewDecoder(reader).Decode(&doc)
	})
	return parseSpdxDoc(f, doc)
}

func parseSpdxDoc(f *model.File, doc *model.SpdxDocument) *model.DepGraph {

	if doc == nil || doc.SPDXVersion == "" {
		return nil
	}

	depIdMap := map[string]*model.DepGraph{}
	_dep := model.NewDepGraphMap(func(s ...string) string {
		return s[0]
	}, func(s ...string) *model.DepGraph {
		return &model.DepGraph{
			Vendor:  s[1],
			Name:    s[2],
			Version: s[3],
		}
	})

	if len(doc.Packages) == 0 {
		return nil
	}

	for _, pkg := range doc.Packages {
		vendor := strings.TrimSpace(strings.TrimPrefix(pkg.Supplier, "Organization:"))
		dep := _dep.LoadOrStore(pkg.SPDXID, vendor, pkg.Name, pkg.Version)
		for _, ref := range pkg.ExternalRefs {
			if ref.ReferenceType == "purl" {
				_, _, _, dep.Language = model.ParsePurl(ref.ReferenceLocator)
				break
			}
		}
		depIdMap[pkg.SPDXID] = dep
	}

	for _, relation := range doc.Relationships {
		depIdMap[relation.SPDXElementID].AppendChild(depIdMap[relation.RelatedSPDXElement])
	}

	root := &model.DepGraph{Path: f.Relpath()}
	for _, dep := range depIdMap {
		if len(dep.Parents) == 0 {
			root.AppendChild(dep)
		}
	}

	return root
}
