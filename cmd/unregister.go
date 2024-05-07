package cmd

import (
	"github.com/lcian/codeforces-cli/client"
)

// Unregister command
func Unregister() error {
	return client.Instance.Unregister(Args.Info.ContestID)
}
