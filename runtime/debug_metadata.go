package ink

import "fmt"

// Required because std math does not yet provide a generic min function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Required because std math does not yet provide a generic max function
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Source: https://github.com/inkle/ink/blob/master/ink-engine-runtime/DebugMetadata.cs

type DebugMetadata struct {
	StartLineNumber      int    // source: public, default value: 0
	EndLineNumber        int    // source: public, default value: 0
	StartCharacterNumber int    // source: public, default value: 0
	EndCharacterNumber   int    // source: public, default value: 0
	FileName             string // source: public, default value: null
	SourceName           string // source: public, default value: null
}

// Source: https://github.com/inkle/ink/blob/master/ink-engine-runtime/DebugMetadata.cs#L21

// Merge ...
func (s *DebugMetadata) Merge(dm *DebugMetadata) *DebugMetadata {
	newDebugMetadata := new(DebugMetadata)

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
		newDebugMetadata.StartCharacterNumber = min(s.StartCharacterNumber, dm.StartCharacterNumber)
	}

	if s.EndLineNumber > dm.EndLineNumber {
		newDebugMetadata.EndLineNumber = s.EndLineNumber
		newDebugMetadata.EndCharacterNumber = s.EndCharacterNumber
	} else if s.EndLineNumber < dm.EndLineNumber {
		newDebugMetadata.EndLineNumber = dm.EndLineNumber
		newDebugMetadata.EndCharacterNumber = dm.EndCharacterNumber
	} else {
		newDebugMetadata.EndLineNumber = s.EndLineNumber
		newDebugMetadata.EndCharacterNumber = max(s.EndCharacterNumber, dm.EndCharacterNumber)
	}

	return newDebugMetadata
}

func (s *DebugMetadata) String() string {
	if s.FileName != "" {
		return fmt.Sprintf("line %d of %s", s.StartLineNumber, s.FileName)
	} else {
		return fmt.Sprintf("line %d", s.StartLineNumber)
	}
}
