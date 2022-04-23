package runtime

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPathComponent_New_Fail(t *testing.T) {

	// NewPathComponentFromIndex

	c, err := NewPathComponentFromIndex(-10)
	assert.ErrorIs(t, err, ErrPathComponentIndex)
	assert.NotNil(t, c)

	// NewPathComponentFromName

	c, err = NewPathComponentFromName("")
	assert.ErrorIs(t, err, ErrPathComponentName)
	assert.NotNil(t, c)
}

func TestPathComponent_New(t *testing.T) {

	// NewPathComponentFromIndex

	c, err := NewPathComponentFromIndex(0)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	c, err = NewPathComponentFromIndex(10)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	// NewPathComponentFromName

	c, err = NewPathComponentFromName("name")
	assert.Nil(t, err)
	assert.NotNil(t, c)
}

func TestPathComponent_Index_Fail(t *testing.T) {

	// NewPathComponentFromName

	c, err := NewPathComponentFromName("name")
	assert.Nil(t, err)
	assert.Equal(t, -1, c.Index())
}

func TestPathComponent_Index(t *testing.T) {

	// NewPathComponentFromIndex

	c, err := NewPathComponentFromIndex(10)
	assert.Nil(t, err)
	assert.Equal(t, 10, c.Index())

	c, err = NewPathComponentFromIndex(0)
	assert.Nil(t, err)
	assert.Equal(t, 0, c.Index())
}

func TestPathComponent_IsIndex_Fail(t *testing.T) {

	// NewPathComponentFromName

	c, err := NewPathComponentFromName("name")
	assert.Nil(t, err)
	assert.False(t, c.IsIndex())
}

func TestPathComponent_IsIndex(t *testing.T) {

	// NewPathComponentFromIndex

	c, err := NewPathComponentFromIndex(10)
	assert.Nil(t, err)
	assert.True(t, c.IsIndex())

	c, err = NewPathComponentFromIndex(0)
	assert.Nil(t, err)
	assert.True(t, c.IsIndex())
}

func TestPathComponent_IsParent_Fail(t *testing.T) {

	// NewPathComponentFromIndex

	c, err := NewPathComponentFromIndex(0)
	assert.Nil(t, err)
	assert.False(t, c.IsParent())

	// NewPathComponentFromName

	c, err = NewPathComponentFromName("name")
	assert.Nil(t, err)
	assert.False(t, c.IsParent())
}

func TestPathComponent_IsParent(t *testing.T) {

	// NewPathComponentFromName

	c, err := NewPathComponentFromName(PathParentID)
	assert.Nil(t, err)
	assert.True(t, c.IsParent())
}

func TestPathComponent_Name_Fail(t *testing.T) {

	// NewPathComponentFromIndex

	c, err := NewPathComponentFromIndex(0)
	assert.Nil(t, err)
	assert.Equal(t, "", c.Name())

	c, err = NewPathComponentFromIndex(10)
	assert.Nil(t, err)
	assert.Equal(t, "", c.Name())
}

func TestPathComponent_Name(t *testing.T) {

	// NewPathComponentFromName

	c, err := NewPathComponentFromName("name")
	assert.Nil(t, err)
	assert.Equal(t, "name", c.Name())
}

func TestPathComponent_String(t *testing.T) {

	// NewPathComponentFromIndex

	c, err := NewPathComponentFromIndex(0)
	assert.Nil(t, err)
	assert.Equal(t, "0", c.String())

	// NewPathComponentFromName

	c, err = NewPathComponentFromName("name")
	assert.Nil(t, err)
	assert.Equal(t, "name", c.String())

	c, err = NewPathComponentFromName(PathParentID)
	assert.Nil(t, err)
	assert.Equal(t, PathParentID, c.String())
}

func TestPathComponent_Equality_Fail(t *testing.T) {

	// NewPathComponentFromIndex

	c1, err1 := NewPathComponentFromIndex(0)
	c2, err2 := NewPathComponentFromIndex(10)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.False(t, c1 == c2)

	// NewPathComponentFromName

	c1, err1 = NewPathComponentFromName("name")
	c2, err2 = NewPathComponentFromName("name_different")
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.False(t, c1 == c2)

	// NewPathComponentFromIndex/NewPathComponentFromName

	c1, err1 = NewPathComponentFromName("name")
	c2, err2 = NewPathComponentFromIndex(10)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.False(t, c1 == c2)
}

func TestPathComponent_Equality(t *testing.T) {

	// NewPathComponentFromIndex

	c1, err1 := NewPathComponentFromIndex(0)
	c2, err2 := NewPathComponentFromIndex(0)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.True(t, c1 == c2)

	c1, err1 = NewPathComponentFromIndex(10)
	c2, err2 = NewPathComponentFromIndex(10)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.True(t, c1 == c2)

	// NewPathComponentFromName

	c1, err1 = NewPathComponentFromName("name")
	c2, err2 = NewPathComponentFromName("name")
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.True(t, c1 == c2)
}
