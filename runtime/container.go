package runtime

//
//var _ NamedContent = (*Container)(nil)
//var _ Object = (*Container)(nil)
//
//type CountFlag int
//
//const (
//	CountFlagVisits         CountFlag = 1
//	CountFlagTurns          CountFlag = 2
//	CountFlagCountStartOnly CountFlag = 4
//)
//
//type Container struct {
//	*ObjectImpl
//	name                           string
//	content                        []Object
//	namedContent                   map[string]NamedContent
//	namedOnlyContent               map[string]NamedContent
//	visitsShouldBeCounted          bool
//	turnIndexShouldBeCounted       bool
//	countingAtStartOnly            bool
//	countFlags                     CountFlag
//	pathToFirstLeafContent         *Path
//	internalPathToFirstLeafContent *Path
//}
//
//func (s *Container) InternalPathToFirstLeafContent() *Path {
//	var components []*PathComponent
//	container := s
//	for container != nil {
//		if len(container.Content()) > 0 {
//			components = append(components, newPathComponentFromIndex(0))
//			con, ok := container.content[0].(*Container)
//			if !ok {
//				panic("should explicitly cast to container like the c# version")
//			}
//			container = con
//		}
//	}
//	return NewPathFromComponents(components)
//}
//
//func (s *Container) ContentAtPathWithPathStartPathLength(path *Path, partialPathStart int /*= 0*/, partialPathLength int /*= -1*/) SearchResult {
//
//	if partialPathLength == -1 {
//		partialPathLength = path.Length()
//	}
//
//	result := SearchResult{Approximate: false}
//
//	var currentContainer *Container = s
//	var currentObj Object = s
//
//	for i := partialPathStart + 1; i < partialPathLength; i++ {
//		var comp = path.Component(i)
//
//		// Path component was wrong type
//		if currentContainer == nil {
//			result.Approximate = true
//			break
//		}
//
//		var foundObj = currentContainer.contentWithPathComponent(comp)
//
//		// Couldn't resolve entire path?
//		if foundObj == nil {
//			result.Approximate = true
//			break
//		}
//
//		currentObj = foundObj
//		castContainer, ok := foundObj.(*Container)
//		if ok {
//			currentContainer = castContainer
//		}
//	}
//
//	result.Obj = currentObj
//
//	return result
//}
//
//func (s *Container) ContentAtPathWithPathStart(path *Path, partialPathStart int /*= 0*/) SearchResult {
//	return s.ContentAtPathWithPathStartPathLength(path, partialPathStart, -1)
//}
//
//func (s *Container) ContentAtPath(path *Path) SearchResult {
//	return s.ContentAtPathWithPathStartPathLength(path, 0, -1)
//}
//
//func (s *Container) AddContent(contentObjects ...Object) {
//
//	for _, contentObj := range contentObjects {
//
//		s.content = append(s.content, contentObj)
//
//		if contentObj.Parent() != nil {
//			// This threw an exception before (in the c# version)
//			panic(fmt.Sprintf("content is already in %v", getTypeName(contentObj.Parent())))
//		}
//
//		contentObj.SetParent(s)
//
//		s.TryAddNamedContent(contentObj)
//	}
//}
//
//func (s *Container) InsertContent(contentObj Object, index int) {
//
//	// Insert at index
//	s.content = append(s.content[:index+1], s.content[index:]...)
//	s.content[index] = contentObj
//
//	if contentObj.Parent() != nil {
//		// This threw an exception before (in the c# version)
//		panic(fmt.Sprintf("content is already in %v", getTypeName(contentObj.Parent())))
//	}
//
//	contentObj.SetParent(s)
//
//	s.TryAddNamedContent(contentObj)
//}
//
//func (s *Container) TryAddNamedContent(contentObj Object) {
//
//	var namedContentObj, ok = contentObj.(NamedContent)
//	if ok && namedContentObj.HasValidName() {
//		s.AddToNamedContentOnly(namedContentObj)
//	}
//}
//
//func (s *Container) contentWithPathComponent(component *PathComponent) Object {
//	if component.IsIndex() {
//
//		if component.Index() >= 0 && component.Index() < len(s.content) {
//			return s.content[component.Index()]
//		} else {
//			// When path is out of range, quietly return nil
//			// (useful as we step/increment forwards through content)
//			return nil
//		}
//
//	} else if component.IsParent() {
//		return s
//	} else {
//		if v, ok := s.namedContent[component.Name()]; ok {
//			obj, ok := v.(Object)
//			if !ok {
//				panic("should cast to object explicitly")
//			}
//			return obj
//		} else {
//			return nil
//		}
//	}
//}
//
//func (s *Container) AddToNamedContentOnly(namedContentObj NamedContent) {
//
//	if _, ok := namedContentObj.(Object); !ok {
//		panic("Can only add Runtime.Objects to a Runtime.Container")
//	}
//
//	runtimeObj := namedContentObj.(Object)
//	runtimeObj.SetParent(runtimeObj)
//	s.namedContent[namedContentObj.Name()] = namedContentObj
//}
//
//func (s *Container) AddContentsOfContainer(otherContainer *Container) {
//
//	s.content = append(s.content, otherContainer.Content()...)
//	for _, obj := range otherContainer.Content() {
//		obj.SetParent(s)
//		s.TryAddNamedContent(obj)
//	}
//}
//
//// TODO: implement path (s.path) located in Object
//
////func (s *Container) PathToFirstLeafContent() *Path {
////
////	if s.pathToFirstLeafContent == nil {
////		s.pathToFirstLeafContent = s.path.PathByAppendingPath(s.internalPathToFirstLeafContent)
////	}
////
////	return s.pathToFirstLeafContent
////}
//
//func (s *Container) CountFlags() int {
//
//	var flags CountFlag = 0
//	if s.visitsShouldBeCounted {
//		flags |= CountFlagVisits
//	}
//	if s.turnIndexShouldBeCounted {
//		flags |= CountFlagTurns
//	}
//	if s.countingAtStartOnly {
//		flags |= CountFlagCountStartOnly
//	}
//
//	//// If we're only storing CountStartOnly, it serves no purpose,
//	//// since it's dependent on the other two to be used at all.
//	//// (e.g. for setting the fact that *if* a gather or choice's
//	//// content is counted, then is should only be counter at the start)
//	//// So this is just an optimisation for storage.
//	if flags == CountFlagCountStartOnly {
//		flags = 0
//	}
//
//	return int(flags)
//}
//
//func (s *Container) SetCountFlags(n int) {
//
//	var flag = CountFlag(n)
//	if (flag & CountFlagVisits) > 0 {
//		s.visitsShouldBeCounted = true
//	}
//	if (flag & CountFlagTurns) > 0 {
//		s.turnIndexShouldBeCounted = true
//	}
//	if (flag & CountFlagCountStartOnly) > 0 {
//		s.countingAtStartOnly = true
//	}
//}
//
//func (s *Container) VisitsShouldBeCounted() bool {
//	return s.visitsShouldBeCounted
//}
//
//func (s *Container) SetVisitsShouldBeCounted(state bool) {
//	s.visitsShouldBeCounted = state
//}
//
//func (s *Container) TurnIndexShouldBeCounted() bool {
//	return s.turnIndexShouldBeCounted
//}
//
//func (s *Container) SetTurnIndexShouldBeCounted(state bool) {
//	s.turnIndexShouldBeCounted = state
//}
//
//func (s *Container) CountingAtStartOnly() bool {
//	return s.countingAtStartOnly
//}
//
//func (s *Container) SetCountingAtStartOnly(state bool) {
//	s.countingAtStartOnly = state
//}
//
//func (s *Container) Name() string {
//	return s.name
//}
//
//func (s *Container) HasValidName() bool {
//	if len(s.Name()) > 0 {
//		return true
//	}
//	return false
//}
//
//func (s *Container) NamedContent() map[string]NamedContent {
//	return s.namedContent
//}
//
//func (s *Container) SetNamedContent(namedContent map[string]NamedContent) {
//	s.namedContent = namedContent
//}
//
//func (s *Container) NamedOnlyContent() map[string]Object {
//	namedOnlyContentMap := map[string]Object{}
//
//	for k, v := range s.namedContent {
//		object, ok := v.(Object)
//		if !ok {
//			panic("what? This should be an object because in the original c# source code this is an explicit cast")
//		}
//		namedOnlyContentMap[k] = object
//	}
//
//	for _, c := range s.content {
//		named, ok := c.(NamedContent)
//		if ok && named.HasValidName() {
//			delete(namedOnlyContentMap, named.Name())
//		}
//	}
//
//	if len(namedOnlyContentMap) == 0 {
//		namedOnlyContentMap = nil
//	}
//
//	return namedOnlyContentMap
//}
//
//func (s *Container) SetNamedOnlyContent(namedOnlyContent map[string]Object) {
//
//	var existingNamedOnly = s.namedOnlyContent
//	if existingNamedOnly != nil {
//		for k := range existingNamedOnly {
//			delete(s.namedContent, k)
//		}
//	}
//
//	if namedOnlyContent == nil {
//		return
//	}
//
//	for _, v := range namedOnlyContent {
//		named, ok := v.(NamedContent)
//		if ok {
//			s.AddToNamedContentOnly(named)
//		}
//	}
//}
//
//func (s *Container) Content() []Object {
//	return s.content
//}
//
//func (s *Container) SetContent(content []Object) {
//	s.AddContent(content...)
//}

