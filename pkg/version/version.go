package version

import (
	"fmt"
	"runtime/debug"
)

func Version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	ver := info.Main.Version
	if ver == "" {
		ver = info.Main.Sum
	}
	if ver == "" {
		ver = "undefined"
	}
	return fmt.Sprintf("%s build with %s", ver, info.GoVersion)
}
