//go:build !windows
// +build !windows

package image

import (
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/rancher/wharfie/pkg/extract"
)

// extractFiles was extracted as method to assist with differences in Windows.
func extractFiles(img v1.Image, dir string) error {
	return extract.Extract(img, dir)
}
