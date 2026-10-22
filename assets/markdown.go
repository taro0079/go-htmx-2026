package assets

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
)

func ToHtml(source []byte) (string, error) {
	var buf bytes.Buffer

	if err := md.Convert(source, &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func RenderFile(path string) (string, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return ToHtml(source)
}

type Post struct {
	Title   string `yaml:"title"`
	Content template.HTML
}

func SearchPost(query string) ([]Post, error) {
	filtered := []Post{}
	posts, _ := LoadAll()

	if query == "" {
		return posts, nil
	}

	for _, post := range posts {
		if strings.Contains(post.Title, query) || strings.Contains(string(post.Content), query) {
			filtered = append(filtered, post)
		}
	}

	return posts, nil
}

func GetPost(slug string) (Post, error) {
	filePath := "posts/" + slug + ".md"
	data, err := fs.ReadFile(Posts, filePath)

	if err != nil {
		return Post{}, fmt.Errorf("%s: %w", filePath, err)
	}

	var post Post

	content, err := frontmatter.Parse(
		strings.NewReader(string(data)),
		&post,
	)

	if err != nil {
		return Post{}, fmt.Errorf("%s: %w", filePath, err)
	}

	html, err := ToHtml(content)
	if err != nil {
		return Post{}, fmt.Errorf("%s: %w", filePath, err)
	}

	post.Content = template.HTML(string(html))

	return post, nil
}

func LoadAll() ([]Post, error) {
	var posts []Post

	err := fs.WalkDir(Posts, ".", func(
		filePath string,
		d fs.DirEntry,
		err error,
	) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if strings.ToLower(path.Ext(filePath)) != ".md" {
			return nil
		}

		data, err := fs.ReadFile(Posts, filePath)

		if err != nil {
			return fmt.Errorf("%s: %w", filePath, err)
		}

		var post Post

		content, err := frontmatter.Parse(strings.NewReader(string(data)), &post)
		if err != nil {
			return fmt.Errorf("%s: %w", filePath, err)
		}

		html, err := ToHtml(content)
		if err != nil {
			return fmt.Errorf("%s: %w", filePath, err)
		}
		post.Content = template.HTML(string(html))

		posts = append(posts, post)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return posts, nil
}
