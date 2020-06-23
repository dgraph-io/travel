// Package commands contains the functionality for the set of commands
// currently supported by the CLI tooling.
package commands

import "errors"

// ErrHelp provides context that help was given.
var ErrHelp = errors.New("provided help")
