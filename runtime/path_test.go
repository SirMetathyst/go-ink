package runtime

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPath_Equals_Fail(t *testing.T) {

	// NewPath

	path1 := NewPath()
	path2 := NewRelativePath()
	assert.False(t, path1.Equals(path2))

	// NewPathFromComponentsString

	path1 = NewPathFromComponentsString("")
	path2 = NewPathFromComponentsString(".")
	assert.False(t, path1.Equals(path2))

	path1 = NewPathFromComponentsString("2")
	path2 = NewPathFromComponentsString("4")
	assert.False(t, path1.Equals(path2))

	path1 = NewPathFromComponentsString("hello.1.world")
	path2 = NewPathFromComponentsString(".hello.1.world")
	assert.False(t, path1.Equals(path2))

	path1 = NewPathFromComponentsString("hellq.1.world")
	path2 = NewPathFromComponentsString("hello.1.world")
	assert.False(t, path1.Equals(path2))

	path1 = NewPathFromComponentsString("^.hello.1.world")
	path2 = NewPathFromComponentsString(".hello.1.world")
	assert.False(t, path1.Equals(path2))

	// NewPathFromComponents

	path1 = NewPathFromComponents([]PathComponent{})
	path2 = NewRelativePathFromComponents([]PathComponent{})
	assert.False(t, path1.Equals(path2))

	path1 = NewPathFromComponents([]PathComponent{MustNewPathComponentFromIndex(2)})
	path2 = NewPathFromComponents([]PathComponent{MustNewPathComponentFromIndex(4)})
	assert.False(t, path1.Equals(path2))

	path1 = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(1), MustNewPathComponentFromName("world")})
	path1 = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(1), MustNewPathComponentFromName("world")})
	assert.False(t, path1.Equals(path2))

	path1 = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("hellq"), MustNewPathComponentFromIndex(1), MustNewPathComponentFromName("world")})
	path1 = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(1), MustNewPathComponentFromName("world")})
	assert.False(t, path1.Equals(path2))

	path1 = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("^"), MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(1), MustNewPathComponentFromName("world")})
	path1 = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(1), MustNewPathComponentFromName("world")})
	assert.False(t, path1.Equals(path2))
}

func TestPath_Equals(t *testing.T) {

	// NewPath

	path1 := NewPath()
	path2 := NewPath()
	assert.True(t, path1.Equals(path2))

	path1 = NewRelativePath()
	path2 = NewRelativePath()
	assert.True(t, path1.Equals(path2))

	// NewPathFromComponentsString

	path1 = NewPathFromComponentsString("")
	path2 = NewPathFromComponentsString("")
	assert.True(t, path1.Equals(path2))

	path1 = NewPathFromComponentsString(".")
	path2 = NewPathFromComponentsString(".")
	assert.True(t, path1.Equals(path2))

	path1 = NewPathFromComponentsString("5")
	path2 = NewPathFromComponentsString("5")
	assert.True(t, path1.Equals(path2))

	path1 = NewPathFromComponentsString("^")
	path2 = NewPathFromComponentsString("^")
	assert.True(t, path1.Equals(path2))

	path1 = NewPathFromComponentsString("hello.1.world")
	path2 = NewPathFromComponentsString("hello.1.world")
	assert.True(t, path1.Equals(path2))

	path1 = NewPathFromComponentsString(".hello.1.world")
	path2 = NewPathFromComponentsString(".hello.1.world")
	assert.True(t, path1.Equals(path2))

	// NewPathFromComponents

	path1 = NewPathFromComponents([]PathComponent{})
	path2 = NewPathFromComponents([]PathComponent{})
	assert.True(t, path1.Equals(path2))

	path1 = NewRelativePathFromComponents([]PathComponent{})
	path2 = NewRelativePathFromComponents([]PathComponent{})
	assert.True(t, path1.Equals(path2))

	path1 = NewPathFromComponents([]PathComponent{MustNewPathComponentFromIndex(5)})
	path2 = NewPathFromComponents([]PathComponent{MustNewPathComponentFromIndex(5)})
	assert.True(t, path1.Equals(path2))

	path1 = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("^")})
	path2 = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("^")})
	assert.True(t, path1.Equals(path2))

	path1 = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(1), MustNewPathComponentFromName("world")})
	path2 = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(1), MustNewPathComponentFromName("world")})
	assert.True(t, path1.Equals(path2))

	path1 = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(1), MustNewPathComponentFromName("world")})
	path2 = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(1), MustNewPathComponentFromName("world")})
	assert.True(t, path1.Equals(path2))
}

