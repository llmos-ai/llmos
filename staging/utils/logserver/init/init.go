package init

import "github.com/llmos-ai/llmos/utils/logserver"

func init() {
	go logserver.StartServerWithDefaults()
}
