package runtime

import "fmt"

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type DebugMetadata struct {
	StartLineNumber      int
	EndLineNumber        int
	StartCharacterNumber int
	EndCharacterNumber   int
	FileName             string
	SourceName           string
}

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