func TestPath_String(t *testing.T) {

	// NewPath

	path1 := NewPath()
	path2 := NewPath()
	assert.Equal(t, "", path1.String())
	assert.Equal(t, "", path2.String())

	path1 = NewRelativePath()
	path2 = NewRelativePath()
	assert.Equal(t, ".", path1.String())
	assert.Equal(t, ".", path2.String())

	// NewPathFromComponentsString

	path1 = NewPathFromComponentsString("")
	path2 = NewPathFromComponentsString(".")
	assert.Equal(t, "", path1.String())
	assert.Equal(t, ".", path2.String())

	path1 = NewPathFromComponentsString("7")
	path2 = NewPathFromComponentsString("5")
	assert.Equal(t, "7", path1.String())
	assert.Equal(t, "5", path2.String())

	path1 = NewPathFromComponentsString(".")
	path2 = NewPathFromComponentsString("^")
	assert.Equal(t, ".", path1.String())
	assert.Equal(t, "^", path2.String())

	path1 = NewPathFromComponentsString("hello.1.world")
	path2 = NewPathFromComponentsString(".hello.1.world")
	assert.Equal(t, "hello.1.world", path1.String())
	assert.Equal(t, ".hello.1.world", path2.String())

	path1 = NewPathFromComponentsString("hellq.1.world")
	path2 = NewPathFromComponentsString("hello.1.world")
	assert.Equal(t, "hellq.1.world", path1.String())
	assert.Equal(t, "hello.1.world", path2.String())

	path1 = NewPathFromComponentsString("^.hello.1.world")
	path2 = NewPathFromComponentsString(".hello.1.world")
	assert.Equal(t, "^.hello.1.world", path1.String())
	assert.Equal(t, ".hello.1.world", path2.String())

	// NewPathFromComponents

	path1 = NewPathFromComponents([]PathComponent{})
	path2 = NewRelativePathFromComponents([]PathComponent{})
	assert.Equal(t, "", path1.String())
	assert.Equal(t, ".", path2.String())

	path1 = NewPathFromComponents([]PathComponent{MustNewPathComponentFromIndex(7)})
	path2 = NewPathFromComponents([]PathComponent{MustNewPathComponentFromIndex(5)})
	assert.Equal(t, "7", path1.String())
	assert.Equal(t, "5", path2.String())

	path1 = NewRelativePathFromComponents([]PathComponent{})
	path2 = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("^")})
	assert.Equal(t, ".", path1.String())
	assert.Equal(t, "^", path2.String())

	path1 = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(1), MustNewPathComponentFromName("world")})
	path2 = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(1), MustNewPathComponentFromName("world")})
	assert.Equal(t, "hello.1.world", path1.String())
	assert.Equal(t, ".hello.1.world", path2.String())

	path1 = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("hellq"), MustNewPathComponentFromIndex(1), MustNewPathComponentFromName("world")})
	path2 = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(1), MustNewPathComponentFromName("world")})
	assert.Equal(t, "hellq.1.world", path1.String())
	assert.Equal(t, "hello.1.world", path2.String())

	path1 = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("^"), MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(1), MustNewPathComponentFromName("world")})
	path2 = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(1), MustNewPathComponentFromName("world")})
	assert.Equal(t, "^.hello.1.world", path1.String())
	assert.Equal(t, ".hello.1.world", path2.String())
}

func TestPath_IsRelative_Fail(t *testing.T) {

	// NewPath

	path := NewPath()
	assert.False(t, path.IsRelative())

	// NewPathFromComponentsString

	path = NewPathFromComponentsString("")
	assert.False(t, path.IsRelative())

	path = NewPathFromComponentsString("hello.world")
	assert.False(t, path.IsRelative())

	path = NewPathFromComponentsString("^.hello.world")
	assert.False(t, path.IsRelative())

	path = NewPathFromComponentsString("^.hello.world.5")
	assert.False(t, path.IsRelative())

	// NewPathFromComponents

	path = NewPathFromComponents(nil)
	assert.False(t, path.IsRelative())

	path = NewPathFromComponents([]PathComponent{})
	assert.False(t, path.IsRelative())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromName("world")})
	assert.False(t, path.IsRelative())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("^"), MustNewPathComponentFromName("hello"), MustNewPathComponentFromName("world")})
	assert.False(t, path.IsRelative())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("^"), MustNewPathComponentFromName("hello"), MustNewPathComponentFromName("world"), MustNewPathComponentFromIndex(5)})
	assert.False(t, path.IsRelative())
}

func TestPath_IsRelative(t *testing.T) {

	// NewPath

	path := NewRelativePath()
	assert.True(t, path.IsRelative())

	// NewPathFromComponentsString

	path = NewPathFromComponentsString(".^")
	assert.True(t, path.IsRelative())

	path = NewPathFromComponentsString(".hello.world")
	assert.True(t, path.IsRelative())

	path = NewPathFromComponentsString(".^.hello.world")
	assert.True(t, path.IsRelative())

	path = NewPathFromComponentsString(".^.hello.world.5")
	assert.True(t, path.IsRelative())

	// NewRelativePathFromComponents

	path = NewRelativePathFromComponents([]PathComponent{})
	assert.True(t, path.IsRelative())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromName("world")})
	assert.True(t, path.IsRelative())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("^"), MustNewPathComponentFromName("hello"), MustNewPathComponentFromName("world")})
	assert.True(t, path.IsRelative())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("^"), MustNewPathComponentFromName("hello"), MustNewPathComponentFromName("world"), MustNewPathComponentFromIndex(5)})
	assert.True(t, path.IsRelative())
}

