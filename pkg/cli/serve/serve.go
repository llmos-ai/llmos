package serve

import (
	"context"
	//"github.com/containerd/nerdctl/pkg/clientutil"
)

const (
	OpenWebUIImage = "ghcr.io/open-webui/open-webui:main"
	WebUIName      = "open-webui"
)

type ServeOptions struct {
	Address   string
	Model     string
	Namespace string
	Platform  string
}

func NewServe() ServeOptions {
	return ServeOptions{}
}

func (s ServeOptions) StartServe(ctx context.Context) error {

	<-ctx.Done()
	return ctx.Err()
}
