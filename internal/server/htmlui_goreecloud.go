//go:build !nohtmlui

package server

import (
	"bytes"
	"embed"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/pkg/errors"
)

const (
	goreeCloudUIAssetPrefix = "goreecloud-ui/"
	goreeCloudHeadMarker     = "<!-- goreecloud-backup-ui -->"
)

//go:embed goreecloud-ui/*
var goreeCloudUIAssets embed.FS

type goreeCloudAssetFileSystem struct {
	base http.FileSystem
}

func newGoreeCloudAssetFile(base http.FileSystem) http.FileSystem {
	return &goreeCloudAssetFileSystem{base: base}
}

func (f *goreeCloudAssetFileSystem) Open(name string) (http.File, error) {
	cleanName := strings.TrimPrefix(path.Clean("/"+name), "/")

	if strings.HasPrefix(cleanName, goreeCloudUIAssetPrefix) {
		localFile, err := http.FS(goreeCloudUIAssets).Open(cleanName)
		if err != nil {
			return nil, errors.Wrap(err, "open GoreeCloud UI asset")
		}

		return localFile, nil
	}

	upstreamFile, err := f.base.Open(name)
	if err != nil {
		return nil, errors.Wrap(err, "open upstream HTML UI asset")
	}

	if cleanName != "index.html" {
		return upstreamFile, nil
	}

	defer func() {
		_ = upstreamFile.Close()
	}()

	info, err := upstreamFile.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "stat upstream HTML UI index")
	}

	contents, err := io.ReadAll(upstreamFile)
	if err != nil {
		return nil, errors.Wrap(err, "read upstream HTML UI index")
	}

	contents = applyGoreeCloudHTMLIdentity(contents)

	return &memoryHTTPFile{
		Reader: bytes.NewReader(contents),
		info:   goreeCloudFileInfo{FileInfo: info, size: int64(len(contents))},
	}, nil
}

func applyGoreeCloudHTMLIdentity(contents []byte) []byte {
	html := string(contents)
	if strings.Contains(html, goreeCloudHeadMarker) {
		return contents
	}

	html = replaceHTMLTagContents(html, "title", "GoreeCloud Backup")
	html = strings.Replace(
		html,
		`<meta name="description" content="Kopia UI" />`,
		`<meta name="description" content="GoreeCloud Backup — private, encrypted, recoverable backup and restore." />`,
		1,
	)
	html = strings.Replace(
		html,
		`<body id="kopia">`,
		`<body id="kopia" class="glaze-canvas goreecloud-backup">`,
		1,
	)

	head := goreeCloudHeadMarker + `
    <meta name="application-name" content="GoreeCloud Backup" />
    <meta name="theme-color" content="#172033" media="(prefers-color-scheme: light)" />
    <meta name="theme-color" content="#0d1119" media="(prefers-color-scheme: dark)" />
    <link rel="stylesheet" href="/goreecloud-ui/glaze.css" />
    <link rel="stylesheet" href="/goreecloud-ui/glaze.accessibility.css" />
    <link rel="stylesheet" href="/goreecloud-ui/backup.css" />`

	html = strings.Replace(html, "</head>", head+"\n  </head>", 1)

	return []byte(html)
}

func replaceHTMLTagContents(html, tag, replacement string) string {
	openPrefix := "<" + tag

	openStart := strings.Index(html, openPrefix)
	if openStart < 0 {
		return html
	}

	openEndRelative := strings.Index(html[openStart:], ">")
	if openEndRelative < 0 {
		return html
	}

	openEnd := openStart + openEndRelative + 1
	closeTag := "</" + tag + ">"

	closeRelative := strings.Index(html[openEnd:], closeTag)
	if closeRelative < 0 {
		return html
	}

	closeStart := openEnd + closeRelative

	return html[:openEnd] + replacement + html[closeStart:]
}

type memoryHTTPFile struct {
	*bytes.Reader
	info fs.FileInfo
}

func (f *memoryHTTPFile) Close() error { return nil }

func (f *memoryHTTPFile) Stat() (fs.FileInfo, error) {
	return f.info, nil
}

func (f *memoryHTTPFile) Readdir(_ int) ([]fs.FileInfo, error) {
	return nil, &fs.PathError{Op: "readdir", Path: f.info.Name(), Err: fs.ErrInvalid}
}

type goreeCloudFileInfo struct {
	fs.FileInfo
	size int64
}

func (i goreeCloudFileInfo) Size() int64 { return i.size }