func TestPath_Length(t *testing.T) {

	// NewPath

	path := NewPath()
	assert.Equal(t, 0, path.Length())

	path = NewRelativePath()
	assert.Equal(t, 0, path.Length())

	// NewPathFromComponentsString

	path = NewPathFromComponentsString("")
	assert.Equal(t, 0, path.Length())

	path = NewPathFromComponentsString("^")
	assert.Equal(t, 1, path.Length())

	path = NewPathFromComponentsString("^.hello")
	assert.Equal(t, 2, path.Length())

	path = NewPathFromComponentsString("^.hello.world")
	assert.Equal(t, 3, path.Length())

	path = NewPathFromComponentsString("^.hello.world.5")
	assert.Equal(t, 4, path.Length())

	path = NewPathFromComponentsString(".^")
	assert.Equal(t, 1, path.Length())

	path = NewPathFromComponentsString(".^.hello")
	assert.Equal(t, 2, path.Length())

	path = NewPathFromComponentsString(".^.hello.world")
	assert.Equal(t, 3, path.Length())

	path = NewPathFromComponentsString(".^.hello.world.5")
	assert.Equal(t, 4, path.Length())

	// NewRelativePathFromComponents

	path = NewRelativePathFromComponents(nil)
	assert.Equal(t, 0, path.Length())

	path = NewRelativePathFromComponents([]PathComponent{})
	assert.Equal(t, 0, path.Length())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("^")})
	assert.Equal(t, 1, path.Length())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("^"), MustNewPathComponentFromName("hello")})
	assert.Equal(t, 2, path.Length())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("^"), MustNewPathComponentFromName("hello"), MustNewPathComponentFromName("world")})
	assert.Equal(t, 3, path.Length())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("^"), MustNewPathComponentFromName("hello"), MustNewPathComponentFromName("world"), MustNewPathComponentFromIndex(5)})
	assert.Equal(t, 4, path.Length())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("^")})
	assert.Equal(t, 1, path.Length())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("^"), MustNewPathComponentFromName("hello")})
	assert.Equal(t, 2, path.Length())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("^"), MustNewPathComponentFromName("hello"), MustNewPathComponentFromName("world")})
	assert.Equal(t, 3, path.Length())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("^"), MustNewPathComponentFromName("hello"), MustNewPathComponentFromName("world"), MustNewPathComponentFromIndex(5)})
	assert.Equal(t, 4, path.Length())
}

func TestPath_Component_Fail(t *testing.T) {

	// NewPath

	path := NewPath()
	_, ok1 := path.Component(0)
	_, ok2 := path.Component(1)
	_, ok3 := path.Component(2)
	assert.False(t, ok1)
	assert.False(t, ok2)
	assert.False(t, ok3)

	path = NewRelativePath()
	_, ok1 = path.Component(0)
	_, ok2 = path.Component(1)
	_, ok3 = path.Component(2)
	assert.False(t, ok1)
	assert.False(t, ok2)
	assert.False(t, ok3)

	// NewPathFromComponentsString

	path = NewPathFromComponentsString(".")
	_, ok1 = path.Component(0)
	_, ok2 = path.Component(1)
	_, ok3 = path.Component(2)
	assert.False(t, ok1)
	assert.False(t, ok2)
	assert.False(t, ok3)

	path = NewPathFromComponentsString(".^")
	_, ok1 = path.Component(0)
	_, ok2 = path.Component(1)
	_, ok3 = path.Component(2)
	assert.True(t, ok1)
	assert.False(t, ok2)
	assert.False(t, ok3)

	path = NewPathFromComponentsString(".^.hello")
	_, ok1 = path.Component(0)
	_, ok2 = path.Component(1)
	_, ok3 = path.Component(2)
	assert.True(t, ok1)
	assert.True(t, ok2)
	assert.False(t, ok3)

	// NewRelativePathFromComponents

	path = NewRelativePathFromComponents(nil)
	_, ok1 = path.Component(0)
	_, ok2 = path.Component(1)
	_, ok3 = path.Component(2)
	assert.False(t, ok1)
	assert.False(t, ok2)
	assert.False(t, ok3)

	path = NewRelativePathFromComponents([]PathComponent{})
	_, ok1 = path.Component(0)
	_, ok2 = path.Component(1)
	_, ok3 = path.Component(2)
	assert.False(t, ok1)
	assert.False(t, ok2)
	assert.False(t, ok3)

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("^")})
	_, ok1 = path.Component(0)
	_, ok2 = path.Component(1)
	_, ok3 = path.Component(2)
	assert.True(t, ok1)
	assert.False(t, ok2)
	assert.False(t, ok3)

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("^")})
	_, ok1 = path.Component(0)
	_, ok2 = path.Component(1)
	_, ok3 = path.Component(2)
	assert.True(t, ok1)
	assert.False(t, ok2)
	assert.False(t, ok3)

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("^"), MustNewPathComponentFromName("hello")})
	_, ok1 = path.Component(0)
	_, ok2 = path.Component(1)
	_, ok3 = path.Component(2)
	assert.True(t, ok1)
	assert.True(t, ok2)
	assert.False(t, ok3)
}

