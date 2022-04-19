package runtime

import "github.com/stretchr/testify/assert"
import "testing"

func TestPaths(t *testing.T) {

	var path1 = NewPathFromComponentString("hello.1.world")
	var path2 = NewPathFromComponentString("hello.1.world")
	var path3 = NewPathFromComponentString(".hello.1.world")
	var path4 = NewPathFromComponentString(".hello.1.world")

	assert.Equal(t, path1.String(), path2.String())
	assert.Equal(t, path3.String(), path4.String())
	assert.NotEqual(t, path1.String(), path3.String())
}
