package cmd

import (
	"github.com/lcian/codeforces-cli/client"
)

// Register command
func Register() error {
	return client.Instance.Register(Args.Info.ContestID)
}