// TODO: depends on Value types in Value (C#)

/*
func (s *Container) BuildStringOfHierarchy(sb strings.Builder, indentation int, pointedObj Object)  {
	appendIndentation := func() {
spacesPerIndent := 4
for i := 1; i<spacesPerIndent*indentation; i++ {
sb.WriteString(" ")
}
}

appendIndentation ()
sb.WriteString("[");

if s.HasValidName() {
sb.WriteString (fmt.Sprintf(" (%s)", s.Name()))
}

if s == pointedObj {
sb.WriteString ("  <---")
}

sb.WriteRune ('\n')

indentation++

for i := 1; i<len(s.content); i++ {

var obj = s.content [i]

if v, ok := obj.(*Container); ok {
v.BuildStringOfHierarchy (sb, indentation, pointedObj)

} else {
appendIndentation ()
if _, ok := obj.(StringValue) {
sb.WriteString ("\"")
sb.WriteString (obj.ToString ().Replace ("\n", "\\n"));
sb.WriteString ("\"")
} else {
sb.Append (obj.ToString ());
}
}

if (i != content.Count - 1) {
sb.Append (",");
}

if ( !(obj is Container) && obj == pointedObj ) {
sb.Append ("  <---");
}

sb.AppendLine ();
}


var onlyNamed = new Dictionary<string, INamedContent> ();

foreach (var objKV in namedContent) {
if (content.Contains ((Runtime.Object)objKV.Value)) {
continue;
} else {
onlyNamed.Add (objKV.Key, objKV.Value);
}
}

if (onlyNamed.Count > 0) {
appendIndentation ();
sb.AppendLine ("-- named: --");

foreach (var objKV in onlyNamed) {

Debug.Assert (objKV.Value is Container, "Can only print out named Containers");
var container = (Container)objKV.Value;
container.BuildStringOfHierarchy (sb, indentation, pointedObj);

sb.AppendLine ();

}
}


indentation--;

appendIndentation ();
sb.Append ("]");
}*/

