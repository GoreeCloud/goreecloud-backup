//go:build !nohtmlui

package server

import (
	"io"
	"io/fs"
	"net/http"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestGoreeCloudAssetFileTransformsIndex(t *testing.T) {
	base := http.FS(fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte(`<!doctype html><html><head><meta name="description" content="Kopia UI" /><title>KopiaUI</title></head><body id="kopia"><div id="root"></div></body></html>`)},
		"asset.txt":  &fstest.MapFile{Data: []byte("upstream")},
	})

	ui := newGoreeCloudAssetFile(base)
	f, err := ui.Open("index.html")
	require.NoError(t, err)

	data, err := io.ReadAll(f)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	html := string(data)

	require.Contains(t, html, "<title>GoreeCloud Backup</title>")
	require.Contains(t, html, `name="application-name" content="GoreeCloud Backup"`)
	require.Contains(t, html, `class="glaze-canvas goreecloud-backup"`)
	require.Contains(t, html, `/goreecloud-ui/glaze.css`)
	require.Contains(t, html, `/goreecloud-ui/glaze.accessibility.css`)
	require.Contains(t, html, `/goreecloud-ui/backup.css`)
	require.NotContains(t, html, "<title>KopiaUI</title>")
	require.Equal(t, 1, strings.Count(html, goreeCloudHeadMarker))
}

func TestGoreeCloudAssetFileDelegatesUpstreamAssets(t *testing.T) {
	base := http.FS(fstest.MapFS{
		"asset.txt": &fstest.MapFile{Data: []byte("upstream")},
	})

	f, err := newGoreeCloudAssetFile(base).Open("asset.txt")
	require.NoError(t, err)

	data, err := io.ReadAll(f)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	require.Equal(t, "upstream", string(data))
}

func TestGoreeCloudAssetFileServesLocalGlazeAssets(t *testing.T) {
	base := http.FS(fstest.MapFS{})
	ui := newGoreeCloudAssetFile(base)

	for _, name := range []string{
		"goreecloud-ui/glaze.css",
		"goreecloud-ui/glaze.accessibility.css",
		"goreecloud-ui/backup.css",
	} {
		f, err := ui.Open(name)
		require.NoError(t, err, name)

		data, readErr := io.ReadAll(f)
		require.NoError(t, readErr, name)
		require.NotEmpty(t, data, name)
		require.NoError(t, f.Close(), name)
	}
}

func TestApplyGoreeCloudHTMLIdentityIsIdempotent(t *testing.T) {
	input := []byte(`<html><head><title>KopiaUI</title></head><body id="kopia"></body></html>`)
	once := applyGoreeCloudHTMLIdentity(input)
	twice := applyGoreeCloudHTMLIdentity(once)

	require.Equal(t, string(once), string(twice))
	require.Equal(t, 1, strings.Count(string(twice), goreeCloudHeadMarker))
}

func TestGoreeCloudAssetFilePreservesMissingAssetError(t *testing.T) {
	base := http.FS(fstest.MapFS{})
	_, err := newGoreeCloudAssetFile(base).Open("missing.js")
	require.Error(t, err)
	require.ErrorIs(t, err, fs.ErrNotExist)
}
