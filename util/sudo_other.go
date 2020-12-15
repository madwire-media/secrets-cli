// +build !android,!linux,!darwin,!windows

package util

import "errors"

func callSudo(action string, params []string) error {
	return errors.New("privilege escalation not supported on this platform")
}
