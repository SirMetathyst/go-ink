package runtime

// When looking up content within the story (e.g. in Container.ContentAtPath),
// the result is generally found, but if the story is modified, then when loading
// up an old save state, then some old paths may still exist. In this case we
// try to recover by finding an approximate result by working up the story hierarchy
// in the path to find the closest valid container. Instead of crashing horribly,
// we might see some slight oddness in the content, but hopefully it recovers!

type SearchResult struct {
	Obj         Object
	Approximate bool
}

func NewSearchResult() SearchResult {
	return SearchResult{}
}

func (s SearchResult) CorrectObj() Object {

	if s.Approximate {
		return nil
	}

	return s.Obj
}

func (s SearchResult) Container() *Container {

	container, _ := s.Obj.(*Container)
	return container
}
