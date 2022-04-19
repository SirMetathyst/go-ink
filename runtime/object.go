package runtime

// Source: https://github.com/inkle/ink/blob/master/ink-engine-runtime/Object.cs#L10
// Go favours composition over inheritance and does not support it, so we use an interface
// instead and embed the implementation of the interface in our "subtype".

// Object ...
type Object interface {
	DebugMetadata() *DebugMetadata
	SetDebugMetadata(*DebugMetadata)
	OwnDebugMetadata() *DebugMetadata

	// Runtime.Objects can be included in the main Story as a hierarchy.
	// Usually parents are Container objects. (TODO: Always?)

	Parent() Object
	SetParent(o Object)

	RootContentContainer() *Container

	// Source: https://github.com/inkle/ink/blob/master/ink-engine-runtime/Object.cs#L46
	// Original: int? DebugLineNumberOfPath(Path)
	// Go does not have nullable types, the convention in go
	// is to provide the return value followed by a boolean such as:
	// lineNumber, ok := DebugLineNumberOfPath(path)

	// DebugLineNumberOfPath ...
	DebugLineNumberOfPath(p *Path) (int, bool)
}

var _ Object = (*ObjectImpl)(nil)

// Note: Object(C# source) and by extension ObjectImpl
// reference Container but Container is a subtype of Object...
// TODO: Remove dependency?

type ObjectImpl struct {
	parent Object

	// TODO(From C# Source): Come up with some clever solution for not having
	// to have debug metadata on the object itself, perhaps
	// for serialisation purposes at least.

	debugMetadata    *DebugMetadata
	ownDebugMetadata *DebugMetadata
}

func (s *ObjectImpl) Parent() Object {
	return s.parent
}

func (s *ObjectImpl) SetParent(o Object) {
	s.parent = o
}

func (s *ObjectImpl) DebugMetadata() *DebugMetadata {
	if s.debugMetadata == nil && s.parent != nil {
		return s.parent.DebugMetadata()
	}
	return s.debugMetadata
}

func (s *ObjectImpl) SetDebugMetadata(metadata *DebugMetadata) {
	s.debugMetadata = metadata
}

func (s *ObjectImpl) OwnDebugMetadata() *DebugMetadata {
	return s.ownDebugMetadata
}

func (s *ObjectImpl) RootContentContainer() *Container {
	var ancestor Object = s
	for ancestor.Parent() != nil {
		ancestor = ancestor.Parent()

	}
	if container, ok := ancestor.(*Container); ok {
		return container
	}
	return nil
}

func (s *ObjectImpl) DebugLineNumberOfPath(path *Path) (int, bool) {

	if path == nil {
		return -1, false
	}

	// Try to get a line number from debug metadata
	root := s.RootContentContainer()
	if root != nil {
		var targetContent Object = root.ContentAtPath(path).Obj
		if targetContent != nil {
			var dm = targetContent.DebugMetadata()
			if dm != nil {
				return dm.StartLineNumber, true
			}
		}
	}

	return -1, false
}

