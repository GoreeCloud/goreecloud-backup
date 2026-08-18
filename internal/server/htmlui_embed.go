//go:build !nohtmlui

package server

import (
	"net/http"

	"github.com/kopia/htmluibuild"
)

// AssetFile exposes the upstream HTML UI through the GoreeCloud presentation overlay.
//
// The overlay is intentionally presentation-only: it delegates all upstream assets and
// application behavior unchanged while adding GoreeCloud Backup identity and local Glaze UI
// resources at the HTTP file-system boundary.
func AssetFile() http.FileSystem {
	return newGoreeCloudAssetFile(htmluibuild.AssetFile())
}
