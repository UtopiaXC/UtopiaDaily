package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dashboard/dist all:newspaper/dist
var embeddedFS embed.FS

func DashboardFS() (fs.FS, error) {
	return fs.Sub(embeddedFS, "dashboard/dist")
}

func NewspaperFS() (fs.FS, error) {
	return fs.Sub(embeddedFS, "newspaper/dist")
}
