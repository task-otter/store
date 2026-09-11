// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"errors"
)

var errRepoMissing = errors.New("could not find repository root with go.mod")