func TestPath_Component(t *testing.T) {

	// NewPathFromComponentsString

	path := NewPathFromComponentsString(".hello.5.world")
	comp1, ok1 := path.Component(0)
	comp2, ok2 := path.Component(1)
	comp3, ok3 := path.Component(2)
	assert.True(t, ok1)
	assert.Equal(t, "hello", comp1.String())
	assert.True(t, ok2)
	assert.Equal(t, "5", comp2.String())
	assert.True(t, ok3)
	assert.Equal(t, "world", comp3.String())

	path = NewPathFromComponentsString("hello.5.world")
	comp1, ok1 = path.Component(0)
	comp2, ok2 = path.Component(1)
	comp3, ok3 = path.Component(2)
	assert.True(t, ok1)
	assert.Equal(t, "hello", comp1.String())
	assert.True(t, ok2)
	assert.Equal(t, "5", comp2.String())
	assert.True(t, ok3)
	assert.Equal(t, "world", comp3.String())

	// NewRelativePathFromComponents

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(5), MustNewPathComponentFromName("world")})
	_, ok1 = path.Component(0)
	_, ok2 = path.Component(1)
	_, ok3 = path.Component(2)
	assert.True(t, ok1)
	assert.Equal(t, "hello", comp1.String())
	assert.True(t, ok2)
	assert.Equal(t, "5", comp2.String())
	assert.True(t, ok3)
	assert.Equal(t, "world", comp3.String())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(5), MustNewPathComponentFromName("world")})
	_, ok1 = path.Component(0)
	_, ok2 = path.Component(1)
	_, ok3 = path.Component(2)
	assert.True(t, ok1)
	assert.Equal(t, "hello", comp1.String())
	assert.True(t, ok2)
	assert.Equal(t, "5", comp2.String())
	assert.True(t, ok3)
	assert.Equal(t, "world", comp3.String())
}

func TestPath_ContainsNamedComponent_Fail(t *testing.T) {

	// NewPath

	path := NewPath()
	assert.False(t, path.ContainsNamedComponent())

	path = NewRelativePath()
	assert.False(t, path.ContainsNamedComponent())

	// NewPathFromComponentsString

	path = NewPathFromComponentsString("")
	assert.False(t, path.ContainsNamedComponent())

	path = NewPathFromComponentsString(".")
	assert.False(t, path.ContainsNamedComponent())

	path = NewPathFromComponentsString("5")
	assert.False(t, path.ContainsNamedComponent())

	// NewPathFromComponents

	path = NewPathFromComponents([]PathComponent{})
	assert.False(t, path.ContainsNamedComponent())

	path = NewRelativePathFromComponents([]PathComponent{})
	assert.False(t, path.ContainsNamedComponent())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromIndex(5)})
	assert.False(t, path.ContainsNamedComponent())
}

func TestPath_ContainsNamedComponent(t *testing.T) {

	// NewPathFromComponentsString

	path := NewPathFromComponentsString("^")
	assert.True(t, path.ContainsNamedComponent())

	path = NewPathFromComponentsString(".hello")
	assert.True(t, path.ContainsNamedComponent())

	path = NewPathFromComponentsString("5.hello")
	assert.True(t, path.ContainsNamedComponent())

	// NewPathFromComponents

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("^")})
	assert.True(t, path.ContainsNamedComponent())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("hello")})
	assert.True(t, path.ContainsNamedComponent())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromIndex(5), MustNewPathComponentFromName("hello")})
	assert.True(t, path.ContainsNamedComponent())
}

func TestPath_FirstComponent_Fail(t *testing.T) {

	// NewPath

	path := NewPath()
	_, ok := path.FirstComponent()
	assert.False(t, ok)

	path = NewRelativePath()
	_, ok = path.FirstComponent()
	assert.False(t, ok)

	// NewPathFromComponentsString

	path = NewPathFromComponentsString("")
	_, ok = path.FirstComponent()
	assert.False(t, ok)

	path = NewPathFromComponentsString(".")
	_, ok = path.FirstComponent()
	assert.False(t, ok)

	// NewPathFromComponents

	path = NewPathFromComponents([]PathComponent{})
	_, ok = path.FirstComponent()
	assert.False(t, ok)

	path = NewRelativePathFromComponents([]PathComponent{})
	_, ok = path.FirstComponent()
	assert.False(t, ok)
}

