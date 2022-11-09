package gen

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday/v2"
)

const (
	markdownExt = ".md"
	htmlExt     = ".html"
)

// Generator generates HTML files by converting markdown files from input dir
// into output dir.
type Generator struct {
	input  fs.FS
	output string
}

func New(input fs.FS, output string) *Generator {
	// TODO check dir
	return &Generator{
		input:  input,
		output: output,
	}
}

// Run converts all markdown files in input into html files in output.
func (g *Generator) Run(ctx context.Context) error {
	// Retrieve markdown files from input.
	files, err := markdown(g.input)
	if err != nil {
		return fmt.Errorf("markdown: %w", err)
	}

	for _, f := range files {
		b, err := html(g.input, f)
		if err != nil {
			return fmt.Errorf("html: %w", err)
		}

		// generate filename.
		outname := strings.Replace(f, markdownExt, htmlExt, 1)

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

		if _, err := fout.Write(b); err != nil {
			return fmt.Errorf("write file: %w", err)
		}

		fout.Close()
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

// markdown returns all markdown files from infs.
func markdown(infs fs.FS) ([]string, error) {
	var files []string

	err := fs.WalkDir(infs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(d.Name()) != markdownExt {
			return nil
		}

		files = append(files, path)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walkdir in: %w", err)
	}

	return files, err
}

// html converts io.Reader containing markdown into html data.
func html(infs fs.FS, file string) ([]byte, error) {
	f, err := infs.Open(file)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read reader: %w", err)
	}

	return blackfriday.Run(b), nil
}
