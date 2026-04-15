package main

import (
	"embed"
	"os"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// WebKitGTK's Wayland path is fragile across distros/compositors and
	// frequently fails with "Protocol error dispatching to Wayland display".
	// Forcing the X11 backend (via XWayland) is the upstream-recommended
	// workaround and is transparent to the user. Must be set before GTK
	// initialises, so this runs before wails.Run.
	if runtime.GOOS == "linux" {
		if _, set := os.LookupEnv("GDK_BACKEND"); !set {
			os.Setenv("GDK_BACKEND", "x11")
		}
		if _, set := os.LookupEnv("WEBKIT_DISABLE_DMABUF_RENDERER"); !set {
			os.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "1")
		}
	}

	app := NewApp()

	err := wails.Run(&options.App{
		Title:            "Genetica Resolutio",
		Width:            1400,
		Height:           900,
		MinWidth:         900,
		MinHeight:        600,
		BackgroundColour: &options.RGBA{R: 8, G: 12, B: 10, A: 255},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
		// Disable frame dragging via --webkit-app-region so our custom title bar works
		Frameless:        false,
		EnableDefaultContextMenu: false,
		Windows: &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			DisableWindowIcon:                 false,
			IsZoomControlEnabled:              false,
			DisablePinchZoom:                  true,
			Theme:                             windows.Dark,
		},
		Mac: &mac.Options{
			TitleBar: mac.TitleBarHiddenInset(),
			Appearance: mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
		},
		Linux: &linux.Options{
			Icon:                []byte{},
			WindowIsTranslucent: false,
			WebviewGpuPolicy:    linux.WebviewGpuPolicyAlways,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