func TestPath_FirstComponent(t *testing.T) {

	// NewPathFromComponentsString

	path := NewPathFromComponentsString("^")
	comp, ok := path.FirstComponent()
	assert.True(t, ok)
	assert.Equal(t, "^", comp.String())

	path = NewPathFromComponentsString("hello")
	comp, ok = path.FirstComponent()
	assert.True(t, ok)
	assert.Equal(t, "hello", comp.String())

	path = NewPathFromComponentsString("hello.5.world")
	comp, ok = path.FirstComponent()
	assert.True(t, ok)
	assert.Equal(t, "hello", comp.String())

	path = NewPathFromComponentsString("world.5.hello")
	comp, ok = path.FirstComponent()
	assert.True(t, ok)
	assert.Equal(t, "world", comp.String())

	path = NewPathFromComponentsString("10.world.5.hello")
	comp, ok = path.FirstComponent()
	assert.True(t, ok)
	assert.Equal(t, "10", comp.String())

	path = NewPathFromComponentsString(".hello")
	comp, ok = path.FirstComponent()
	assert.True(t, ok)
	assert.Equal(t, "hello", comp.String())

	path = NewPathFromComponentsString(".hello.5.world")
	comp, ok = path.FirstComponent()
	assert.True(t, ok)
	assert.Equal(t, "hello", comp.String())

	path = NewPathFromComponentsString(".world.5.hello")
	comp, ok = path.FirstComponent()
	assert.True(t, ok)
	assert.Equal(t, "world", comp.String())

	path = NewPathFromComponentsString(".10.world.5.hello")
	comp, ok = path.FirstComponent()
	assert.True(t, ok)
	assert.Equal(t, "10", comp.String())

	// NewPathFromComponents

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("^")})
	comp, ok = path.FirstComponent()
	assert.True(t, ok)
	assert.Equal(t, "^", comp.String())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("hello")})
	comp, ok = path.FirstComponent()
	assert.True(t, ok)
	assert.Equal(t, "hello", comp.String())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(5), MustNewPathComponentFromName("world")})
	comp, ok = path.FirstComponent()
	assert.True(t, ok)
	assert.Equal(t, "hello", comp.String())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("world"), MustNewPathComponentFromIndex(5), MustNewPathComponentFromName("hello")})
	comp, ok = path.FirstComponent()
	assert.True(t, ok)
	assert.Equal(t, "world", comp.String())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromIndex(10), MustNewPathComponentFromName("world"), MustNewPathComponentFromIndex(5), MustNewPathComponentFromName("hello")})
	comp, ok = path.FirstComponent()
	assert.True(t, ok)
	assert.Equal(t, "10", comp.String())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("hello")})
	comp, ok = path.FirstComponent()
	assert.True(t, ok)
	assert.Equal(t, "hello", comp.String())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(5), MustNewPathComponentFromName("world")})
	comp, ok = path.FirstComponent()
	assert.True(t, ok)
	assert.Equal(t, "hello", comp.String())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("world"), MustNewPathComponentFromIndex(5), MustNewPathComponentFromName("hello")})
	comp, ok = path.FirstComponent()
	assert.True(t, ok)
	assert.Equal(t, "world", comp.String())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromIndex(10), MustNewPathComponentFromName("world"), MustNewPathComponentFromIndex(5), MustNewPathComponentFromName("hello")})
	comp, ok = path.FirstComponent()
	assert.True(t, ok)
	assert.Equal(t, "10", comp.String())
}

func TestPath_LastComponent_Fail(t *testing.T) {

	// NewPath

	path := NewPath()
	_, ok := path.LastComponent()
	assert.False(t, ok)

	path = NewRelativePath()
	_, ok = path.LastComponent()
	assert.False(t, ok)

	// NewPathFromComponentsString

	path = NewPathFromComponentsString("")
	_, ok = path.LastComponent()
	assert.False(t, ok)

	path = NewPathFromComponentsString(".")
	_, ok = path.LastComponent()
	assert.False(t, ok)

	// NewPathFromComponents

	path = NewPathFromComponents([]PathComponent{})
	_, ok = path.LastComponent()
	assert.False(t, ok)

	path = NewRelativePathFromComponents([]PathComponent{})
	_, ok = path.LastComponent()
	assert.False(t, ok)
}