/*
 /// <summary>
    /// Base class for all ink runtime content.
    /// </summary>
    public class Object
{

public int? DebugLineNumberOfPath(Path path)
{
if (path == null)
return null;

// Try to get a line number from debug metadata
var root = this.rootContentContainer;
if (root) {
Runtime.Object targetContent = root.ContentAtPath (path).obj;
if (targetContent) {
var dm = targetContent.debugMetadata;
if (dm != null) {
return dm.startLineNumber;
}
}
}

return null;
}

public Path path
{
get
{
if (_path == null) {

if (parent == null) {
_path = new Path ();
} else {
// Maintain a Stack so that the order of the components
// is reversed when they're added to the Path.
// We're iterating up the hierarchy from the leaves/children to the root.
var comps = new Stack<Path.Component> ();

var child = this;
Container container = child.parent as Container;

while (container) {

var namedChild = child as INamedContent;
if (namedChild != null && namedChild.hasValidName) {
comps.Push (new Path.Component (namedChild.name));
} else {
comps.Push (new Path.Component (container.content.IndexOf(child)));
}

child = container;
container = container.parent as Container;
}

_path = new Path (comps);
}

}

return _path;
}
}
Path _path;

public SearchResult ResolvePath(Path path)
{
if (path.isRelative) {

Container nearestContainer = this as Container;
if (!nearestContainer) {
Debug.Assert (this.parent != null, "Can't resolve relative path because we don't have a parent");
nearestContainer = this.parent as Container;
Debug.Assert (nearestContainer != null, "Expected parent to be a container");
Debug.Assert (path.GetComponent(0).isParent);
path = path.tail;
}

return nearestContainer.ContentAtPath (path);
} else {
return this.rootContentContainer.ContentAtPath (path);
}
}

public Path ConvertPathToRelative(Path globalPath)
{
// 1. Find last shared ancestor
// 2. Drill up using ".." style (actually represented as "^")
// 3. Re-build downward chain from common ancestor

var ownPath = this.path;

int minPathLength = Math.Min (globalPath.length, ownPath.length);
int lastSharedPathCompIndex = -1;

for (int i = 0; i < minPathLength; ++i) {
var ownComp = ownPath.GetComponent(i);
var otherComp = globalPath.GetComponent(i);

if (ownComp.Equals (otherComp)) {
lastSharedPathCompIndex = i;
} else {
break;
}
}

// No shared path components, so just use global path
if (lastSharedPathCompIndex == -1)
return globalPath;

int numUpwardsMoves = (ownPath.length-1) - lastSharedPathCompIndex;

var newPathComps = new List<Path.Component> ();

for(int up=0; up<numUpwardsMoves; ++up)
newPathComps.Add (Path.Component.ToParent ());

for (int down = lastSharedPathCompIndex + 1; down < globalPath.length; ++down)
newPathComps.Add (globalPath.GetComponent(down));

var relativePath = new Path (newPathComps, relative:true);
return relativePath;
}

// Find most compact representation for a path, whether relative or global
public string CompactPathString(Path otherPath)
{
string globalPathStr = null;
string relativePathStr = null;
if (otherPath.isRelative) {
relativePathStr = otherPath.componentsString;
globalPathStr = this.path.PathByAppendingPath(otherPath).componentsString;
} else {
var relativePath = ConvertPathToRelative (otherPath);
relativePathStr = relativePath.componentsString;
globalPathStr = otherPath.componentsString;
}

if (relativePathStr.Length < globalPathStr.Length)
return relativePathStr;
else
return globalPathStr;
}

public Container rootContentContainer
{
get
{
Runtime.Object ancestor = this;
while (ancestor.parent) {
ancestor = ancestor.parent;
}
return ancestor as Container;
}
}

public Object ()
{
}

public virtual Object Copy()
{
throw new System.NotImplementedException (GetType ().Name + " doesn't support copying");
}

public void SetChild<T>(ref T obj, T value) where T : Runtime.Object
{
if (obj)
obj.parent = null;

obj = value;

if( obj )
obj.parent = this;
}

/// Allow implicit conversion to bool so you don't have to do:
/// if( myObj != null ) ...
public static implicit operator bool (Object obj)
{
var isNull = object.ReferenceEquals (obj, null);
return !isNull;
}

/// Required for implicit bool comparison
public static bool operator ==(Object a, Object b)
{
return object.ReferenceEquals (a, b);
}

/// Required for implicit bool comparison
public static bool operator !=(Object a, Object b)
{
return !(a == b);
}

/// Required for implicit bool comparison
public override bool Equals (object obj)
{
return object.ReferenceEquals (obj, this);
}

/// Required for implicit bool comparison
public override int GetHashCode ()
{
return base.GetHashCode ();
}
}
*/
