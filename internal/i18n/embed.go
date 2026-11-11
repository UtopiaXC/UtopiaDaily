package i18n

import (
	"embed"
	"io/fs"
)

//go:embed locales/*.json
var embeddedFS embed.FS

func BaseLocalesFS() (fs.FS, error) {
	return fs.Sub(embeddedFS, "locales")
}
