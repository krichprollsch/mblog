package gen

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday/v2"
)

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

type post struct {
	Meta    metadata
	Content string
}

// parsePost converts io.Reader containing markdown into post struct.
func parsePost(infs fs.FS, file string) (*post, error) {

	f, err := infs.Open(file)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read reader: %w", err)
	}

	mdrenderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.CommonHTMLFlags,
	})
	mdparser := blackfriday.New(
		blackfriday.WithExtensions(blackfriday.CommonExtensions),
	)

	var meta metadata
	ast := mdparser.Parse(b)
	var buf bytes.Buffer
	var werr error
	ast.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		// check metadata
		if meta.Title == "" && isTitle(node, entering) {
			meta.Title = flat(node)
		}
		if isMeta(node, entering) {
			// extract meta data
			if err := meta.UnmarshalText(node.Literal); err != nil {
				werr = fmt.Errorf("set metadata: %w", err)
				return blackfriday.Terminate
			}

			// skip code rendering
			return blackfriday.SkipChildren
		}
		return mdrenderer.RenderNode(&buf, node, entering)
	})
	if werr != nil {
		return nil, err
	}

	return &post{
		Meta:    meta,
		Content: buf.String(),
	}, nil
}

// isTitle returns true if a node is a title level 1.
func isTitle(node *blackfriday.Node, entering bool) bool {
	return node.Type == blackfriday.Heading &&
		node.HeadingData.Level == 1 &&
		entering
}

// isMeta returns true if the node contains meta data code block.
func isMeta(node *blackfriday.Node, entering bool) bool {
	return node.Type == blackfriday.CodeBlock &&
		node.Prev == nil &&
		node.CodeBlockData.IsFenced &&
		string(node.CodeBlockData.Info) == "meta" &&
		entering
}

// flat remove ast node hierachy an returns the corresponding plain text.
func flat(ast *blackfriday.Node) string {
	var b strings.Builder
	ast.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if node.Type == blackfriday.Text {
			b.Write(node.Literal)
		}

		return blackfriday.GoToNext
	})

	return b.String()
}