func TestPath_LastComponent(t *testing.T) {

	// NewPathFromComponentsString

	path := NewPathFromComponentsString("^")
	comp, ok := path.LastComponent()
	assert.True(t, ok)
	assert.Equal(t, "^", comp.String())

	path = NewPathFromComponentsString("hello")
	comp, ok = path.LastComponent()
	assert.True(t, ok)
	assert.Equal(t, "hello", comp.String())

	path = NewPathFromComponentsString("hello.5.world")
	comp, ok = path.LastComponent()
	assert.True(t, ok)
	assert.Equal(t, "world", comp.String())

	path = NewPathFromComponentsString("world.5.hello")
	comp, ok = path.LastComponent()
	assert.True(t, ok)
	assert.Equal(t, "hello", comp.String())

	path = NewPathFromComponentsString("world.5.hello.10")
	comp, ok = path.LastComponent()
	assert.True(t, ok)
	assert.Equal(t, "10", comp.String())

	path = NewPathFromComponentsString(".hello")
	comp, ok = path.LastComponent()
	assert.True(t, ok)
	assert.Equal(t, "hello", comp.String())

	path = NewPathFromComponentsString(".hello.5.world")
	comp, ok = path.LastComponent()
	assert.True(t, ok)
	assert.Equal(t, "world", comp.String())

	path = NewPathFromComponentsString(".world.5.hello")
	comp, ok = path.LastComponent()
	assert.True(t, ok)
	assert.Equal(t, "hello", comp.String())

	path = NewPathFromComponentsString(".world.5.hello.10")
	comp, ok = path.LastComponent()
	assert.True(t, ok)
	assert.Equal(t, "10", comp.String())

	// NewPathFromComponents

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("^")})
	comp, ok = path.LastComponent()
	assert.True(t, ok)
	assert.Equal(t, "^", comp.String())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("hello")})
	comp, ok = path.LastComponent()
	assert.True(t, ok)
	assert.Equal(t, "hello", comp.String())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(5), MustNewPathComponentFromName("world")})
	comp, ok = path.LastComponent()
	assert.True(t, ok)
	assert.Equal(t, "world", comp.String())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("world"), MustNewPathComponentFromIndex(5), MustNewPathComponentFromName("hello")})
	comp, ok = path.LastComponent()
	assert.True(t, ok)
	assert.Equal(t, "hello", comp.String())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("world"), MustNewPathComponentFromIndex(5), MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(10)})
	comp, ok = path.LastComponent()
	assert.True(t, ok)
	assert.Equal(t, "10", comp.String())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("hello")})
	comp, ok = path.LastComponent()
	assert.True(t, ok)
	assert.Equal(t, "hello", comp.String())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(5), MustNewPathComponentFromName("world")})
	comp, ok = path.LastComponent()
	assert.True(t, ok)
	assert.Equal(t, "world", comp.String())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("world"), MustNewPathComponentFromIndex(5), MustNewPathComponentFromName("hello")})
	comp, ok = path.LastComponent()
	assert.True(t, ok)
	assert.Equal(t, "hello", comp.String())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("world"), MustNewPathComponentFromIndex(5), MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(10)})
	comp, ok = path.LastComponent()
	assert.True(t, ok)
	assert.Equal(t, "10", comp.String())
}

func TestPath_Tail(t *testing.T) {

	// NewPath

	path := NewPath()
	tailPath := path.Tail()
	assert.Equal(t, 0, tailPath.Length())
	assert.True(t, tailPath.IsRelative())

	path = NewRelativePath()
	tailPath = path.Tail()
	assert.Equal(t, 0, tailPath.Length())
	assert.True(t, tailPath.IsRelative())

	// NewPathFromComponentsString

	path = NewPathFromComponentsString("")
	tailPath = path.Tail()
	assert.Equal(t, 0, tailPath.Length())
	assert.True(t, tailPath.IsRelative())

	path = NewPathFromComponentsString("^")
	tailPath = path.Tail()
	assert.Equal(t, 0, tailPath.Length())
	assert.True(t, tailPath.IsRelative())

	path = NewPathFromComponentsString("hello.5")
	tailPath = path.Tail()
	tailPathComp, ok := tailPath.Component(0)
	assert.Equal(t, 1, tailPath.Length())
	assert.True(t, ok)
	assert.False(t, tailPath.IsRelative())
	assert.Equal(t, "5", tailPathComp.String())

	path = NewPathFromComponentsString("hello.5.world")
	tailPath = path.Tail()
	tailPathComp, ok = tailPath.Component(1)
	assert.Equal(t, 2, tailPath.Length())
	assert.True(t, ok)
	assert.False(t, tailPath.IsRelative())
	assert.Equal(t, "world", tailPathComp.String())

	path = NewPathFromComponentsString(".hello.5")
	tailPath = path.Tail()
	tailPathComp, ok = tailPath.Component(0)
	assert.Equal(t, 1, tailPath.Length())
	assert.True(t, ok)
	assert.False(t, tailPath.IsRelative())
	assert.Equal(t, "5", tailPathComp.String())

	path = NewPathFromComponentsString(".hello.5.world")
	tailPath = path.Tail()
	tailPathComp, ok = tailPath.Component(1)
	assert.Equal(t, 2, tailPath.Length())
	assert.True(t, ok)
	assert.False(t, tailPath.IsRelative())
	assert.Equal(t, "world", tailPathComp.String())

	// NewPathFromComponents

	path = NewPathFromComponents([]PathComponent{})
	tailPath = path.Tail()
	assert.Equal(t, 0, tailPath.Length())
	assert.True(t, tailPath.IsRelative())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("^")})
	tailPath = path.Tail()
	assert.Equal(t, 0, tailPath.Length())
	assert.True(t, tailPath.IsRelative())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(5)})
	tailPath = path.Tail()
	tailPathComp, ok = tailPath.Component(0)
	assert.Equal(t, 1, tailPath.Length())
	assert.True(t, ok)
	assert.False(t, tailPath.IsRelative())
	assert.Equal(t, "5", tailPathComp.String())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(5), MustNewPathComponentFromName("world")})
	tailPath = path.Tail()
	tailPathComp, ok = tailPath.Component(1)
	assert.Equal(t, 2, tailPath.Length())
	assert.True(t, ok)
	assert.False(t, tailPath.IsRelative())
	assert.Equal(t, "world", tailPathComp.String())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(5)})
	tailPath = path.Tail()
	tailPathComp, ok = tailPath.Component(0)
	assert.Equal(t, 1, tailPath.Length())
	assert.True(t, ok)
	assert.False(t, tailPath.IsRelative())
	assert.Equal(t, "5", tailPathComp.String())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromIndex(5), MustNewPathComponentFromName("world")})
	tailPath = path.Tail()
	tailPathComp, ok = tailPath.Component(1)
	assert.Equal(t, 2, tailPath.Length())
	assert.True(t, ok)
	assert.False(t, tailPath.IsRelative())
	assert.Equal(t, "world", tailPathComp.String())
}

