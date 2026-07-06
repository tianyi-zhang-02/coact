package cli

import (
	"fmt"
	"runtime"

	"github.com/coactdev/coact/internal/buildinfo"
)

func cmdVersion() int {
	fmt.Printf("coact %s (%s) %s/%s\n",
		buildinfo.Version, buildinfo.Commit, runtime.GOOS, runtime.GOARCH)
	return 0
}
