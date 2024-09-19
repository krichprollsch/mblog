package gen

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	markdownExt = ".md"
	htmlExt     = ".html"
	homepage    = "index.html"
)

const (
	TmplIndex = "index.tmpl"
	TmplPost  = "post.tmpl"
	TmplPage  = "page.tmpl"
)

// Generator generates HTML files by converting markdown files from input dir
// into output dir.
type Generator struct {
	input  fs.FS
	output string

	templates fs.FS
}

func New(tmpl, input fs.FS, output string) *Generator {
	// TODO check dir
	return &Generator{
		input:  input,
		output: output,

		templates: tmpl,
	}
}

func parseTmpl(FS fs.FS) (map[string]*template.Template, error) {
	// Parse the templates.
	templates := make(map[string]*template.Template)
	for _, t := range []string{TmplIndex, TmplPost, TmplPage} {
		tmpl, err := template.New(t).ParseFS(FS, t)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", t, err)
		}

		templates[t] = tmpl
	}

	return templates, nil
}

// Run converts all markdown files in input into html files in output.
func (g *Generator) Run(ctx context.Context) error {
	// Parse templates.
	templates, err := parseTmpl(g.templates)
	if err != nil {
		return fmt.Errorf("templates: %w", err)
	}

	// Retrieve markdown files from input.
	files, err := markdown(g.input)
	if err != nil {
		return fmt.Errorf("markdown: %w", err)
	}

	var index *post

	metas := make([]metadata, 0, len(files))
	for _, f := range files {
		post, err := parsePost(g.input, f)
		if err != nil {
			return fmt.Errorf("html: %w", err)
		}

		// generate filename.
		outname := strings.Replace(f, markdownExt, htmlExt, 1)
		post.Meta.Href = outname

		if outname == homepage {
			index = post
			continue
		}

		// Create sub folders.
		if err := mkdir(g.output, outname); err != nil {
			return fmt.Errorf("mkdir: %w", err)
		}

		// Write html rendered file.
		fout, err := os.Create(filepath.Join(g.output, outname))
		if err != nil {
			return fmt.Errorf("create file: %w", err)
		}
		defer fout.Close()

		var tmpl *template.Template
		if t := post.Meta.Template; t != "" {
			// user defined template
			var ok bool
			tmpl, ok = templates[t]
			if !ok {
				// the template doesn't exists
				return fmt.Errorf("template not found: %s", t)
			}
		} else {
			// default template
			if post.Meta.IsPost() {
				tmpl = templates[TmplPost]
			} else {
				tmpl = templates[TmplPage]
			}
		}

		err = tmpl.Execute(fout, struct {
			Meta    metadata
			Content template.HTML
		}{
			Meta:    post.Meta,
			Content: template.HTML(post.Content),
		})
		metas = append(metas, post.Meta)

		if err != nil {
			return fmt.Errorf("write post: %w", err)
		}

		fout.Close()
	}

	// sort meta data by date desc.
	sort.Slice(metas, func(i, j int) bool { return metas[i].Date.After(metas[j].Date) })

	// Write homepage file.
	fout, err := os.Create(filepath.Join(g.output, homepage))
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	if index == nil {
		err = templates[TmplIndex].Execute(fout, metas)
		if err != nil {
			return fmt.Errorf("write homepage: %w", err)
		}

		return nil

	}

	tmpl, err := template.New("index.md").Parse(index.Content)
	if err != nil {
		return fmt.Errorf("parse index.md: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, metas)
	if err != nil {
		return fmt.Errorf("write homepage: %w", err)
	}

	t := TmplIndex
	if tt := index.Meta.Template; tt != "" {
		t = tt
	}

	err = templates[t].Execute(fout, struct {
		Meta    metadata
		Content template.HTML
	}{
		Meta:    index.Meta,
		Content: template.HTML(buf.String()),
	})
	if err != nil {
		return fmt.Errorf("write homepage: %w", err)
	}

	return nil
}

// mkdir generate directory tree.
func mkdir(root, filename string) error {
	dir, _ := filepath.Split(filename)
	dir = filepath.Join(root, dir)
	if err := os.MkdirAll(dir, 0744); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	return nil
}