func TestPath_NewPathByAppendingComponent(t *testing.T) {

	// NewPath

	path := NewPath()
	path = path.NewPathByAppendingComponent(MustNewPathComponentFromName("^"))
	assert.Equal(t, "^", path.String())

	path = NewRelativePath()
	path = path.NewPathByAppendingComponent(MustNewPathComponentFromName("^"))
	assert.Equal(t, ".^", path.String())

	// NewPathFromComponentsString

	path = NewPathFromComponentsString("")
	path = path.NewPathByAppendingComponent(MustNewPathComponentFromName("^"))
	assert.Equal(t, "^", path.String())

	path = NewPathFromComponentsString("^")
	path = path.NewPathByAppendingComponent(MustNewPathComponentFromName("^"))
	assert.Equal(t, "^.^", path.String())

	path = NewPathFromComponentsString("hello")
	path = path.NewPathByAppendingComponent(MustNewPathComponentFromName("5"))
	assert.Equal(t, "hello.5", path.String())

	path = NewPathFromComponentsString("hello")
	path = path.NewPathByAppendingComponent(MustNewPathComponentFromName("5"))
	path = path.NewPathByAppendingComponent(MustNewPathComponentFromName("world"))
	assert.Equal(t, "hello.5.world", path.String())

	path = NewPathFromComponentsString(".")
	path = path.NewPathByAppendingComponent(MustNewPathComponentFromName("^"))
	assert.Equal(t, ".^", path.String())

	path = NewPathFromComponentsString(".^")
	path = path.NewPathByAppendingComponent(MustNewPathComponentFromName("^"))
	assert.Equal(t, ".^.^", path.String())

	path = NewPathFromComponentsString(".hello")
	path = path.NewPathByAppendingComponent(MustNewPathComponentFromName("5"))
	assert.Equal(t, ".hello.5", path.String())

	path = NewPathFromComponentsString(".hello")
	path = path.NewPathByAppendingComponent(MustNewPathComponentFromName("5"))
	path = path.NewPathByAppendingComponent(MustNewPathComponentFromName("world"))
	assert.Equal(t, ".hello.5.world", path.String())

	// NewPathFromComponents

	path = NewPathFromComponents(nil)
	path = path.NewPathByAppendingComponent(MustNewPathComponentFromName("^"))
	assert.Equal(t, "^", path.String())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("^")})
	path = path.NewPathByAppendingComponent(MustNewPathComponentFromName("^"))
	assert.Equal(t, "^.^", path.String())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("hello")})
	path = path.NewPathByAppendingComponent(MustNewPathComponentFromName("5"))
	assert.Equal(t, "hello.5", path.String())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("hello")})
	path = path.NewPathByAppendingComponent(MustNewPathComponentFromName("5"))
	path = path.NewPathByAppendingComponent(MustNewPathComponentFromName("world"))
	assert.Equal(t, "hello.5.world", path.String())

	path = NewRelativePathFromComponents(nil)
	path = path.NewPathByAppendingComponent(MustNewPathComponentFromName("^"))
	assert.Equal(t, ".^", path.String())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("^")})
	path = path.NewPathByAppendingComponent(MustNewPathComponentFromName("^"))
	assert.Equal(t, ".^.^", path.String())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("hello")})
	path = path.NewPathByAppendingComponent(MustNewPathComponentFromName("5"))
	assert.Equal(t, ".hello.5", path.String())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("hello")})
	path = path.NewPathByAppendingComponent(MustNewPathComponentFromName("5"))
	path = path.NewPathByAppendingComponent(MustNewPathComponentFromName("world"))
	assert.Equal(t, ".hello.5.world", path.String())
}

