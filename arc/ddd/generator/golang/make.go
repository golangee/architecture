package golang

import (
	"fmt"
	"github.com/golangee/architecture/arc/adl"
	"github.com/golangee/architecture/arc/generator/astutil"
	"github.com/golangee/architecture/arc/generator/doc"
	"github.com/golangee/architecture/arc/generator/golang"
	"github.com/golangee/architecture/arc/generator/stereotype"
	"github.com/golangee/src/ast"
	"strings"
)

func renderGolangciyaml(dst *ast.Mod, src *adl.Module) error {
	const cfg = `
run:
  timeout: '5m'

  build-tags:
  - 'all'

  skip-dirs:

  skip-dirs-use-default: false

  modules-download-mode: 'readonly'

  allow-parallel-runners: true

linters:
  enable:
  - 'asciicheck'
  - 'bodyclose'
  - 'deadcode'
  - 'depguard'
  - 'dogsled'
  - 'errcheck'
  - 'errorlint'
  - 'exportloopref'
  - 'gofmt'
  - 'gofumpt'
  - 'goheader'
  - 'goimports'
  - 'revive'
  - 'gomodguard'
  - 'goprintffuncname'
  - 'gosec'
  - 'gosimple'
  - 'govet'
  - 'ineffassign'
  - 'makezero'
  - 'misspell'
  - 'noctx'
  - 'paralleltest'
  - 'prealloc'
  - 'predeclared'
  - 'sqlclosecheck'
  - 'staticcheck'
  - 'structcheck'
  - 'stylecheck'
  - 'typecheck'
  - 'unconvert'
  - 'unused'
  - 'varcheck'
  - 'whitespace'
  - 'wsl'

issues:
  exclude:
  - '^SA3000:' # staticcheck: redundant, fixed in #34129
  - '^Range statement' # paralleltest: usually false positives

  max-issues-per-linter: 0

  max-same-issues: 0

severity:
  default-severity: error
`

	pkg := astutil.MkPkg(dst, golang.MakePkgPath(dst.Name))
	pkg.AddRawFiles(ast.NewRawTpl("golangci.yaml", "text/x-yaml", ast.NewTpl(
		makeEscapedPreamble(src.Preamble, "# ")+cfg,
	)))

	return nil
}

