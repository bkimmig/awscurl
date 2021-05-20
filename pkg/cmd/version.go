package cmd

import (
	"fmt"
)

var (
	version = "0.0.1b"
	commit  = "none"
)

func getVersion() string {
	return fmt.Sprintf("%s, build %s", version, commit)
}
