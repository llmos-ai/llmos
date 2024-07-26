//go:build windows
// +build windows

package image

import (
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/rancher/wharfie/pkg/extract"
)

// extractFiles was abstracted because Windows containers don't have a concept of a scratch container.
// So files need to be placed in a subdirectory.
func extractFiles(img v1.Image, dir string) error {
	extractPaths := map[string]string{
		"/Files/bin": dir,
	}
	return extract.ExtractDirs(img, extractPaths)
}
