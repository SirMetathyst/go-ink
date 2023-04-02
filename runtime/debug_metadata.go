package runtime

import (
	"fmt"
	"math"
)

type DebugMetadata struct {
	StartLineNumber      int
	EndLineNumber        int
	StartCharacterNumber int
	EndCharacterNumber   int
	FileName             string
	SourceName           string
}

func (s *DebugMetadata) NewDebugMetadata() *DebugMetadata {
	return new(DebugMetadata)
}

// Merge Currently only used in VariableReference in order to
// merge the debug metadata of a Path.Of.Indentifiers into
// one single range.
func (s *DebugMetadata) Merge(dm *DebugMetadata) *DebugMetadata {

	newDebugMetadata := new(DebugMetadata)

	// These are not supposed to be differ between 'this' and 'dm'.
	newDebugMetadata.FileName = s.FileName
	newDebugMetadata.SourceName = s.SourceName

	if s.StartLineNumber < dm.StartLineNumber {

		newDebugMetadata.StartLineNumber = s.StartLineNumber
		newDebugMetadata.StartCharacterNumber = s.StartCharacterNumber
	} else if s.StartLineNumber > dm.StartLineNumber {

		newDebugMetadata.StartLineNumber = dm.StartLineNumber
		newDebugMetadata.StartCharacterNumber = dm.StartCharacterNumber
	} else {

		newDebugMetadata.StartLineNumber = s.StartLineNumber
		newDebugMetadata.StartCharacterNumber = int(math.Min(float64(s.StartCharacterNumber), float64(dm.StartCharacterNumber)))
	}

	if s.EndLineNumber > dm.EndLineNumber {

		newDebugMetadata.EndLineNumber = s.EndLineNumber
		newDebugMetadata.EndCharacterNumber = s.EndCharacterNumber
	} else if s.EndLineNumber < dm.EndLineNumber {

		newDebugMetadata.EndLineNumber = dm.EndLineNumber
		newDebugMetadata.EndCharacterNumber = dm.EndCharacterNumber
	} else {

		newDebugMetadata.EndLineNumber = s.EndLineNumber
		newDebugMetadata.EndCharacterNumber = int(math.Min(float64(s.EndCharacterNumber), float64(dm.EndCharacterNumber)))
	}

	return newDebugMetadata
}

func (s *DebugMetadata) String() string {

	if s.FileName != "" {
		return fmt.Sprintf("line %d of %s", s.StartLineNumber, s.FileName)
	}

	return fmt.Sprintf("line %d", +s.StartLineNumber)
}