/*

public class Container : Runtime.Object, INamedContent
	{


		public Path pathToFirstLeafContent
		{
			get {
                if( _pathToFirstLeafContent == null )
                    _pathToFirstLeafContent = path.PathByAppendingPath (internalPathToFirstLeafContent);

                return _pathToFirstLeafContent;
			}
		}
        Path _pathToFirstLeafContent;


		public Container ()
		{
            _content = new List<Runtime.Object> ();
			namedContent = new Dictionary<string, INamedContent> ();
		}


        public void BuildStringOfHierarchy(StringBuilder sb, int indentation, Runtime.Object pointedObj)
        {
            Action appendIndentation = () => {
                const int spacesPerIndent = 4;
                for(int i=0; i<spacesPerIndent*indentation;++i) {
                    sb.Append(" ");
                }
            };

            appendIndentation ();
            sb.Append("[");

            if (this.hasValidName) {
                sb.AppendFormat (" ({0})", this.name);
            }

            if (this == pointedObj) {
                sb.Append ("  <---");
            }

            sb.AppendLine ();

            indentation++;

            for (int i=0; i<content.Count; ++i) {

                var obj = content [i];

                if (obj is Container) {

                    var container = (Container)obj;

                    container.BuildStringOfHierarchy (sb, indentation, pointedObj);

                } else {
                    appendIndentation ();
                    if (obj is StringValue) {
                        sb.Append ("\"");
                        sb.Append (obj.ToString ().Replace ("\n", "\\n"));
                        sb.Append ("\"");
                    } else {
                        sb.Append (obj.ToString ());
                    }
                }

                if (i != content.Count - 1) {
                    sb.Append (",");
                }

                if ( !(obj is Container) && obj == pointedObj ) {
                    sb.Append ("  <---");
                }

                sb.AppendLine ();
            }


            var onlyNamed = new Dictionary<string, INamedContent> ();

            foreach (var objKV in namedContent) {
                if (content.Contains ((Runtime.Object)objKV.Value)) {
                    continue;
                } else {
                    onlyNamed.Add (objKV.Key, objKV.Value);
                }
            }

            if (onlyNamed.Count > 0) {
                appendIndentation ();
                sb.AppendLine ("-- named: --");

                foreach (var objKV in onlyNamed) {

                    Debug.Assert (objKV.Value is Container, "Can only print out named Containers");
                    var container = (Container)objKV.Value;
                    container.BuildStringOfHierarchy (sb, indentation, pointedObj);

                    sb.AppendLine ();

                }
            }


            indentation--;

            appendIndentation ();
            sb.Append ("]");
        }

        public virtual string BuildStringOfHierarchy()
        {
            var sb = new StringBuilder ();

            BuildStringOfHierarchy (sb, 0, null);

            return sb.ToString ();
        }

	}

*/
