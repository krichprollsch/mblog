package gen

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	markdownExt = ".md"
	htmlExt     = ".html"
	homepage    = "index.html"
)

const (
	TmplHome = "home.tmpl"
	TmplPost = "post.tmpl"
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
	for _, t := range []string{TmplHome, TmplPost} {
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

	metas := make([]metadata, len(files))
	for i, f := range files {
		post, err := parsePost(g.input, f)
		if err != nil {
			return fmt.Errorf("html: %w", err)
		}

		// generate filename.
		outname := strings.Replace(f, markdownExt, htmlExt, 1)
		post.Meta.Href = outname

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

		err = templates[TmplPost].Execute(fout, struct {
			Meta    metadata
			Content template.HTML
		}{
			Meta:    post.Meta,
			Content: template.HTML(post.Content),
		})
		metas[i] = post.Meta

		if err != nil {
			return fmt.Errorf("write post: %w", err)
		}

		fout.Close()
	}

	// Write homepage file.
	fout, err := os.Create(filepath.Join(g.output, homepage))
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	err = templates[TmplHome].Execute(fout, metas)
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
