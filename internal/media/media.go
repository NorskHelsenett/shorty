package media

import (
	"embed"
	"os"

	"github.com/NorskHelsenett/ror/pkg/rlog"
)

//go:embed nhnlogo.png
var media embed.FS

var ImageFile string

func Load() {
	data, _ := media.ReadFile("nhnlogo.png")
	//write data to a temp file
	tmpfile, err := os.CreateTemp("", "nhnlogo-*.png")
	if err != nil {
		rlog.Error("Failed to create temp file for logo", err)
	} else {
		if _, err := tmpfile.Write(data); err != nil {
			rlog.Error("Failed to write to temp logo file", err)
		}
		if err := tmpfile.Close(); err != nil {
			rlog.Error("Failed to close temp logo file", err)
		}
		ImageFile = tmpfile.Name()
	}
}

func Unload() {
	if ImageFile != "" {
		if err := os.Remove(ImageFile); err != nil {
			rlog.Error("Failed to remove temp logo file", err)
		}
	}
}