func renderMakefile(dst *ast.Mod, src *adl.Module) error {
	const mkfile = `

prefix ?= /usr/local
bindir ?= $(prefix)/bin
DESTDIR ?= ./build
TMPDIR ?= $(shell mktemp -d)

CI_JOB_ID ?= $(shell date +%s)
CI_COMMIT_TAG ?= $(shell git name-rev --name-only HEAD)
CI_JOB_STARTED_AT ?= $(shell date +"%Y-%m-%dT%T%z") # RFC3339 | ISO8601
CI_COMMIT_SHA ?= $(shell git rev-parse HEAD)
CI_SERVER_HOST ?= $(shell hostname)

UNAME := $(shell uname)
ifeq ($(UNAME),Darwin)
	openHugoCMD := open http://localhost:4563
endif

buildInfo = {{.Get "modPath"}}/internal/buildinfo
LDFLAGS = -X $(buildInfo).JobID=${CI_JOB_ID} -X $(buildInfo).CommitTag=${CI_COMMIT_TAG} -X $(buildInfo).JobStartedAt=${CI_JOB_STARTED_AT} -X $(buildInfo).CommitSha=${CI_COMMIT_SHA} -X $(buildInfo).Host=${CI_SERVER_HOST}

# doc: #go install --tags=extended github.com/gohugoio/hugo@latest 

{{- range .Get "install"}}
{{.VarName}} = "{{.Path}}"
{{- end}}

test: ## Executes all tests.
	@go test -race -timeout=5m ./...
.PHONY: test

lint: ## Runs the linter. See golangci.yaml for details.
	@command -v golangci-lint > /dev/null 2>&1 || (cd $${TMPDIR} && go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.40.1)
	golangci-lint run --config golangci.yaml
.PHONY: lint

generate: ## runs go generate
	@go generate ./...
.PHONY: generate

help: ## Shows this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.PHONY: help

{{ range .Get "install"}}
run-{{.Foldername}}: ## runs {{.Foldername}} without installing
	go run -ldflags "${LDFLAGS}" $({{.VarName}})
{{ end}}

# required by GNU standard

check: lint test ## build and runs tests and linters

install: ## builds and installs on the current system
{{- range .Get "install"}}
	go build -ldflags "${LDFLAGS}" -o $(DESTDIR)$(bindir)/{{.Filename}} $({{.VarName}})
{{- end}}
.PHONY: dist

dist: ## builds for all major platforms, example: make DESTDIR=/tmp/stage dist
{{- range .Get "dists"}}
	GOOS={{.Os}} GOARCH={{.Arch}} go build -ldflags "${LDFLAGS}" -o $(DESTDIR)/bin/{{.Os}}_{{.Arch}}/{{.Filename}} $({{.VarName}})
{{- end}}
.PHONY: dist


# download hugo theme only once
$(DESTDIR)/docs/html/themes/hugo-book/:
	mkdir -p $(DESTDIR)/docs/html/themes
	curl https://codeload.github.com/alex-shpak/hugo-book/zip/refs/heads/master --output $(DESTDIR)/docs/html/hugo-book.zip
	unzip $(DESTDIR)/docs/html/hugo-book.zip -d $(DESTDIR)/docs/html/themes/hugo-book
	mv $(DESTDIR)/docs/html/themes/hugo-book/hugo-book-master/* $(DESTDIR)/docs/html/themes/hugo-book
	rm -f $(DESTDIR)/docs/html/hugo-book.zip
	rm -rf $(DESTDIR)/docs/html/themes/hugo-book/hugo-book-master
	rm -rf $(DESTDIR)/docs/html/themes/hugo-book/images
	rm -rf $(DESTDIR)/docs/html/themes/hugo-book/exampleSite
	rm -f $(DESTDIR)/docs/html/themes/hugo-book/README.md

html: $(DESTDIR)/docs/html/themes/hugo-book/ ## generates the html documentation.
	rm -rf $(DESTDIR)/docs/html/content
	cp -R docs/ $(DESTDIR)/docs/html/
	cd $(DESTDIR)/docs/html/ && hugo 
.PHONY: html

serve-html: html ## builds the html documentation, serves it and opens a browser window
	$(openHugoCMD)
	cd $(DESTDIR)/docs/html/ && hugo server -D -p 4563
.PHONY: serve-html

clean:
	rm -rf $(DESTDIR)/docs
	rm -rf $(DESTDIR)/bin
.PHONY: clean

all: generate check dist html ## generate, check and build dist
.PHONY: all

.DEFAULT_GOAL := all
`

	fnameFunc := func(os, arch, path string) string {
		name := astutil.LastPathSegment(path)
		if os == "js" && arch == "wasm" {
			name += ".wasm"
		}

		if os == "windows" {
			name += ".exe"
		}

		return name
	}

	var targets []makeTarget
	var currentOsTargets []makeTarget
	for _, cmd := range stereotype.FindCMDPkgs(dst) {
		currentOsTargets = append(currentOsTargets, makeTarget{
			Path:       cmd.Path,
			Filename:   fnameFunc("", "", cmd.Path),
			Foldername: astutil.LastPathSegment(cmd.Path),
		})

		for _, dist := range src.Generator.Go.GoDist {
			targets = append(targets, makeTarget{
				Arch:       dist.Arch.String(),
				Os:         dist.Os.String(),
				Path:       cmd.Path,
				Foldername: astutil.LastPathSegment(cmd.Path),
				Filename:   fnameFunc(strings.ToLower(dist.Os.String()), strings.ToLower(dist.Arch.String()), cmd.Path),
			})
		}
	}

	pkg := astutil.MkPkg(dst, golang.MakePkgPath(dst.Name))
	pkg.AddRawFiles(
		ast.NewRawTpl("Makefile", "text/x-makefile", ast.NewTpl(
			makeEscapedPreamble(src.Preamble, "# ")+mkfile,
		).
			Put("dists", targets).
			Put("install", currentOsTargets).
			Put("modPath", dst.Name)),
	)

	if err := renderGolangciyaml(dst, src); err != nil {
		return fmt.Errorf("cannot create golangci.yml: %w", err)
	}

	stereotype.Doc(dst, "", stereotype.DocDevelop,
		doc.NewElement("div").Append(
			doc.NewElement("h1").Append(doc.NewText("how to use the makefile")),
			doc.NewText(""),
		))

	stereotype.Doc(dst, "de", stereotype.DocDevelop,
		doc.NewElement("div").Append(
			doc.NewElement("h1").Append(doc.NewText("so geht es: makefile")),
			doc.NewText(""),
		))

	return nil
}

type makeTarget struct {
	Arch       string
	Os         string
	Path       string
	Filename   string
	Foldername string
}

func (m makeTarget) VarName() string {
	return "path_" + strings.ReplaceAll(m.Foldername, "-", "_")
}
