package ports

import (
	"context"
)

// AkademikProvisioner is kesantrian's own port for everything it needs from
// the akademik module: resolving the default program (for manual/user-request
// santri creation) and persisting the santri→program mapping. Defined here
// (not as a dependency on akademik's package) to keep module boundaries
// clean and avoid an import cycle, since akademik already depends on
// kesantrian. See docs/architecture/module-boundaries.md.
type AkademikProvisioner interface {
	GetDefaultProgramID(ctx context.Context) (*string, error)
	AssignSantriProgram(ctx context.Context, santriID, programID string) error
}