func TestPath_NewPathByAppendingPath(t *testing.T) {

	// NewPath

	path := NewPath()
	path = path.NewPathByAppendingPath(NewPathFromComponentsString("^"))
	assert.Equal(t, "", path.String())
	assert.False(t, path.IsRelative())

	path = NewRelativePath()
	path = path.NewPathByAppendingPath(NewPathFromComponentsString("^"))
	assert.Equal(t, ".", path.String())
	assert.True(t, path.IsRelative())

	// NewPathFromComponentsString

	path = NewPathFromComponentsString("hello")
	path = path.NewPathByAppendingPath(NewPathFromComponentsString("^"))
	assert.Equal(t, "", path.String())
	assert.False(t, path.IsRelative())

	path = NewPathFromComponentsString("hello.world")
	path = path.NewPathByAppendingPath(NewPathFromComponentsString("^"))
	assert.Equal(t, "hello", path.String())
	assert.False(t, path.IsRelative())

	path = NewPathFromComponentsString("hello.world.5")
	path = path.NewPathByAppendingPath(NewPathFromComponentsString("^"))
	assert.Equal(t, "hello.world", path.String())
	assert.False(t, path.IsRelative())

	path = NewPathFromComponentsString(".hello")
	path = path.NewPathByAppendingPath(NewPathFromComponentsString("^"))
	assert.Equal(t, ".", path.String())
	assert.True(t, path.IsRelative())

	path = NewPathFromComponentsString(".hello.world")
	path = path.NewPathByAppendingPath(NewPathFromComponentsString("^"))
	assert.Equal(t, ".hello", path.String())
	assert.True(t, path.IsRelative())

	path = NewPathFromComponentsString(".hello.world.5")
	path = path.NewPathByAppendingPath(NewPathFromComponentsString("^"))
	assert.Equal(t, ".hello.world", path.String())
	assert.True(t, path.IsRelative())

	path = NewPathFromComponentsString("hello")
	path = path.NewPathByAppendingPath(NewPathFromComponentsString("world"))
	assert.Equal(t, "hello.world", path.String())
	assert.False(t, path.IsRelative())

	path = NewPathFromComponentsString("hello.world")
	path = path.NewPathByAppendingPath(NewPathFromComponentsString("5"))
	assert.Equal(t, "hello.world.5", path.String())
	assert.False(t, path.IsRelative())

	path = NewPathFromComponentsString(".hello")
	path = path.NewPathByAppendingPath(NewPathFromComponentsString("world"))
	assert.Equal(t, ".hello.world", path.String())
	assert.True(t, path.IsRelative())

	path = NewPathFromComponentsString(".hello.world")
	path = path.NewPathByAppendingPath(NewPathFromComponentsString("5"))
	assert.Equal(t, ".hello.world.5", path.String())
	assert.True(t, path.IsRelative())

	// NewRelativePathFromComponents

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("hello")})
	path = path.NewPathByAppendingPath(NewPathFromComponentsString("^"))
	assert.Equal(t, "", path.String())
	assert.False(t, path.IsRelative())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromName("world")})
	path = path.NewPathByAppendingPath(NewPathFromComponentsString("^"))
	assert.Equal(t, "hello", path.String())
	assert.False(t, path.IsRelative())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromName("world"), MustNewPathComponentFromIndex(5)})
	path = path.NewPathByAppendingPath(NewPathFromComponentsString("^"))
	assert.Equal(t, "hello.world", path.String())
	assert.False(t, path.IsRelative())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("hello")})
	path = path.NewPathByAppendingPath(NewPathFromComponentsString("^"))
	assert.Equal(t, ".", path.String())
	assert.True(t, path.IsRelative())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromName("world")})
	path = path.NewPathByAppendingPath(NewPathFromComponentsString("^"))
	assert.Equal(t, ".hello", path.String())
	assert.True(t, path.IsRelative())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromName("world"), MustNewPathComponentFromIndex(5)})
	path = path.NewPathByAppendingPath(NewPathFromComponentsString("^"))
	assert.Equal(t, ".hello.world", path.String())
	assert.True(t, path.IsRelative())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("hello")})
	path = path.NewPathByAppendingPath(NewPathFromComponentsString("world"))
	assert.Equal(t, "hello.world", path.String())
	assert.False(t, path.IsRelative())

	path = NewPathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromName("world")})
	path = path.NewPathByAppendingPath(NewPathFromComponentsString("5"))
	assert.Equal(t, "hello.world.5", path.String())
	assert.False(t, path.IsRelative())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("hello")})
	path = path.NewPathByAppendingPath(NewPathFromComponentsString("world"))
	assert.Equal(t, ".hello.world", path.String())
	assert.True(t, path.IsRelative())

	path = NewRelativePathFromComponents([]PathComponent{MustNewPathComponentFromName("hello"), MustNewPathComponentFromName("world")})
	path = path.NewPathByAppendingPath(NewPathFromComponentsString("5"))
	assert.Equal(t, ".hello.world.5", path.String())
	assert.True(t, path.IsRelative())
}
