package model

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"time"

	"github.com/ArminasAer/aerlon/internal/database"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/util"
)

type Post struct {
	ID          uuid.UUID      `json:"id"`
	Title       string         `json:"title"`
	Date        time.Time      `json:"date"`
	Slug        string         `json:"slug"`
	Series      string         `json:"series"`
	Categories  pq.StringArray `json:"categories"`
	Markdown    string         `json:"markdown"`
	Published   bool           `json:"published"`
	Featured    bool           `json:"featured"`
	PostSnippet string         `db:"post_snippet" json:"post_snippet"`
	CreatedAt   string         `db:"created_at" json:"created_at"`
	UpdatedAt   string         `db:"updated_at" json:"updated_at"`
}

var md = goldmark.New(
	goldmark.WithExtensions(
		meta.Meta,
		extension.GFM,
		highlighting.NewHighlighting(
			highlighting.WithWrapperRenderer(func(w util.BufWriter, context highlighting.CodeBlockContext, entering bool) {
				lang, _ := context.Language()

				if entering {
					if lang == nil {
						w.WriteString("<pre><code>")
						return
					}
					w.WriteString(fmt.Sprintf(`<div class="code-block"><span class="language-name">%s</span>`, lang))
				} else {
					if lang == nil {
						w.WriteString("</pre></code>")
						return
					}
					w.WriteString(`</code></pre></div>`)
				}
			}),
			highlighting.WithFormatOptions(
				chromahtml.WithLineNumbers(true),
				// chromahtml.PreventSurroundingPre(true),
				chromahtml.WithClasses(true),
			),
		),
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
)

func nodeParser(value string) (string, error) {
	cmd := exec.Command("node", "./parser/parser-compiled.js", value)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return "", err
	}

	outStr, errStr := stdout.String(), stderr.String()
	if errStr != "" {
		return "", errors.New(errStr)
	}

	return outStr, nil
}

func (p *Post) ConvertMarkdownToHTML() error {
	// var buf bytes.Buffer
	// err := md.Convert([]byte(p.Markdown), &buf)
	// if err != nil {
	// 	return err
	// }

	mrk, err := nodeParser(p.Markdown)
	if err != nil {
		println("HERE>>>>>")
		return err
	}

	// p.Markdown = buf.String()
	p.Markdown = mrk

	return nil
}

func SortPostsByDate(posts []*Post) {
	sort.Slice(posts, func(i, j int) bool {
		a := posts[i].Date.Unix()
		b := posts[j].Date.Unix()

		if a == b {
			return posts[i].Title < posts[j].Title
		} else if a > b {
			return a > b
		}
		return b < a
	})
}

func GetPostFromDB(DB *database.DBPool, id uuid.UUID) (*Post, error) {
	var post *Post
	err := DB.Get(&post, "SELECT * FROM post WHERE id = $1", id)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func GetPostsFromDB(DB *database.DBPool) ([]*Post, error) {
	var posts []*Post
	err := DB.Select(&posts, "SELECT * FROM post")
	if err != nil {
		return nil, err
	}

	return posts, nil
}
