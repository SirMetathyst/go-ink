package runtime

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strings"
)

// InkVersionCurrent
// The current version of the ink story file format.
const InkVersionCurrent = 21

// Version numbers are for engine itself and story file, rather
// than the story state save format
//  -- old engine, new format: always fail
//  -- new engine, old format: possibly cope, based on this number
// When incrementing the version number above, the question you
// should ask yourself is:
//  -- Will the engine be able to load an old story file from
//     before I made these changes to the engine?
//     If possible, you should support it, though it's not as
//     critical as loading old save games, since it's an
//     in-development problem only.

// The minimum legacy version of ink that can be loaded by the current version of the code.
const inkVersionMinimumCompatible = 10

type OnEvaluateFunction func(functionName string, arguments []interface{})

type OnChoosePathString func(path string, arguments []interface{})

type OnCompleteEvaluateFunction func(functionName string, arguments []interface{}, textOutput string, result interface{})

type VariableObserver func(variableName, newValue interface{})

// Assumption: prevText is the snapshot where we saw a newline, and we're checking whether we're really done
//             with that line. Therefore prevText will definitely end in a newline.
//
// We take tags into account too, so that a tag following a content line:
//   Content
//   # tag
// ... doesn't cause the tag to be wrongly associated with the content above.
type OutputStateChange int

const (
	NoChange OutputStateChange = iota
	ExtendedBeyondNewline
	NewlineRemoved
)

// Story
// A Story is the core class that represents a complete Ink narrative, and
// manages the evaluation and state of it.
type Story struct {
	ObjectImpl

	// Public

	// Error handler for all runtime errors in ink - i.e. problems
	// with the source ink itself that are only discovered when playing
	// the story.
	// It's strongly recommended that you assign an error handler to your
	// story instance to avoid getting exceptions for ink errors.
	OnError *ErrorHandlerEvent

	// Callback for when ContinueInternal is complete
	OnDidContinue *ActionEvent

	// Callback for when a choice is about to be executed
	OnMakeChoice *ActionT1Event[*Choice]

	// Callback for when a function is about to be evaluated
	OnEvaluateFunction *OnEvaluateFunctionEvent

	// Callback for when a function has been evaluated
	// This is necessary because evaluating a function can cause continuing
	OnCompleteEvaluateFunction *OnCompleteEvaluateFunctionEvent

	// Callback for when a path string is chosen
	OnChoosePathString *OnChoosePathStringEvent

	// An ink file can provide a fallback functions for when when an EXTERNAL has been left
	// unbound by the client, and the fallback function will be called instead. Useful when
	// testing a story in playmode, when it's not possible to write a client-side C# external
	// function, but you don't want it to fail to run.
	AllowExternalFunctionFallbacks bool

	// Private
	_mainContentContainer                   *Container
	_listDefinitions                        *ListDefinitionsOrigin
	_externals                              map[string]*ExternalFunctionDef
	_variableObservers                      map[string]VariableObserver
	_hasValidatedExternals                  bool
	_temporaryEvaluationContainer           *Container
	_state                                  *StoryState
	_asyncContinueActive                    bool
	_stateSnapshotAtLastNewline             *StoryState // set null in class def
	_sawLookaheadUnsafeFunctionAfterNewline bool        // set false in class def
	_recursiveContinueCount                 int         // set 0 in class def
	_asyncSaving                            bool
	_prevContainers                         []*Container
}

func (s *Story) CurrentChoices() []*Choice {

	// Don't include invisible choices for external usage.
	choices := []*Choice{}
	for _, c := range s._state.CurrentChoices() {
		if !c.IsInvisibleDefault {
			c.Index = len(choices)
			choices = append(choices, c)
		}
	}
	return choices
}

// CurrentText
// The latest line of text to be generated from a Continue() call.
func (s *Story) CurrentText() string {
	s.IfAsyncWeCant("call currentText since it's a work in progress")
	return s._state.CurrentText()
}

// CurrentTags
// Gets a list of tags as defined with '#' in source that were seen
// during the latest Continue() call.
func (s *Story) CurrentTags() []string {
	s.IfAsyncWeCant("call currentTags since it's a work in progress")
	return s._state.CurrentTags()
}

// CurrentErrors
// Any errors generated during evaluation of the Story.
func (s *Story) CurrentErrors() []string {
	return s._state.CurrentErrors()
}

// HasWarning
// Whether the currentWarnings list contains any warnings.
func (s *Story) HasWarning() bool {
	return s._state.HasWarning()
}

// VariablesState
// The VariablesState object contains all the global variables in the story.
// However, note that there's more to the state of a Story than just the
// global variables. This is a convenience accessor to the full state object.
func (s *Story) VariablesState() *VariablesState {
	return s._state.VariablesState()
}

func (s *Story) ListDefinitions() *ListDefinitionsOrigin {
	return s._listDefinitions
}

// State
// The entire current state of the story including (but not limited to):
//
//  * Global variables
//  * Temporary variables
//  * Read/visit and turn counts
//  * The callstack and evaluation stacks
//  * The current threads
func (s *Story) State() *StoryState {
	return s._state
}

/*
        /// <summary>
        /// Start recording ink profiling information during calls to Continue on Story.
        /// Return a Profiler instance that you can request a report from when you're finished.
        /// </summary>
		public Profiler StartProfiling() {
            IfAsyncWeCant ("start profiling");
			_profiler = new Profiler();
			return _profiler;
		}
*/

/*
        /// <summary>
        /// Stop recording ink profiling information during calls to Continue on Story.
        /// To generate a report from the profiler, call
        /// </summary>
		public void EndProfiling() {
			_profiler = null;
		}
*/

// NewStoryFrom
// Warning: When creating a Story using this constructor, you need to
// call ResetState on it before use. Intended for compiler use only.
// For normal use, use the constructor that takes a json string.
// (default) lists: nil
func NewStoryFrom(contentContainer *Container, lists []*ListDefinition) *Story {

	newStory := new(Story)
	newStory._prevContainers = []*Container{}
	newStory._mainContentContainer = contentContainer

	if lists != nil {
		newStory._listDefinitions = NewListDefinitionsOrigin(lists)
	}

	newStory._externals = make(map[string]*ExternalFunctionDef)

	return newStory
}

// NewStory
// Construct a Story object using a JSON string compiled through inklecate.
func NewStory(jsonString string) *Story {

	newStory := new(Story)
	newStory._prevContainers = []*Container{}

	rootObject := TextToDictionary(jsonString)

	versionObj := rootObject["inkVersion"]
	if versionObj == nil {
		panic("ink version number not found. Are you sure it's a valid .ink.json file?")
	}

	formatFromFile := versionObj.(int)
	if formatFromFile > InkVersionCurrent {
		panic("Version of ink used to build story was newer than the current version of the engine")
	}

	if formatFromFile < inkVersionMinimumCompatible {
		panic("Version of ink used to build story is too old to be loaded by this version of the engine")
	}

	if formatFromFile != InkVersionCurrent {
		fmt.Println("WARNING: Version of ink used to build story doesn't match current version of engine. Non-critical, but recommend synchronising.")
	}

	rootToken := rootObject["root"]
	if rootToken == nil {
		panic("Root node for ink not found. Are you sure it's a valid .ink.json file?")
	}

	if listDefsObj, ok := rootObject["listDefs"]; ok {
		newStory._listDefinitions = JTokenToListDefinitions(listDefsObj)
	}

	newStory._mainContentContainer, _ = JTokenToRuntimeObject(rootToken).(*Container)
	newStory.ResetState()

	return newStory
}

// ToJson
// The Story itself in JSON representation.
func (s *Story) ToJson() string {
	writer := NewWriter()
	s.ToJsonWriter(writer)
	return writer.String()
}

func (s *Story) ToJsonWriter(writer *Writer) {

	writer.WriteObjectStart()

	writer.WriteIntProperty("inkVersion", InkVersionCurrent)

	// Main container content
	writer.WritePropertyStart("root")
	WriteRuntimeContainer(writer, s._mainContentContainer, false)
	writer.WritePropertyEnd()

	// List definitions
	if s._listDefinitions != nil {

		writer.WritePropertyStart("listDefs")
		writer.WriteObjectStart()

		for _, def := range s._listDefinitions.Lists() {

			writer.WritePropertyStart(def.Name())
			writer.WriteObjectStart()

			for item, val := range def.Items() {
				writer.WriteIntProperty(item.ItemName(), val)
			}

			writer.WriteObjectEnd()
			writer.WritePropertyEnd()
		}

		writer.WriteObjectEnd()
		writer.WritePropertyEnd()
	}

	writer.WriteObjectEnd()
}

func (s *Story) ResetState() {

	// TODO: Could make this possible
	s.IfAsyncWeCant("ResetState")

	s._state = NewStoryState(s)
	s._state.VariablesState().VariableChangedEvent = new(VariableChangedEvent)
	s._state.VariablesState().VariableChangedEvent.Register(s.VariableStateDidChangeEvent)

	s.ResetGlobals()
}

func (s *Story) ResetErrors() {
	s._state.ResetErrors()
}

// ResetCallstack
// Unwinds the callstack. Useful to reset the Story's evaluation
// without actually changing any meaningful state, for example if
// you want to exit a section of story prematurely and tell it to
// go elsewhere with a call to ChoosePathString(...).
// Doing so without calling ResetCallstack() could cause unexpected
// issues if, for example, the Story was in a tunnel already.
func (s *Story) ResetCallstack() {

	s.IfAsyncWeCant("ResetCallstack")
	s._state.ForceEnd()
}

func (s *Story) ResetGlobals() {

	if _, ok := s._mainContentContainer.NamedContent()["global decl"]; ok {
		originalPointer := s._state.CurrentPointer()

		s.ChoosePath(NewPathFromString("global decl"), false)

		// Continue, but without validating external bindings,
		// since we may be doing this reset at initialisation time.
		s.ContinueInternal(0)

		s._state.SetCurrentPointer(originalPointer)
	}

	s._state.VariablesState().SnapshotDefaultGlobals()
}

func (s *Story) SwitchFlow(flowName string) {

	s.IfAsyncWeCant("switch flow")
	if s._asyncSaving {
		panic("Story is already in background saving mode, can't switch flow to " + flowName)
	}

	s._state.switchFlow_Internal(flowName)
}

// Continue
// Continue the story for one line of content, if possible.
// If you're not sure if there's more content available, for example if you
// want to check whether you're at a choice point or at the end of the story,
// you should call <c>canContinue</c> before calling this function.
func (s *Story) Continue() string {

	s.ContinueAsync(0)
	return s.CurrentText()
}

// CanContinue
// Check whether more content is available if you were to call <c>Continue()</c> - i.e.
// are we mid story rather than at a choice point or at the end.
func (s *Story) CanContinue() bool {

	return s._state.CanContinue()
}

// AsyncContinueComplete
// If ContinueAsync was called (with milliseconds limit > 0) then this property
// will return false if the ink evaluation isn't yet finished, and you need to call
// it again in order for the Continue to fully complete.
func (s *Story) AsyncContinueComplete() bool {

	return !s._asyncContinueActive
}

// ContinueAsync
// An "asnychronous" version of Continue that only partially evaluates the ink,
// with a budget of a certain time limit. It will exit ink evaluation early if
// the evaluation isn't complete within the time limit, with the
// asyncContinueComplete property being false.
// This is useful if ink evaluation takes a long time, and you want to distribute
// it over multiple game frames for smoother animation.
// If you pass a limit of zero, then it will fully evaluate the ink in the same
// way as calling Continue (and in fact, this exactly what Continue does internally).
func (s *Story) ContinueAsync(millisecsLimitAsync float64) {

	if !s._hasValidatedExternals {
		s.ValidateExternalBindings()
	}

	s.ContinueInternal(millisecsLimitAsync)
}

func (s *Story) ContinueInternal(millisecsLimitAsync float64) {

	// if( _profiler != null )
	// 	 _profiler.PreContinue(); C#

	isAsyncTimeLimited := millisecsLimitAsync > 0

	s._recursiveContinueCount++

	// Doing either:
	//  - full run through non-async (so not active and don't want to be)
	//  - Starting async run-through
	if !s._asyncContinueActive {
		s._asyncContinueActive = isAsyncTimeLimited

		if !s.CanContinue() {
			panic("Can't continue - should check canContinue before calling Continue")
		}

		s._state.DidSafeExit = false
		s._state.ResetOutput(nil)

		// It's possible for ink to call game to call ink to call game etc
		// In this case, we only want to batch observe variable changes
		// for the outermost call.
		if s._recursiveContinueCount == 1 {
			s._state.VariablesState().SetBatchObservingVariableChanges(true)
		}
	}

	// Start timing
	// var durationStopwatch = new Stopwatch ();
	// durationStopwatch.Start (); C#
	//durationStopwatch := time.Now()

	outputStreamEndsInNewline := false
	s._sawLookaheadUnsafeFunctionAfterNewline = false

	for do := true; do; do = s.CanContinue() {

		/*
		   try {
		       outputStreamEndsInNewline = ContinueSingleStep ();
		   } catch(StoryException e) {
		       AddError (e.Message, useEndLineNumber:e.useEndLineNumber);
		       break;
		   }
		*/
		outputStreamEndsInNewline = s.ContinueSingleStep()

		if outputStreamEndsInNewline {
			break
		}

		/*
		   // Run out of async time?
		              if (_asyncContinueActive && durationStopwatch.ElapsedMilliseconds > millisecsLimitAsync) {
		                  break;
		              }
		*/
	}

	//  durationStopwatch.Stop ();

	// 4 outcomes:
	//  - got newline (so finished this line of text)
	//  - can't continue (e.g. choices or ending)
	//  - ran out of time during evaluation
	//  - error
	//
	// Successfully finished evaluation in time (or in error)
	if outputStreamEndsInNewline || !s.CanContinue() {

		// Need to rewind, due to evaluating further than we should?
		if s._stateSnapshotAtLastNewline != nil {
			s.RestoreStateSnapshot()
		}

		// Finished a section of content / reached a choice point?
		if !s.CanContinue() {
			if s.State().CallStack().CanPopThread() {
				s.AddError("Thread available to pop, threads should always be flat by the end of evaluation?", false, false)
			}

			if len(s.State().GeneratedChoices()) == 0 && !s.State().DidSafeExit && s._temporaryEvaluationContainer == nil {
				if s.State().CallStack().CanPopWith(Tunnel) {
					s.AddError("unexpectedly reached end of content. Do you need a '->->' to return from a tunnel?", false, false)
				} else if s.State().CallStack().CanPopWith(Function) {
					s.AddError("unexpectedly reached end of content. Do you need a '~ return'?", false, false)
				} else if s.State().CallStack().CanPop() {
					s.AddError("ran out of content. Do you need a '-> DONE' or '-> END'?", false, false)
				} else {
					s.AddError("unexpectedly reached end of content for unknown reason. Please debug compiler!", false, false)
				}
			}
		}

		s.State().DidSafeExit = false
		s._sawLookaheadUnsafeFunctionAfterNewline = false

		if s._recursiveContinueCount == 1 {
			s._state.VariablesState().SetBatchObservingVariableChanges(false)
		}

		s._asyncContinueActive = false
		if s.OnDidContinue != nil {
			s.OnDidContinue.Emit()
		}
	}

	s._recursiveContinueCount--

	// if( _profiler != null )
	//     _profiler.PostContinue(); C#

	// Report any errors that occured during evaluation.
	// This may either have been StoryExceptions that were thrown
	// and caught during evaluation, or directly added with AddError.
	if s.State().HasError() || s.State().HasWarning() {
		// if onError != null C#
		if s.OnError != nil {
			if s._state.HasError() {
				for _, err := range s.State().CurrentErrors() {
					s.OnError.Emit(err, ErrorTypeError)
				}
			}
			if s.State().HasWarning() {
				for _, err := range s._state.CurrentWarnings() {
					s.OnError.Emit(err, ErrorTypeWarning)
				}
			}
			s.ResetErrors()
		} else {
			// Throw an exception since there's no error handler
			var sb strings.Builder
			sb.WriteString("Ink had ")
			if s.State().HasError() {
				sb.WriteString(fmt.Sprint(len(s.State().CurrentErrors())))
				if len(s.State().CurrentErrors()) == 1 {
					sb.WriteString(" error")
				} else {
					sb.WriteString(" errors")
				}
				if s.State().HasWarning() {
					sb.WriteString(" and ")
				}
			}
			if s.State().HasWarning() {
				sb.WriteString(fmt.Sprint(len(s.State().CurrentWarnings())))
				if len(s.State().CurrentWarnings()) == 1 {
					sb.WriteString(" warning")
				} else {
					sb.WriteString(" warnings")
				}
			}
			sb.WriteString(". It is strongly suggested that you assign an error handler to story.onError. The first issue was: ")
			if s.State().HasError() {
				sb.WriteString(s.State().CurrentErrors()[0])
			} else {
				sb.WriteString(s.State().CurrentWarnings()[0])
			}

			// If you get this exception, please assign an error handler to your story.
			// If you're using Unity, you can do something like this when you create
			// your story:
			//
			// var story = new Ink.Runtime.Story(jsonTxt);
			// story.onError = (errorMessage, errorType) => {
			//     if( errorType == ErrorType.Warning )
			//         Debug.LogWarning(errorMessage);
			//     else
			//         Debug.LogError(errorMessage);
			// };
			//
			//
			panic(sb.String())
		}
	}
}

func (s *Story) ContinueSingleStep() bool {

	//if (_profiler != null)
	//	_profiler.PreStep (); C#

	// Run main step function (walks through content)
	s.Step()

	//if (_profiler != null)
	//	_profiler.PostStep ();

	// Run out of content and we have a default invisible choice that we can follow?
	if !s.CanContinue() && !s.State().CallStack().ElementIsEvaluateFromGame() {
		s.TryFollowDefaultInvisibleChoice()
	}

	//if (_profiler != null)
	//	_profiler.PreSnapshot ();

	// Don't save/rewind during string evaluation, which is e.g. used for choices
	if !s.State().InStringEvaluation() {

		// We previously found a newline, but were we just double checking that
		// it wouldn't immediately be removed by glue?
		if s._stateSnapshotAtLastNewline != nil {

			// Has proper text or a tag been added? Then we know that the newline
			// that was previously added is definitely the end of the line.
			change := s.CalculateNewlineOutputStateChange(
				s._stateSnapshotAtLastNewline.CurrentText(), s.State().CurrentText(),
				len(s._stateSnapshotAtLastNewline.CurrentTags()), len(s.State().CurrentTags()))

			// The last time we saw a newline, it was definitely the end of the line, so we
			// want to rewind to that point.
			if change == ExtendedBeyondNewline || s._sawLookaheadUnsafeFunctionAfterNewline {
				s.RestoreStateSnapshot()

				// Hit a newline for sure, we're done
				return true
			}

			if change == NewlineRemoved {
				// Newline that previously existed is no longer valid - e.g.
				// glue was encounted that caused it to be removed.
				s.DiscardSnapshot()
			}
		}

		// Current content ends in a newline - approaching end of our evaluation
		if s.State().OutputStreamEndsInNewline() {

			// If we can continue evaluation for a bit:
			// Create a snapshot in case we need to rewind.
			// We're going to continue stepping in case we see glue or some
			// non-text content such as choices.
			if s.CanContinue() {

				// Don't bother to record the state beyond the current newline.
				// e.g.:
				// Hello world\n            // record state at the end of here
				// ~ complexCalculation()   // don't actually need this unless it generates text
				if s._stateSnapshotAtLastNewline == nil {
					s.StateSnapshot()
				}
			} else {
				// Can't continue, so we're about to exit - make sure we
				// don't have an old state hanging around.
				s.DiscardSnapshot()
			}
		}
	}

	//if (_profiler != null)
	//	_profiler.PostSnapshot ();

	// outputStreamEndsInNewline = false C# (commented out in original source)
	return false
}

func (s *Story) CalculateNewlineOutputStateChange(prevText string, currText string, prevTagCount int, currTagCount int) OutputStateChange {

	// Simple case: nothing's changed, and we still have a newline
	// at the end of the current content
	newlineStillExists := len(currText) >= len(prevText) && len(prevText) > 0 && currText[len(prevText)-1] == '\n'

	if prevTagCount == currTagCount && len(prevText) == len(currText) && newlineStillExists {
		return NoChange
	}

	// Old newline has been removed, it wasn't the end of the line after all
	if !newlineStillExists {
		return NewlineRemoved
	}

	// Tag added - definitely the start of a new line
	if currTagCount > prevTagCount {
		return ExtendedBeyondNewline
	}

	// There must be new content - check whether it's just whitespace
	for i := len(prevText); i < len(currText); i++ {
		c := currText[i]
		if c != ' ' && c != '\t' {
			return ExtendedBeyondNewline
		}
	}

	// There's new text but it's just spaces and tabs, so there's still the potential
	// for glue to kill the newline.
	return NoChange
}

// ContinueMaximally
// Continue the story until the next choice point or until it runs out of content.
// This is as opposed to the Continue() method which only evaluates one line of
// output at a time.
func (s *Story) ContinueMaximally() string {

	s.IfAsyncWeCant("ContinueMaximally")

	var sb strings.Builder

	for s.CanContinue() {
		sb.WriteString(s.Continue())
	}

	return sb.String()
}

func (s *Story) ContentAtPath(path *Path) SearchResult {

	return s.MainContentContainer().ContentAtPath(path, 0, -1)
}

func (s *Story) KnotContainerWithName(name string) *Container {

	if namedContainer, ok := s.MainContentContainer().NamedContent()[name]; ok {
		c, _ := namedContainer.(*Container)
		return c
	}

	return nil
}

func (s *Story) PointerAtPath(path *Path) Pointer {

	if path.Length() == 0 {
		return NullPointer
	}

	p := Pointer{}
	pathLengthToUse := path.Length()

	var result SearchResult

	if path.LastComponent().IsIndex() {
		pathLengthToUse = path.Length() - 1
		result = s.MainContentContainer().ContentAtPath(path, 0, pathLengthToUse)
		p.Container = result.Container()
		p.Index = path.LastComponent().Index()
	} else {
		result = s.MainContentContainer().ContentAtPath(path, 0, -1)
		p.Container = result.Container()
		p.Index = -1
	}

	if result.Obj == nil || result.Obj == s.MainContentContainer() && pathLengthToUse > 0 {
		s.Error("Failed to find content at path '" + path.String() + "', and no approximation of it was possible.")
	} else if result.Approximate {
		s.Warning("Failed to find content at path '" + path.String() + "', so it was approximated to: '" + result.Obj.Path(result.Obj).String() + "'.")
	}

	return p
}

// StateSnapshot
// Maximum snapshot stack:
//  - stateSnapshotDuringSave -- not retained, but returned to game code
//  - _stateSnapshotAtLastNewline (has older patch)
//  - _state (current, being patched)
func (s *Story) StateSnapshot() {
	s._stateSnapshotAtLastNewline = s._state
	s._state = s._state.CopyAndStartPatching()
}

func (s *Story) RestoreStateSnapshot() {

	// Patched state had temporarily hijacked our
	// VariablesState and set its own callstack on it,
	// so we need to restore that.
	// If we're in the middle of saving, we may also
	// need to give the VariablesState the old patch.
	s._stateSnapshotAtLastNewline.RestoreAfterPatch()

	s._state = s._stateSnapshotAtLastNewline
	s._stateSnapshotAtLastNewline = nil

	// If save completed while the above snapshot was
	// active, we need to apply any changes made since
	// the save was started but before the snapshot was made.
	if !s._asyncSaving {
		s._state.ApplyAnyPatch()
	}
}

func (s *Story) DiscardSnapshot() {

	// Normally we want to integrate the patch
	// into the main global/counts dictionaries.
	// However, if we're in the middle of async
	// saving, we simply stay in a "patching" state,
	// albeit with the newer cloned patch.
	if !s._asyncSaving {
		s._state.ApplyAnyPatch()
	}

	// No longer need the snapshot.
	s._stateSnapshotAtLastNewline = nil
}

// CopyStateForBackgroundThreadSave
// Advanced usage!
// If you have a large story, and saving state to JSON takes too long for your
// framerate, you can temporarily freeze a copy of the state for saving on
// a separate thread. Internally, the engine maintains a "diff patch".
// When you've finished saving your state, call BackgroundSaveComplete()
// and that diff patch will be applied, allowing the story to continue
// in its usual mode.
func (s *Story) CopyStateForBackgroundThreadSave() *StoryState {

	s.IfAsyncWeCant("start saving on a background thread")
	if s._asyncSaving {
		panic("Story is already in background saving mode, can't call CopyStateForBackgroundThreadSave again!")
	}
	stateToSave := s._state
	s._state = s._state.CopyAndStartPatching()
	s._asyncSaving = true
	return stateToSave
}

// BackgroundSaveComplete
// See CopyStateForBackgroundThreadSave. This method releases the
// "frozen" save state, applying its patch that it was using internally.
func (s *Story) BackgroundSaveComplete() {

	// CopyStateForBackgroundThreadSave must be called outside
	// of any async ink evaluation, since otherwise you'd be saving
	// during an intermediate state.
	// However, it's possible to *complete* the save in the middle of
	// a glue-lookahead when there's a state stored in _stateSnapshotAtLastNewline.
	// This state will have its own patch that is newer than the save patch.
	// We hold off on the final apply until the glue-lookahead is finished.
	// In that case, the apply is always done, it's just that it may
	// apply the looked-ahead changes OR it may simply apply the changes
	// made during the save process to the old _stateSnapshotAtLastNewline state.
	if s._stateSnapshotAtLastNewline == nil {
		s._state.ApplyAnyPatch()
	}

	s._asyncSaving = false
}

func (s *Story) Step() {

	shouldAddToStream := true

	// Get current content
	pointer := s.State().CurrentPointer()
	if pointer.IsNull() {
		return
	}

	// Step directly to the first element of content in a container (if necessary)
	containerToEnter, _ := pointer.Resolve().(*Container)
	for containerToEnter != nil {

		// Mark container as being entered
		s.VisitContainer(containerToEnter, true)

		// No content? the most we can do is step past it
		if len(containerToEnter.Content()) == 0 {
			break
		}

		pointer = StartOfPointer(containerToEnter)
		containerToEnter, _ = pointer.Resolve().(*Container)
	}

	s.State().SetCurrentPointer(pointer)

	//if( _profiler != null ) {
	//	_profiler.Step(state.callStack);
	//}

	// Is the current content object:
	//  - Normal content
	//  - Or a logic/flow statement - if so, do it
	// Stop flow if we hit a stack pop when we're unable to pop (e.g. return/done statement in knot
	// that was diverted to rather than called as a function)
	currentContentObj := pointer.Resolve()
	isLogicOrFlowControl := s.PerformLogicAndFlowControl(currentContentObj)

	// Has flow been forced to end by flow control above?
	if s.State().CurrentPointer().IsNull() {
		return
	}

	if isLogicOrFlowControl {
		shouldAddToStream = false
	}

	// Choice with condition?
	choicePoint, _ := currentContentObj.(*ChoicePoint)
	if choicePoint != nil {
		choice := s.ProcessChoice(choicePoint)
		if choice != nil {
			//state.generatedChoices.Add (choice);
			//s.State()._currentFlow.CurrentChoices = append(s.State()._currentFlow.CurrentChoices, choice)
			//s._state.GeneratedChoices().Add(choice)
			s._state._currentFlow.CurrentChoices = append(s._state._currentFlow.CurrentChoices, choice)
		}

		currentContentObj = nil
		shouldAddToStream = false
	}

	// If the container has no content, then it will be
	// the "content" itself, but we skip over it.
	if _, isContainer := currentContentObj.(*Container); isContainer {
		shouldAddToStream = false
	}

	// Content to add to evaluation stack or the output stream
	if shouldAddToStream {

		// If we're pushing a variable pointer onto the evaluation stack, ensure that it's specific
		// to our current (possibly temporary) context index. And make a copy of the pointer
		// so that we're not editing the original runtime object.
		varPointer, _ := currentContentObj.(*VariablePointerValue)
		if varPointer != nil && varPointer.ContextIndex() == -1 {

			// Create new object so we're not overwriting the story's own data
			contextIdx := s.State().CallStack().ContextForVariableNamed(varPointer.Value())
			currentContentObj = NewVariablePointerValueFromValue(varPointer.Value(), contextIdx)
		}

		// Expression evaluation content
		if s.State().InExpressionEvaluation() {
			s.State().PushEvaluationStack(currentContentObj)
		} else {
			// Output stream content (i.e. not expression evaluation)
			s.State().PushToOutputStream(currentContentObj)
		}
	}

	// Increment the content pointer, following diverts if necessary
	s.NextContent()

	// Starting a thread should be done after the increment to the content pointer,
	// so that when returning from the thread, it returns to the content after this instruction.
	controlCmd, _ := currentContentObj.(*ControlCommand)
	if controlCmd != nil && controlCmd.CommandType == CommandTypeStartThread {
		s.State().CallStack().PushThread()
	}
}

// VisitContainer
// Mark a container as having been visited
func (s *Story) VisitContainer(container *Container, atStart bool) {

	if !container.CountingAtStartOnly || atStart {
		if container.VisitsShouldBeCounted {
			s.State().IncrementVisitCountForContainer(container)
		}
		if container.TurnIndexShouldBeCounted {
			s.State().RecordTurnIndexVisitToContainer(container)
		}
	}
}

func (s *Story) VisitChangedContainersDueToDivert() {

	previousPointer := s.State().PreviousPointer()
	pointer := s.State().CurrentPointer()

	// Unless we're pointing *directly* at a piece of content, we don't do
	// counting here. Otherwise, the main stepping function will do the counting.
	if pointer.IsNull() || pointer.Index == -1 {
		return
	}

	// First, find the previously open set of containers
	s._prevContainers = s._prevContainers[:0]
	if !previousPointer.IsNull() {
		prevAncestor, _ := previousPointer.Resolve().(*Container)
		if prevAncestor == nil {
			prevAncestor = previousPointer.Container
		}
		for prevAncestor != nil {
			s._prevContainers = append(s._prevContainers, prevAncestor)
			prevAncestor, _ = prevAncestor.Parent().(*Container)
		}
	}

	// If the new object is a container itself, it will be visited automatically at the next actual
	// content step. However, we need to walk up the new ancestry to see if there are more new containers
	currentChildOfContainer := pointer.Resolve()

	if currentChildOfContainer == nil {
		return
	}

	currentContainerAncestor := currentChildOfContainer.Parent().(*Container)

	allChildrenEnteredAtStart := true

	check := func() bool {
		containsCurrent := false
		for _, v := range s._prevContainers {
			if v == currentContainerAncestor {
				containsCurrent = true
				break
			}
		}

		return !containsCurrent || currentContainerAncestor.CountingAtStartOnly
	}

	for currentContainerAncestor != nil && check() {

		// Check whether this ancestor container is being entered at the start,
		// by checking whether the child object is the first.
		enteringAtStart := len(currentContainerAncestor.Content()) > 0 &&
			currentChildOfContainer == currentContainerAncestor.Content()[0] &&
			allChildrenEnteredAtStart

		// Don't count it as entering at start if we're entering random somewhere within
		// a container B that happens to be nested at index 0 of container A. It only counts
		// if we're diverting directly to the first leaf node.
		if !enteringAtStart {
			allChildrenEnteredAtStart = false
		}

		// Mark a visit to this container
		s.VisitContainer(currentContainerAncestor, enteringAtStart)

		currentChildOfContainer = currentContainerAncestor
		currentContainerAncestor, _ = currentContainerAncestor.Parent().(*Container)
	}
}

func (s *Story) PopChoiceStringAndTags(tags *[]string) string {

	choiceOnlyStrVal := s.State().PopEvaluationStack().(*StringValue)

	isTag := func() bool {
		_, is := s.State().PeekEvaluationStack().(*Tag)
		return is
	}

	for len(s.State().EvaluationStack()) > 0 && isTag() {
		if tags == nil {
			*tags = []string{}
		}
		tag := s.State().PopEvaluationStack().(*Tag)
		*tags = append([]string{tag.Text()}, *tags...)
	}

	return choiceOnlyStrVal.Value()
}

func (s *Story) ProcessChoice(choicePoint *ChoicePoint) *Choice {

	showChoice := true

	// Don't create choice if choice point doesn't pass conditional
	if choicePoint.HasCondition {
		conditionValue := s.State().PopEvaluationStack()
		if !s.IsTruthy(conditionValue) {
			showChoice = false
		}
	}

	startText := ""
	choiceOnlyText := ""
	tags := []string{}

	if choicePoint.HasChoiceOnlyContent {
		choiceOnlyText = s.PopChoiceStringAndTags(&tags)
	}

	if choicePoint.HasStartContent {
		startText = s.PopChoiceStringAndTags(&tags)
	}

	// Don't create choice if player has already read this content
	if choicePoint.OnceOnly {
		visitCount := s.State().VisitCountForContainer(choicePoint.ChoiceTarget())
		if visitCount > 0 {
			showChoice = false
		}
	}

	// We go through the full process of creating the choice above so
	// that we consume the content for it, since otherwise it'll
	// be shown on the output stream.
	if !showChoice {
		return nil
	}

	choice := NewChoice()
	choice.TargetPath = choicePoint.PathOnChoice()
	choice.SourcePath = choicePoint.Path(choicePoint).String()
	choice.IsInvisibleDefault = choicePoint.IsInvisibleDefault
	choice.Tags = tags

	// We need to capture the state of the callstack at the point where
	// the choice was generated, since after the generation of this choice
	// we may go on to pop out from a tunnel (possible if the choice was
	// wrapped in a conditional), or we may pop out from a thread,
	// at which point that thread is discarded.
	// Fork clones the thread, gives it a new ID, but without affecting
	// the thread stack itself.
	choice.ThreadAtGeneration = s.State().CallStack().ForkThread()

	// Set final text for the choice
	choice.Text = strings.Trim(startText+choiceOnlyText, " \t")

	return choice
}

// IsTruthy
// Does the expression result represented by this object evaluate to true?
// e.g. is it a Number that's not equal to 1?
func (s *Story) IsTruthy(obj Object) bool {

	truthy := false
	if value, isValue := obj.(Value); isValue {
		val := value
		if divTarget, isDivertTargetValue := val.(*DivertTargetValue); isDivertTargetValue {
			s.Error("Shouldn't use a divert target (to " + divTarget.TargetPath().String() + ") as a conditional value. Did you intend a function call 'likeThis()' or a read count check 'likeThis'? (no arrows)")
			return false
		}

		fmt.Println(val)

		return val.IsTruthy()
	}
	return truthy
}

// PerformLogicAndFlowControl
// Checks whether contentObj is a control or flow object rather than a piece of content,
// and performs the required command if necessary.
func (s *Story) PerformLogicAndFlowControl(contentObj Object) bool {

	if contentObj == nil {
		return false
	}

	// Divert
	if currentDivert, isDivert := contentObj.(*Divert); isDivert {

		if currentDivert.IsConditional {
			conditionValue := s.State().PopEvaluationStack()

			// False conditional? Cancel divert
			if !s.IsTruthy(conditionValue) {
				return true
			}
		}

		if currentDivert.HasVariableTarget() {

			varName := currentDivert.VariableDivertName

			varContents := s.State().VariablesState().GetVariableWithName(varName, -1)

			if varContents == nil {
				s.Error("Tried to divert using a target from a variable that could not be found (" + varName + ")")
			} else if _, isDivertTargetValue := varContents.(*DivertTargetValue); !isDivertTargetValue {

				intContent, _ := varContents.(*IntValue)

				errorMessage := "Tried to divert to a target from a variable, but the variable (" + varName + ") didn't contain a divert target, it "
				if intContent != nil && intContent.Value() == 0 {
					errorMessage += "was empty/null (the value 0)."
				} else {
					errorMessage += "contained '" + fmt.Sprint(varContents) + "'."
				}

				s.Error(errorMessage)
			}

			target := varContents.(*DivertTargetValue)
			s.State().DivertedPointer = s.PointerAtPath(target.TargetPath())
		} else if currentDivert.IsExternal {
			s.CallExternalFunction(currentDivert.TargetPathString(), currentDivert.ExternalArgs)
			return true
		} else {
			s.State().DivertedPointer = currentDivert.TargetPointer()
		}

		if currentDivert.PushesToStack {
			s.State().CallStack().Push(currentDivert.StackPushType, 0, len(s.State().OutputStream()))
		}

		if s.State().DivertedPointer.IsNull() && !currentDivert.IsExternal {

			// Human readable name available - runtime divert is part of a hard-written divert that to missing content
			if currentDivert != nil && currentDivert.DebugMetadata().SourceName != "" {
				s.Error("Divert target doesn't exist: " + currentDivert.DebugMetadata().SourceName)
			} else {
				s.Error("Divert resolution failed: " + currentDivert.String())
			}
		}

		return true
	}

	// Control Command
	if evalCommand, isControlCommand := contentObj.(*ControlCommand); isControlCommand {
		// Start/end an expression evaluation? Or print out the result?

		switch evalCommand.CommandType {

		case CommandTypeEvalStart:

			// Assert (state.inExpressionEvaluation == false, "Already in expression evaluation?");
			s.State().SetInExpressionEvaluation(true)

		case CommandTypeEvalEnd:

			//Assert (state.inExpressionEvaluation == true, "Not in expression evaluation mode");
			s.State().SetInExpressionEvaluation(false)

		case CommandTypeEvalOutput:

			// If the expression turned out to be empty, there may not be anything on the stack
			if len(s.State().EvaluationStack()) > 0 {

				output := s.State().PopEvaluationStack()

				// Functions may evaluate to Void, in which case we skip output
				if _, isVoid := output.(*Void); !isVoid {

					// TODO: Should we really always blanket convert to string?
					// It would be okay to have numbers in the output stream the
					// only problem is when exporting text for viewing, it skips over numbers etc.

					text := NewStringValueFromString(output.(fmt.Stringer).String())

					s.State().PushToOutputStream(text)
				}
			}

		case CommandTypeNoOp:

		case CommandTypeDuplicate:

			s.State().PushEvaluationStack(s.State().PeekEvaluationStack())

		case CommandTypePopEvaluatedValue:

			s.State().PopEvaluationStack()

		case CommandTypePopFunction:
			fallthrough

		case CommandTypePopTunnel:

			var popType PushPopType
			if evalCommand.CommandType == CommandTypePopFunction {
				popType = Function
			} else {
				popType = Tunnel
			}

			// Tunnel onwards is allowed to specify an optional override
			// divert to go to immediately after returning: ->-> target
			var overrideTunnelReturnTarget *DivertTargetValue
			if popType == Tunnel {
				popped := s.State().PopEvaluationStack()
				overrideTunnelReturnTarget, _ = popped.(*DivertTargetValue)
				if overrideTunnelReturnTarget == nil {
					panic("Expected void if ->-> doesn't override target")
					//Assert (popped is Void, "Expected void if ->-> doesn't override target");
				}
			}

			if s.State().TryExitFunctionEvaluationFromGame() {
				break
			}

			if s.State().CallStack().CurrentElement().PushPopType() != popType || !s.State().CallStack().CanPop() {

				names := make(map[PushPopType]string, 0)
				names[Function] = "function return statement (~ return)"
				names[Tunnel] = "tunnel onwards statement (->->)"

				expected := names[s.State().CallStack().CurrentElement().PushPopType()]
				if !s.State().CallStack().CanPop() {
					expected = "end of flow (-> END or choice)"
				}

				var errorMsg = fmt.Sprintf("Found %s, when expected %s", names[popType], expected)

				s.Error(errorMsg)
			} else {

				s.State().PopCallstack(-1)

				// Does tunnel onwards override by diverting to a new ->-> target?
				if overrideTunnelReturnTarget != nil {
					s.State().DivertedPointer = s.PointerAtPath(overrideTunnelReturnTarget.TargetPath())
				}
			}

		case CommandTypeBeginString:

			s.State().PushToOutputStream(evalCommand)

			if s.State().InExpressionEvaluation() == false {
				panic("Expected to be in an expression when evaluating a string")
			}

			//Assert (state.inExpressionEvaluation == true, "Expected to be in an expression when evaluating a string");
			s.State().SetInExpressionEvaluation(false)

		case CommandTypeBeginTag:

			s.State().PushToOutputStream(evalCommand)

		case CommandTypeEndTag:

			// EndTag has 2 modes:
			//  - When in string evaluation (for choices)
			//  - Normal
			//
			// The only way you could have an EndTag in the middle of
			// string evaluation is if we're currently generating text for a
			// choice, such as:
			//
			//   + choice # tag
			//
			// In the above case, the ink will be run twice:
			//  - First, to generate the choice text. String evaluation
			//    will be on, and the final string will be pushed to the
			//    evaluation stack, ready to be popped to make a Choice
			//    object.
			//  - Second, when ink generates text after choosing the choice.
			//    On this ocassion, it's not in string evaluation mode.
			//
			// On the writing side, we disallow manually putting tags within
			// strings like this:
			//
			//   {"hello # world"}
			//
			// So we know that the tag must be being generated as part of
			// choice content. Therefore, when the tag has been generated,
			// we push it onto the evaluation stack in the exact same way
			// as the string for the choice content.
			if s.State().InStringEvaluation() {

				contentStackForTag := NewStack[Object]()
				outputCountConsumed := 0

				for i := len(s.State().OutputStream()) - 1; i >= 0; i-- {

					obj := s.State().OutputStream()[i]

					outputCountConsumed++
					if command, ok := obj.(*ControlCommand); ok {
						if command.CommandType == CommandTypeBeginTag {
							break
						}

						s.Error("Unexpected ControlCommand while extracting tag from choice")
						break
					}

					if _, isStringValue := obj.(*StringValue); isStringValue {
						contentStackForTag.Push(obj)
					}
				}

				// Consume the content that was produced for this string
				s.State().PopFromOutputStream(outputCountConsumed)

				var sb strings.Builder

				for _, val := range contentStackForTag.items {
					strVal := val.(*StringValue)
					sb.WriteString(strVal.Value())
				}

				choiceTag := NewTag(s.State().CleanOutputWhitespace(sb.String()))

				// Pushing to the evaluation stack means it gets picked up
				// when a Choice is generated from the next Choice Point.
				s.State().PushEvaluationStack(choiceTag)
			} else {
				// Otherwise! Simply push EndTag, so that in the output stream we
				// have a structure of: [BeginTag, "the tag content", EndTag]

				s.State().PushToOutputStream(evalCommand)
			}

		case CommandTypeEndString:
			// Dynamic strings and tags are built in the same way

			// Since we're iterating backward through the content,
			// build a stack so that when we build the string,
			// it's in the right order
			contentStackForString := NewStack[Object]()
			contentToRetain := NewStack[Object]()

			outputCountConsumed := 0
			for i := len(s.State().OutputStream()) - 1; i >= 0; i-- {

				obj := s.State().OutputStream()[i]

				outputCountConsumed++

				if command, ok := obj.(*ControlCommand); ok && command.CommandType == CommandTypeBeginString {
					break
				}

				if _, isTag := obj.(*Tag); isTag {
					contentToRetain.Push(obj)
				}

				if _, isStringValue := obj.(*StringValue); isStringValue {
					contentStackForString.Push(obj)
				}
			}

			// Consume the content that was produced for this string
			s.State().PopFromOutputStream(outputCountConsumed)

			// Rescue the tags that we want actually to keep on the output stack
			// rather than consume as part of the string we're building.
			// At the time of writing, this only applies to Tag objects generated
			// by choices, which are pushed to the stack during string generation.
			for _, rescuedTag := range contentToRetain.items {
				s.State().PushToOutputStream(rescuedTag)
			}

			// Build string out of the content we collected
			var sb strings.Builder
			for _, c := range contentStackForString.items {
				sb.WriteString(c.(fmt.Stringer).String())
			}

			// Return to expression evaluation (from content mode)
			s.State().SetInExpressionEvaluation(true)
			s.State().PushEvaluationStack(NewStringValueFromString(sb.String()))

		case CommandTypeChoiceCount:

			choiceCount := len(s.State().GeneratedChoices())
			s.State().PushEvaluationStack(NewIntValueFromInt(choiceCount))

		case CommandTypeTurns:

			s.State().PushEvaluationStack(NewIntValueFromInt(s.State().CurrentTurnIndex() + 1))

		case CommandTypeTurnsSince:
			fallthrough
		case CommandTypeReadCount:

			target := s.State().PopEvaluationStack()
			if _, isDivertTargetValue := target.(*DivertTargetValue); !isDivertTargetValue {
				extraNote := ""
				if _, isIntValue := target.(*IntValue); isIntValue {
					extraNote = ". Did you accidentally pass a read count ('knot_name') instead of a target ('-> knot_name')?"
				}
				s.Error("TURNS_SINCE expected a divert target (knot, stitch, label name), but saw " + target.(fmt.Stringer).String() + extraNote)
				break
			}

			divertTarget, _ := target.(*DivertTargetValue)
			container := s.ContentAtPath(divertTarget.TargetPath()).CorrectObj().(*Container)

			eitherCount := 0
			if container != nil {
				if evalCommand.CommandType == CommandTypeTurnsSince {
					eitherCount = s.State().TurnsSinceForContainer(container)
				} else {
					eitherCount = s.State().VisitCountForContainer(container)
				}
			} else {
				if evalCommand.CommandType == CommandTypeTurnsSince {
					eitherCount = -1 // turn count, default to never/unknown
				} else {
					eitherCount = 0 // visit count, assume 0 to default to allowing entry
				}
				s.Warning("Failed to find container for " + evalCommand.String() + " lookup at " + divertTarget.TargetPath().String())
			}

			s.State().PushEvaluationStack(NewIntValueFromInt(eitherCount))

		case CommandTypeRandom:

			maxInt, _ := s.State().PopEvaluationStack().(*IntValue)
			minInt, _ := s.State().PopEvaluationStack().(*IntValue)

			if minInt == nil {
				s.Error("Invalid value for minimum parameter of RANDOM(min, max)")
			}

			if maxInt == nil {
				s.Error("Invalid value for maximum parameter of RANDOM(min, max)")
			}

			// +1 because it's inclusive of min and max, for e.g. RANDOM(1,6) for a dice roll.
			randomRange := maxInt.Value() - minInt.Value() + 1
			/*
				   int randomRange;
					try {
						randomRange = checked(maxInt.value - minInt.value + 1);
					} catch (System.OverflowException) {
						randomRange = int.MaxValue;
						Error("RANDOM was called with a range that exceeds the size that ink numbers can use.");
					}
			*/
			if randomRange <= 0 {
				s.Error("RANDOM was called with minimum as " + fmt.Sprint(minInt.Value()) + " and maximum as " + fmt.Sprint(maxInt.Value()) + ". The maximum must be larger")
			}

			resultSeed := s.State().StorySeed + s.State().PreviousRandom
			random := rand.New(rand.NewSource(int64(resultSeed)))

			nextRandom := random.Int()
			chosenValue := nextRandom%randomRange + minInt.Value()
			s.State().PushEvaluationStack(NewIntValueFromInt(chosenValue))

			// Next random number (rather than keeping the Random object around)
			s.State().PreviousRandom = nextRandom

		case CommandTypeSeedRandom:

			seed := s.State().PopEvaluationStack().(*IntValue)
			if seed == nil {
				s.Error("Invalid value passed to SEED_RANDOM")
			}

			// Story seed affects both RANDOM and shuffle behaviour
			s.State().StorySeed = seed.Value()
			s.State().PreviousRandom = 0

			// SEED_RANDOM returns nothing.
			s.State().PushEvaluationStack(NewVoid())

		case CommandTypeVisitIndex:

			count := s.State().VisitCountForContainer(s.State().CurrentPointer().Container) - 1 // index not count
			s.State().PushEvaluationStack(NewIntValueFromInt(count))

		case CommandTypeSequenceShuffleIndex:

			shuffleIndex := s.NextSequenceShuffleIndex()
			s.State().PushEvaluationStack(NewIntValueFromInt(shuffleIndex))

		case CommandTypeStartThread:
		// Handled in main step function

		case CommandTypeDone:

			// We may exist in the context of the initial
			// act of creating the thread, or in the context of
			// evaluating the content.
			if s.State().CallStack().CanPopThread() {
				s.State().CallStack().PopThread()
			} else {
				// In normal flow - allow safe exit without warning
				s.State().DidSafeExit = true

				// Stop flow in current thread
				s.State().SetCurrentPointer(NullPointer)
			}

		case CommandTypeEnd:

			// Force flow to end completely
			s.State().ForceEnd()

		case CommandTypeListFromInt:

			intVal, _ := s.State().PopEvaluationStack().(*IntValue)
			listNameVal, _ := s.State().PopEvaluationStack().(*StringValue)

			if intVal == nil {
				panic("Passed non-integer when creating a list element from a numerical value.")
			}

			var generatedListValue *ListValue

			if foundListDef, ok := s.ListDefinitions().TryListGetDefinition(listNameVal.Value()); ok {
				if foundItem, ok := foundListDef.TryGetItemWithValue(intVal.Value()); ok {
					generatedListValue = NewListValueFromInkListItem(foundItem, intVal.Value())
				}
			} else {
				panic("Failed to find LIST called " + listNameVal.Value())
			}

			if generatedListValue == nil {
				generatedListValue = NewListValue()
			}

			s.State().PushEvaluationStack(generatedListValue)

		case CommandTypeListRange:

			min, _ := s.State().PopEvaluationStack().(Value)
			max, _ := s.State().PopEvaluationStack().(Value)

			targetList, _ := s.State().PopEvaluationStack().(*ListValue)

			if targetList == nil || min == nil || max == nil {
				panic("Expected list, minimum and maximum for LIST_RANGE")
			}

			result := targetList.Value().ListWithSubRange(min.ValueObject(), max.ValueObject())

			s.State().PushEvaluationStack(NewListValueFromList(result))

		case CommandTypeListRandom:

			listVal, _ := s.State().PopEvaluationStack().(*ListValue)
			if listVal == nil {
				panic("Expected list for LIST_RANDOM")
			}

			list := listVal.Value()
			var newList *InkList

			// List was empty: return empty list
			if list.Count() == 0 {
				newList = NewInkList()
			} else {
				// Non-empty source list

				// Generate a random index for the element to take
				resultSeed := s.State().StorySeed + s.State().PreviousRandom
				random := rand.New(rand.NewSource(int64(resultSeed)))

				nextRandom := random.Int()
				listItemIndex := nextRandom % list.Count()

				// Iterate through to get the random element
				randomItem := KeyValuePair[InkListItem, int]{}
				i := 0
				for key, value := range list._items {
					if listItemIndex == i {
						randomItem.Key = key
						randomItem.Value = value
						break
					}
					i++
				}

				// Origin list is simply the origin of the one element
				newList = NewInkListFromOriginStory(randomItem.Key.OriginName(), s)
				newList.Set(randomItem.Key, randomItem.Value)

				s.State().PreviousRandom = nextRandom
			}

			s.State().PushEvaluationStack(NewListValueFromList(newList))

		default:
			s.Error("unhandled ControlCommand: " + fmt.Sprint(evalCommand))
		}

		return true
	}

	// Variable Assignment
	if varAss, isVariableAssignment := contentObj.(*VariableAssignment); isVariableAssignment {

		assignedVal := s.State().PopEvaluationStack()

		// When in temporary evaluation, don't create new variables purely within
		// the temporary context, but attempt to create them globally
		//var prioritiseHigherInCallStack = _temporaryEvaluationContainer != null;

		s.State().VariablesState().Assign(varAss, assignedVal)

		return true
	}

	// Variable reference
	if varRef, isVariableReference := contentObj.(*VariableReference); isVariableReference {

		var foundValue Object

		// Explicit read count value
		if varRef.PathForCount != nil {

			container := varRef.ContainerForCount()
			count := s.State().VisitCountForContainer(container)
			foundValue = NewIntValueFromInt(count)
		} else {

			foundValue = s.State().VariablesState().GetVariableWithName(varRef.Name, -1)

			if foundValue == nil {
				s.Warning("Variable not found: '" + varRef.Name + "'. Using default value of 0 (false). This can happen with temporary variables if the declaration hasn't yet been hit. Globals are always given a default value on load if a value doesn't exist in the save state.")
				foundValue = NewIntValueFromInt(0)
			}
		}

		s.State().PushEvaluationStack(foundValue)

		return true
	}

	// Native function call
	if nfunc, isNativeFunctionCall := contentObj.(*NativeFunctionCall); isNativeFunctionCall {

		funcParams := s.State().PopEvaluationStackEx(nfunc.NumberOfParameters())
		result := nfunc.Call(funcParams)
		s.State().PushEvaluationStack(result)
		return true
	}

	// No control content, must be ordinary content
	return false
}

type ExternalFunctionDef struct {
	function      func(args []interface{}) interface{}
	lookaheadSafe bool
}

// ChoosePathString
// Change the current position of the story to the given path. From here you can
// call Continue() to evaluate the next line.
//
// The path string is a dot-separated path as used internally by the engine.
// These examples should work:
//
//    myKnot
//    myKnot.myStitch
//
// Note however that this won't necessarily work:
//
//    myKnot.myStitch.myLabelledChoice
//
// ...because of the way that content is nested within a weave structure.
//
// By default this will reset the callstack beforehand, which means that any
// tunnels, threads or functions you were in at the time of calling will be
// discarded. This is different from the behaviour of ChooseChoiceIndex, which
// will always keep the callstack, since the choices are known to come from the
// correct state, and known their source thread.
//
// You have the option of passing false to the resetCallstack parameter if you
// don't want this behaviour, and will leave any active threads, tunnels or
// function calls in-tact.
//
// This is potentially dangerous! If you're in the middle of a tunnel,
// it'll redirect only the inner-most tunnel, meaning that when you tunnel-return
// using '->->', it'll return to where you were before. This may be what you
// want though. However, if you're in the middle of a function, ChoosePathString
// will throw an exception.
//
// (default) resetCallstack: true
func (s *Story) ChoosePathString(path string, resetCallstack bool, arguments ...interface{}) {

	s.IfAsyncWeCant("call ChoosePathString right now")

	//if(onChoosePathString != null) onChoosePathString(path, arguments);
	if s.OnChoosePathString != nil {
		s.OnChoosePathString.Emit(path, arguments)
	}

	if resetCallstack {
		s.ResetCallstack()
	} else {
		// ChoosePathString is potentially dangerous since you can call it when the stack is
		// pretty much in any state. Let's catch one of the worst offenders.
		if s.State().CallStack().CurrentElement().PushPopType() == Function {
			funcDetail := ""
			container := s.State().CallStack().CurrentElement().CurrentPointer.Container
			if container != nil {
				funcDetail = "(" + container.Path(container).String() + ") "
			}
			panic("Story was running a function " + funcDetail + "when you called ChoosePathString(" + path + ") - this is almost certainly not not what you want! Full stack trace: \n" + s.State().CallStack().CallStackTrace())
		}
	}

	s.State().PassArgumentsToEvaluationStack(arguments)
	s.ChoosePath(NewPathFromString(path), true)
}

func (s *Story) IfAsyncWeCant(activityStr string) {

	if s._asyncContinueActive {
		panic("Can't " + activityStr + ". Story is in the middle of a ContinueAsync(). Make more ContinueAsync() calls or a single Continue() call beforehand.")
	}
}

// ChoosePath
// (default) incrementingTurnIndex: true
func (s *Story) ChoosePath(p *Path, incrementingTurnIndex bool) {

	s.State().SetChosenPath(p, incrementingTurnIndex)

	// Take a note of newly visited containers for read counts etc
	s.VisitChangedContainersDueToDivert()
}

// ChooseChoiceIndex
// Chooses the Choice from the currentChoices list with the given
// index. Internally, this sets the current content path to that
// pointed to by the Choice, ready to continue story evaluation.
func (s *Story) ChooseChoiceIndex(choiceIdx int) {

	choices := s.CurrentChoices()
	//Assert (choiceIdx >= 0 && choiceIdx < choices.Count, "choice out of range");

	// Replace callstack with the one from the thread at the choosing point,
	// so that we can jump into the right place in the flow.
	// This is important in case the flow was forked by a new thread, which
	// can create multiple leading edges for the story, each of
	// which has its own context.
	choiceToChoose := choices[choiceIdx]
	//if(onMakeChoice != null) onMakeChoice(choiceToChoose);
	if s.OnMakeChoice != nil {
		s.OnMakeChoice.Emit(choiceToChoose)
	}
	s.State().CallStack().SetCurrentThread(choiceToChoose.ThreadAtGeneration)

	s.ChoosePath(choiceToChoose.TargetPath, true)
}

// HasFunction
// Checks if a function exists.
func (s *Story) HasFunction(functionName string) bool {

	/*
	   try {
	       return KnotContainerWithName (functionName) != null;
	   } catch {
	       return false;
	   }
	*/

	return s.KnotContainerWithName(functionName) != nil
}

// EvaluateFunction
// Evaluates a function defined in ink.
/*
   public object EvaluateFunction (string functionName, params object [] arguments)
   {
       string _;
       return EvaluateFunction (functionName, out _, arguments);
   }
*/

// EvaluateFunction
// Evaluates a function defined in ink, and gathers the possibly multi-line text as generated by the function.
// This text output is any text written as normal content within the function, as opposed to the return value, as returned with `~ return`.
func (s *Story) EvaluateFunction(functionName string, arguments ...interface{}) (string, interface{}) {

	textOutput := ""

	//if(onEvaluateFunction != null) onEvaluateFunction(functionName, arguments);
	if s.OnEvaluateFunction != nil {
		s.OnEvaluateFunction.Emit(functionName, arguments)
	}

	s.IfAsyncWeCant("evaluate a function")

	//if(functionName == null) {
	//	throw new System.Exception ("Function is null");
	//} else if(functionName == string.Empty || functionName.Trim() == string.Empty) {
	//	throw new System.Exception ("Function is empty or white space.");
	//}

	if strings.TrimSpace(functionName) == "" {
		panic("Function is empty or white space.")
	}

	// Get the content that we need to run
	funcContainer := s.KnotContainerWithName(functionName)
	if funcContainer == nil {
		panic("Function doesn't exist: '" + functionName + "'")
	}

	// Snapshot the output stream
	outputStreamBefore := NewSliceFromSlice(s._state.OutputStream())
	s._state.ResetOutput(nil)

	// State will temporarily replace the callstack in order to evaluate
	s.State().StartFunctionEvaluationFromGame(funcContainer, arguments)

	// Evaluate the function, and collect the string output
	var stringOutput strings.Builder
	for s.CanContinue() {
		stringOutput.WriteString(s.Continue())
	}
	textOutput = stringOutput.String()

	// Restore the output stream in case this was called
	// during main story evaluation.
	s._state.ResetOutput(outputStreamBefore)

	result := s.State().CompleteFunctionEvaluationFromGame()
	//if(onCompleteEvaluateFunction != null) onCompleteEvaluateFunction(functionName, arguments, textOutput, result);
	if s.OnCompleteEvaluateFunction != nil {
		s.OnCompleteEvaluateFunction.Emit(functionName, arguments, textOutput, result)
	}

	return textOutput, result
}

func (s *Story) EvaluateExpression(exprContainer *Container) Object {

	startCallStackHeight := len(s.State().CallStack().Elements())

	s.State().CallStack().Push(Tunnel, 0, 0)

	s._temporaryEvaluationContainer = exprContainer

	s.State().GoToStart()

	evalStackHeight := len(s.State().EvaluationStack())

	s.Continue()

	s._temporaryEvaluationContainer = nil

	// Should have fallen off the end of the Container, which should
	// have auto-popped, but just in case we didn't for some reason,
	// manually pop to restore the state (including currentPath).
	if len(s.State().CallStack().Elements()) > startCallStackHeight {
		s.State().PopCallstack(-1)
	}

	endStackHeight := len(s.State().EvaluationStack())
	if endStackHeight > evalStackHeight {
		return s.State().PopEvaluationStack()
	}

	return nil
}

func (s *Story) CallExternalFunction(funcName string, numberOfArguments int) {

	var fallbackFunctionContainer *Container
	funcDef, foundExternal := s._externals[funcName]

	// Should this function break glue? Abort run if we've already seen a newline.
	// Set a bool to tell it to restore the snapshot at the end of this instruction.
	if foundExternal && !funcDef.lookaheadSafe && s._stateSnapshotAtLastNewline != nil {
		s._sawLookaheadUnsafeFunctionAfterNewline = true
		return
	}

	// Try to use fallback function?
	if !foundExternal {
		if s.AllowExternalFunctionFallbacks {
			fallbackFunctionContainer = s.KnotContainerWithName(funcName)
			//Assert (fallbackFunctionContainer != null, "Trying to call EXTERNAL function '" + funcName + "' which has not been bound, and fallback ink function could not be found.");

			s.State().CallStack().Push(
				Function,
				0,
				len(s.State().OutputStream()))
			s.State().DivertedPointer = StartOfPointer(fallbackFunctionContainer)
			return
		} else {
			//Assert (false, "Trying to call EXTERNAL function '" + funcName + "' which has not been bound (and ink fallbacks disabled).");
		}
	}

	// Pop arguments
	var arguments []interface{}
	for i := 0; i < numberOfArguments; i++ {
		poppedObj, _ := s.State().PopEvaluationStack().(Value)
		valueObj := poppedObj.ValueObject()
		arguments = append(arguments, valueObj)
	}

	// Reverse arguments from the order they were popped,
	// so they're the right way round again.
	//arguments.Reverse ();
	var argumentsReordered []interface{}
	for i := len(argumentsReordered); i >= 0; i-- {
		argumentsReordered = append(argumentsReordered, arguments[i])
	}

	// Run the function!
	funcResult := funcDef.function(argumentsReordered)

	// Convert return value (if any) to the a type that the ink engine can use
	var returnObj Object
	if funcResult != nil {
		returnObj = CreateValue(funcResult)
		//Assert (returnObj != null, "Could not create ink value from returned object of type " + funcResult.GetType());
	} else {
		returnObj = NewVoid()
	}

	s.State().PushEvaluationStack(returnObj)
}

// BindExternalFunctionalGeneral
// Most general form of function binding that returns an object
// and takes an array of object parameters.
// The only way to bind a function with more than 3 arguments.
// </summary>
// <param name="funcName">EXTERNAL ink function name to bind to.</param>
// <param name="func">The C# function to bind.</param>
// <param name="lookaheadSafe">The ink engine often evaluates further
// than you might expect beyond the current line just in case it sees
// glue that will cause the two lines to become one. In this case it's
// possible that a function can appear to be called twice instead of
// just once, and earlier than you expect. If it's safe for your
// function to be called in this way (since the result and side effect
// of the function will not change), then you can pass 'true'.
// Usually, you want to pass 'false', especially if you want some action
// (default) lookaheadSafe: true
func (s *Story) BindExternalFunctionalGeneral(funcName string, gfunc func(args []interface{}) interface{}, lookaheadSafe bool) {

	s.IfAsyncWeCant("bind an external function")
	//Assert(!_externals.ContainsKey(funcName), "Function '"+funcName+"' has already been bound.")
	s._externals[funcName] = &ExternalFunctionDef{
		function:      gfunc,
		lookaheadSafe: lookaheadSafe,
	}
}

func TryCoerce[T any](value interface{}) interface{} {

	if value == nil {
		return nil
	}

	if _, isT := value.(T); isT {
		return value
	}

	var t T

	if v, isFloat64 := value.(float64); isFloat64 && reflect.TypeOf(t) == reflect.TypeOf(int(0)) {
		intVal := int(math.Round(v))
		return intVal
	}

	if v, isInt := value.(int); isInt && reflect.TypeOf(t) == reflect.TypeOf(float64(0)) {
		floatVal := float64(v)
		return floatVal
	}

	if v, isInt := value.(int); isInt && reflect.TypeOf(t) == reflect.TypeOf(bool(true)) {
		if v == 0 {
			return false
		}
		return true
	}

	if v, isBool := value.(bool); isBool && reflect.TypeOf(t) == reflect.TypeOf(int(0)) {
		if v == true {
			return 1
		} else {
			return 0
		}
	}

	if reflect.TypeOf(t) == reflect.TypeOf(string("")) {
		if v, isString := value.(string); isString {
			return v
		} else {
			return value.(fmt.Stringer).String()
		}
	}

	//Assert (false, "Failed to cast " + value.GetType ().Name + " to " + typeof(T).Name);
	panic("Failed to cast " + reflect.TypeOf(value).Name() + " to " + reflect.TypeOf(t).Name())

	return nil
}

// Convenience overloads for standard functions and actions of various arities
// Is there a better way of doing this?!

/*
        /// <summary>
        /// Bind a C# Action to an ink EXTERNAL function declaration.
        /// </summary>
        /// <param name="funcName">EXTERNAL ink function name to bind to.</param>
        /// <param name="act">The C# action to bind.</param>
        /// <param name="lookaheadSafe">The ink engine often evaluates further
        /// than you might expect beyond the current line just in case it sees
        /// glue that will cause the two lines to become one. In this case it's
        /// possible that a function can appear to be called twice instead of
        /// just once, and earlier than you expect. If it's safe for your
        /// function to be called in this way (since the result and side effect
        /// of the function will not change), then you can pass 'true'.
        /// Usually, you want to pass 'false', especially if you want some action
        /// to be performed in game code when this function is called.</param>
        public void BindExternalFunction(string funcName, Action act, bool lookaheadSafe=false)
        {
			Assert(act != null, "Can't bind a null function");

            BindExternalFunctionGeneral (funcName, (object[] args) => {
                Assert(args.Length == 0, "External function expected no arguments");
                act();
                return null;
            }, lookaheadSafe);
        }
*/

/*
  /// <summary>
        /// Bind a C# function to an ink EXTERNAL function declaration.
        /// </summary>
        /// <param name="funcName">EXTERNAL ink function name to bind to.</param>
        /// <param name="func">The C# function to bind.</param>
        /// <param name="lookaheadSafe">The ink engine often evaluates further
        /// than you might expect beyond the current line just in case it sees
        /// glue that will cause the two lines to become one. In this case it's
        /// possible that a function can appear to be called twice instead of
        /// just once, and earlier than you expect. If it's safe for your
        /// function to be called in this way (since the result and side effect
        /// of the function will not change), then you can pass 'true'.
        /// Usually, you want to pass 'false', especially if you want some action
        /// to be performed in game code when this function is called.</param>
        public void BindExternalFunction<T>(string funcName, Func<T, object> func, bool lookaheadSafe=false)
        {
			Assert(func != null, "Can't bind a null function");

            BindExternalFunctionGeneral (funcName, (object[] args) => {
                Assert(args.Length == 1, "External function expected one argument");
                return func( (T)TryCoerce<T>(args[0]) );
            }, lookaheadSafe);
        }
*/

/*
  /// <summary>
        /// Bind a C# action to an ink EXTERNAL function declaration.
        /// </summary>
        /// <param name="funcName">EXTERNAL ink function name to bind to.</param>
        /// <param name="act">The C# action to bind.</param>
        /// <param name="lookaheadSafe">The ink engine often evaluates further
        /// than you might expect beyond the current line just in case it sees
        /// glue that will cause the two lines to become one. In this case it's
        /// possible that a function can appear to be called twice instead of
        /// just once, and earlier than you expect. If it's safe for your
        /// function to be called in this way (since the result and side effect
        /// of the function will not change), then you can pass 'true'.
        /// Usually, you want to pass 'false', especially if you want some action
        /// to be performed in game code when this function is called.</param>
        public void BindExternalFunction<T>(string funcName, Action<T> act, bool lookaheadSafe=false)
        {
			Assert(act != null, "Can't bind a null function");

            BindExternalFunctionGeneral (funcName, (object[] args) => {
                Assert(args.Length == 1, "External function expected one argument");
                act( (T)TryCoerce<T>(args[0]) );
                return null;
            }, lookaheadSafe);
        }
*/

/*
 /// <summary>
        /// Bind a C# function to an ink EXTERNAL function declaration.
        /// </summary>
        /// <param name="funcName">EXTERNAL ink function name to bind to.</param>
        /// <param name="func">The C# function to bind.</param>
        /// <param name="lookaheadSafe">The ink engine often evaluates further
        /// than you might expect beyond the current line just in case it sees
        /// glue that will cause the two lines to become one. In this case it's
        /// possible that a function can appear to be called twice instead of
        /// just once, and earlier than you expect. If it's safe for your
        /// function to be called in this way (since the result and side effect
        /// of the function will not change), then you can pass 'true'.
        /// Usually, you want to pass 'false', especially if you want some action
        /// to be performed in game code when this function is called.</param>
        public void BindExternalFunction<T1, T2>(string funcName, Func<T1, T2, object> func, bool lookaheadSafe = false)
        {
			Assert(func != null, "Can't bind a null function");

            BindExternalFunctionGeneral (funcName, (object[] args) => {
                Assert(args.Length == 2, "External function expected two arguments");
                return func(
                    (T1)TryCoerce<T1>(args[0]),
                    (T2)TryCoerce<T2>(args[1])
                );
            }, lookaheadSafe);
        }
*/

/*
/// <summary>
        /// Bind a C# action to an ink EXTERNAL function declaration.
        /// </summary>
        /// <param name="funcName">EXTERNAL ink function name to bind to.</param>
        /// <param name="act">The C# action to bind.</param>
        /// <param name="lookaheadSafe">The ink engine often evaluates further
        /// than you might expect beyond the current line just in case it sees
        /// glue that will cause the two lines to become one. In this case it's
        /// possible that a function can appear to be called twice instead of
        /// just once, and earlier than you expect. If it's safe for your
        /// function to be called in this way (since the result and side effect
        /// of the function will not change), then you can pass 'true'.
        /// Usually, you want to pass 'false', especially if you want some action
        /// to be performed in game code when this function is called.</param>
        public void BindExternalFunction<T1, T2>(string funcName, Action<T1, T2> act, bool lookaheadSafe=false)
        {
			Assert(act != null, "Can't bind a null function");

            BindExternalFunctionGeneral (funcName, (object[] args) => {
                Assert(args.Length == 2, "External function expected two arguments");
                act(
                    (T1)TryCoerce<T1>(args[0]),
                    (T2)TryCoerce<T2>(args[1])
                );
                return null;
            }, lookaheadSafe);
        }
*/

/*
  /// <summary>
        /// Bind a C# function to an ink EXTERNAL function declaration.
        /// </summary>
        /// <param name="funcName">EXTERNAL ink function name to bind to.</param>
        /// <param name="func">The C# function to bind.</param>
        /// <param name="lookaheadSafe">The ink engine often evaluates further
        /// than you might expect beyond the current line just in case it sees
        /// glue that will cause the two lines to become one. In this case it's
        /// possible that a function can appear to be called twice instead of
        /// just once, and earlier than you expect. If it's safe for your
        /// function to be called in this way (since the result and side effect
        /// of the function will not change), then you can pass 'true'.
        /// Usually, you want to pass 'false', especially if you want some action
        /// to be performed in game code when this function is called.</param>
        public void BindExternalFunction<T1, T2, T3>(string funcName, Func<T1, T2, T3, object> func, bool lookaheadSafe=false)
        {
			Assert(func != null, "Can't bind a null function");

            BindExternalFunctionGeneral (funcName, (object[] args) => {
                Assert(args.Length == 3, "External function expected three arguments");
                return func(
                    (T1)TryCoerce<T1>(args[0]),
                    (T2)TryCoerce<T2>(args[1]),
                    (T3)TryCoerce<T3>(args[2])
                );
            }, lookaheadSafe);
        }
*/

/*
/// <summary>
        /// Bind a C# action to an ink EXTERNAL function declaration.
        /// </summary>
        /// <param name="funcName">EXTERNAL ink function name to bind to.</param>
        /// <param name="act">The C# action to bind.</param>
        /// <param name="lookaheadSafe">The ink engine often evaluates further
        /// than you might expect beyond the current line just in case it sees
        /// glue that will cause the two lines to become one. In this case it's
        /// possible that a function can appear to be called twice instead of
        /// just once, and earlier than you expect. If it's safe for your
        /// function to be called in this way (since the result and side effect
        /// of the function will not change), then you can pass 'true'.
        /// Usually, you want to pass 'false', especially if you want some action
        /// to be performed in game code when this function is called.</param>
        public void BindExternalFunction<T1, T2, T3>(string funcName, Action<T1, T2, T3> act, bool lookaheadSafe=false)
        {
			Assert(act != null, "Can't bind a null function");

            BindExternalFunctionGeneral (funcName, (object[] args) => {
                Assert(args.Length == 3, "External function expected three arguments");
                act(
                    (T1)TryCoerce<T1>(args[0]),
                    (T2)TryCoerce<T2>(args[1]),
                    (T3)TryCoerce<T3>(args[2])
                );
                return null;
            }, lookaheadSafe);
        }
*/

/*
 /// <summary>
        /// Bind a C# function to an ink EXTERNAL function declaration.
        /// </summary>
        /// <param name="funcName">EXTERNAL ink function name to bind to.</param>
        /// <param name="func">The C# function to bind.</param>
        /// <param name="lookaheadSafe">The ink engine often evaluates further
        /// than you might expect beyond the current line just in case it sees
        /// glue that will cause the two lines to become one. In this case it's
        /// possible that a function can appear to be called twice instead of
        /// just once, and earlier than you expect. If it's safe for your
        /// function to be called in this way (since the result and side effect
        /// of the function will not change), then you can pass 'true'.
        /// Usually, you want to pass 'false', especially if you want some action
        /// to be performed in game code when this function is called.</param>
        public void BindExternalFunction<T1, T2, T3, T4>(string funcName, Func<T1, T2, T3, T4, object> func, bool lookaheadSafe=false)
        {
			Assert(func != null, "Can't bind a null function");

            BindExternalFunctionGeneral (funcName, (object[] args) => {
                Assert(args.Length == 4, "External function expected four arguments");
                return func(
                    (T1)TryCoerce<T1>(args[0]),
                    (T2)TryCoerce<T2>(args[1]),
                    (T3)TryCoerce<T3>(args[2]),
                    (T4)TryCoerce<T4>(args[3])
                );
            }, lookaheadSafe);
        }
*/

/*
   /// <summary>
        /// Bind a C# action to an ink EXTERNAL function declaration.
        /// </summary>
        /// <param name="funcName">EXTERNAL ink function name to bind to.</param>
        /// <param name="act">The C# action to bind.</param>
        /// <param name="lookaheadSafe">The ink engine often evaluates further
        /// than you might expect beyond the current line just in case it sees
        /// glue that will cause the two lines to become one. In this case it's
        /// possible that a function can appear to be called twice instead of
        /// just once, and earlier than you expect. If it's safe for your
        /// function to be called in this way (since the result and side effect
        /// of the function will not change), then you can pass 'true'.
        /// Usually, you want to pass 'false', especially if you want some action
        /// to be performed in game code when this function is called.</param>
        public void BindExternalFunction<T1, T2, T3, T4>(string funcName, Action<T1, T2, T3, T4> act, bool lookaheadSafe=false)
        {
			Assert(act != null, "Can't bind a null function");

            BindExternalFunctionGeneral (funcName, (object[] args) => {
                Assert(args.Length == 4, "External function expected four arguments");
                act(
                    (T1)TryCoerce<T1>(args[0]),
                    (T2)TryCoerce<T2>(args[1]),
                    (T3)TryCoerce<T3>(args[2]),
                    (T4)TryCoerce<T4>(args[3])
                );
                return null;
            }, lookaheadSafe);
        }
*/

/*
  /// <summary>
        /// Remove a binding for a named EXTERNAL ink function.
        /// </summary>
        public void UnbindExternalFunction(string funcName)
        {
            IfAsyncWeCant ("unbind an external a function");
            Assert (_externals.ContainsKey (funcName), "Function '" + funcName + "' has not been bound.");
            _externals.Remove (funcName);
        }
*/

// Check that all EXTERNAL ink functions have a valid bound C# function.
// Note that this is automatically called on the first call to Continue().
func (s *Story) ValidateExternalBindings() {

	missingExternals := make(map[string]struct{}, 0)

	s.ValidateExternalBindingsEx(s._mainContentContainer, missingExternals)
	s._hasValidatedExternals = true

	// No problem! Validation complete
	if len(missingExternals) == 0 {
		s._hasValidatedExternals = true
	} else {

		firstStr := ""
		if len(missingExternals) > 1 {
			firstStr = "s"
		}

		var missingExternalsArr []string
		for itemKey, _ := range missingExternals {
			missingExternalsArr = append(missingExternalsArr, itemKey)
		}

		thirdStr := ""
		if s.AllowExternalFunctionFallbacks {
			thirdStr = ", and no fallback ink function found."
		} else {
			thirdStr = " (ink fallbacks disabled)"
		}

		// Error for all missing externals
		var message = fmt.Sprintf("ERROR: Missing function binding for external%s: '%s' %s",
			firstStr,
			strings.Join(missingExternalsArr, "', '"),
			thirdStr,
		)

		s.Error(message)
	}
}

func (s *Story) ValidateExternalBindingsEx(c *Container, missingExternals map[string]struct{}) {

	for _, innerContent := range c.Content() {
		container, _ := innerContent.(*Container)
		if container == nil || !container.HasValidName() {
			s.ValidateExternalBindingsObject(innerContent, missingExternals)
		}
	}
	for _, innerValue := range c.NamedContent() {
		o, _ := innerValue.(Object)
		s.ValidateExternalBindingsObject(o, missingExternals)
	}
}

func (s *Story) ValidateExternalBindingsObject(o Object, missingExternals map[string]struct{}) {
	container, _ := o.(*Container)
	if container != nil {
		s.ValidateExternalBindingsEx(container, missingExternals)
		return
	}

	if divert, isDivert := o.(*Divert); isDivert && divert.IsExternal {
		name := divert.TargetPathString()

		if _, contains := s._externals[name]; contains {
			if s.AllowExternalFunctionFallbacks {
				_, fallbackFound := s.MainContentContainer().NamedContent()[name]
				if !fallbackFound {
					missingExternals[name] = struct{}{}
				}
			} else {
				missingExternals[name] = struct{}{}
			}
		}
	}
}

// ObserveVariable
// When the named global variable changes it's value, the observer will be
// called to notify it of the change. Note that if the value changes multiple
// times within the ink, the observer will only be called once, at the end
// of the ink's evaluation. If, during the evaluation, it changes and then
// changes back again to its original value, it will still be called.
// Note that the observer will also be fired if the value of the variable
// is changed externally to the ink, by directly setting a value in
// story.variablesState.
func (s *Story) ObserveVariable(variableName string, observer func(variableName, newValue interface{})) {

	s.IfAsyncWeCant("observe a new variable")

	if s._variableObservers == nil {
		s._variableObservers = make(map[string]VariableObserver)
	}

	if !s.State().VariablesState().GlobalVariableExistsWithName(variableName) {
		panic("Cannot observe variable '" + variableName + "' because it wasn't declared in the ink story.")
	}

	// TODO
	//
	//if (_variableObservers.ContainsKey (variableName)) {
	//	_variableObservers[variableName] += observer;
	//} else {
	//	_variableObservers[variableName] = observer;
	//}
}

/*
   /// <summary>
   /// Convenience function to allow multiple variables to be observed with the same
   /// observer delegate function. See the singular ObserveVariable for details.
   /// The observer will get one call for every variable that has changed.
   /// </summary>
   /// <param name="variableNames">The set of variables to observe.</param>
   /// <param name="observer">The delegate function to call when any of the named variables change.</param>
   public void ObserveVariables(IList<string> variableNames, VariableObserver observer)
   {
       foreach (var varName in variableNames) {
           ObserveVariable (varName, observer);
       }
   }
*/

/*
 /// <summary>
        /// Removes the variable observer, to stop getting variable change notifications.
        /// If you pass a specific variable name, it will stop observing that particular one. If you
        /// pass null (or leave it blank, since it's optional), then the observer will be removed
        /// from all variables that it's subscribed to. If you pass in a specific variable name and
        /// null for the the observer, all observers for that variable will be removed.
        /// </summary>
        /// <param name="observer">(Optional) The observer to stop observing.</param>
        /// <param name="specificVariableName">(Optional) Specific variable name to stop observing.</param>
        public void RemoveVariableObserver(VariableObserver observer = null, string specificVariableName = null)
        {
            IfAsyncWeCant ("remove a variable observer");

            if (_variableObservers == null)
                return;

            // Remove observer for this specific variable
            if (specificVariableName != null) {
                if (_variableObservers.ContainsKey (specificVariableName)) {
                    if( observer != null) {
                        _variableObservers [specificVariableName] -= observer;
                        if (_variableObservers[specificVariableName] == null) {
                            _variableObservers.Remove(specificVariableName);
                        }
                    }
                    else {
                        _variableObservers.Remove(specificVariableName);
                    }
                }
            }

            // Remove observer for all variables
            else if( observer != null) {
                var keys = new List<string>(_variableObservers.Keys);
                foreach (var varName in keys) {
                    _variableObservers[varName] -= observer;
                    if (_variableObservers[varName] == null) {
                        _variableObservers.Remove(varName);
                    }
                }
            }
        }
*/

func (s *Story) VariableStateDidChangeEvent(variableName string, newValueObj Object) {

	if s._variableObservers == nil {
		return
	}

	// TODO

	/*
	 VariableObserver observers = null;
	            if (_variableObservers.TryGetValue (variableName, out observers)) {

	                if (!(newValueObj is Value)) {
	                    throw new System.Exception ("Tried to get the value of a variable that isn't a standard type");
	                }
	                var val = newValueObj as Value;

	                observers (variableName, val.valueObject);
	            }
	*/
}

/*
 /// <summary>
        /// Get any global tags associated with the story. These are defined as
        /// hash tags defined at the very top of the story.
        /// </summary>
        public List<string> globalTags {
            get {
                return TagsAtStartOfFlowContainerWithPathString ("");
            }
        }
*/

/*
   /// <summary>
        /// Gets any tags associated with a particular knot or knot.stitch.
        /// These are defined as hash tags defined at the very top of a
        /// knot or stitch.
        /// </summary>
        /// <param name="path">The path of the knot or stitch, in the form "knot" or "knot.stitch".</param>
        public List<string> TagsForContentAtPath (string path)
        {
            return TagsAtStartOfFlowContainerWithPathString (path);
        }
*/

func (s *Story) TagsAtStartOfFlowContainerWithPathString(pathString string) []string {

	path := NewPathFromString(pathString)

	// Expected to be global story, knot or stitch
	flowContainer := s.ContentAtPath(path).Container()
	for true {
		firstContent := flowContainer.Content()[0]
		if flowC, isContainer := firstContent.(*Container); isContainer {
			flowContainer = flowC
		} else {
			break
		}
	}

	// Any initial tag objects count as the "main tags" associated with that story/knot/stitch
	inTag := false
	var tags []string

	for _, c := range flowContainer.Content() {

		command, _ := c.(*ControlCommand)
		if command != nil {
			if command.CommandType == CommandTypeBeginTag {
				inTag = true
			} else if command.CommandType == CommandTypeEndTag {
				inTag = false
			}
		} else if inTag {
			str, _ := c.(*StringValue)
			if str != nil {
				tags = append(tags, str.Value())
			} else {
				s.Error("Tag contained non-text content. Only plain text is allowed when using globalTags or TagsAtContentPath. If you want to evaluate dynamic content, you need to use story.Continue().")
			}
		} else {
			// Any other content - we're done
			// We only recognise initial text-only tags
			break
		}
	}

	return tags
}

/*
   /// <summary>
   /// Useful when debugging a (very short) story, to visualise the state of the
   /// story. Add this call as a watch and open the extended text. A left-arrow mark
   /// will denote the current point of the story.
   /// It's only recommended that this is used on very short debug stories, since
   /// it can end up generate a large quantity of text otherwise.
   /// </summary>
   public virtual string BuildStringOfHierarchy()
   {
       var sb = new StringBuilder ();

       mainContentContainer.BuildStringOfHierarchy (sb, 0, state.currentPointer.Resolve());

       return sb.ToString ();
   }
*/

/*
   string BuildStringOfContainer (Container container)
   {
   	var sb = new StringBuilder ();

       container.BuildStringOfHierarchy (sb, 0, state.currentPointer.Resolve());

   	return sb.ToString();
   }
*/

func (s *Story) NextContent() {

	// Setting previousContentObject is critical for VisitChangedContainersDueToDivert
	s.State().SetPreviousPointer(s.State().CurrentPointer())

	// Divert step?
	if !s.State().DivertedPointer.IsNull() {

		s.State().SetCurrentPointer(s.State().DivertedPointer)
		s.State().DivertedPointer = NullPointer

		// Internally uses state.previousContentObject and state.currentContentObject
		s.VisitChangedContainersDueToDivert()

		// Diverted location has valid content?
		if !s.State().CurrentPointer().IsNull() {
			return
		}

		// Otherwise, if diverted location doesn't have valid content,
		// drop down and attempt to increment.
		// This can happen if the diverted path is intentionally jumping
		// to the end of a container - e.g. a Conditional that's re-joining
	}

	successfulPointerIncrement := s.IncrementContentPointer()

	// Ran out of content? Try to auto-exit from a function,
	// or finish evaluating the content of a thread
	if !successfulPointerIncrement {

		didPop := false

		if s.State().CallStack().CanPopWith(Function) {

			// Pop from the call stack
			s.State().PopCallstack(Function)

			// This pop was due to dropping off the end of a function that didn't return anything,
			// so in this case, we make sure that the evaluator has something to chomp on if it needs it
			if s.State().InExpressionEvaluation() {
				s.State().PushEvaluationStack(NewVoid())
			}

			didPop = true
		} else if s.State().CallStack().CanPopThread() {

			s.State().CallStack().PopThread()

			didPop = true
		} else {
			s.State().TryExitFunctionEvaluationFromGame()
		}

		// Step past the point where we last called out
		if didPop && !s.State().CurrentPointer().IsNull() {
			s.NextContent()
		}
	}
}

func (s *Story) IncrementContentPointer() bool {

	successfulIncrement := true

	pointer := s.State().CallStack().CurrentElement().CurrentPointer
	pointer.Index++

	// Each time we step off the end, we fall out to the next container, all the
	// while we're in indexed rather than named content
	for pointer.Index >= len(pointer.Container.Content()) {

		successfulIncrement = false

		nextAncestor, _ := pointer.Container.Parent().(*Container)
		if nextAncestor == nil {
			break
		}

		indexInAncestor := -1
		for index, v := range nextAncestor.Content() {
			if v == pointer.Container {
				indexInAncestor = index
				break
			}
		}
		if indexInAncestor == -1 {
			break
		}

		pointer = NewPointer(nextAncestor, indexInAncestor)

		// Increment to next content in outer container
		pointer.Index++
		successfulIncrement = true
	}

	if !successfulIncrement {
		pointer = NullPointer
	}

	s.State().CallStack().CurrentElement().CurrentPointer = pointer

	return successfulIncrement
}

func (s *Story) TryFollowDefaultInvisibleChoice() bool {

	allChoices := s._state.CurrentChoices()

	// Is a default invisible choice the ONLY choice?
	var invisibleChoices []*Choice
	for _, choice := range allChoices {
		if choice.IsInvisibleDefault {
			invisibleChoices = append(invisibleChoices, choice)
		}
	}

	if len(invisibleChoices) == 0 || len(allChoices) > len(invisibleChoices) {
		return false
	}

	choice := invisibleChoices[0]

	// Invisible choice may have been generated on a different thread,
	// in which case we need to restore it before we continue
	s.State().CallStack().SetCurrentThread(choice.ThreadAtGeneration)

	// If there's a chance that this state will be rolled back to before
	// the invisible choice then make sure that the choice thread is
	// left intact, and it isn't re-entered in an old state.
	if s._stateSnapshotAtLastNewline != nil {
		s.State().CallStack().SetCurrentThread(s.State().CallStack().ForkThread())
	}

	s.ChoosePath(choice.TargetPath, false)

	return true
}

func (s *Story) NextSequenceShuffleIndex() int {

	numElementsIntVal, _ := s.State().PopEvaluationStack().(*IntValue)
	if numElementsIntVal == nil {
		s.Error("expected number of elements in sequence for shuffle index")
		return 0
	}

	seqContainer := s.State().CurrentPointer().Container

	numElements := numElementsIntVal.Value()

	seqCountVal, _ := s.State().PopEvaluationStack().(*IntValue)
	seqCount := seqCountVal.Value()
	loopIndex := seqCount / numElements
	iterationIndex := seqCount % numElements

	// Generate the same shuffle based on:
	//  - The hash of this container, to make sure it's consistent
	//    each time the runtime returns to the sequence
	//  - How many times the runtime has looped around this full shuffle
	seqPathStr := seqContainer.Path(seqContainer).String()
	sequenceHash := 0
	for _, c := range seqPathStr {
		sequenceHash += int(c)
	}
	randomSeed := sequenceHash + loopIndex + s.State().StorySeed
	random := rand.New(rand.NewSource(int64(randomSeed)))
	var unpickedIndices []int
	for i := 0; i < numElements; i++ {
		unpickedIndices = append(unpickedIndices, i)
	}

	for i := 0; i <= iterationIndex; i++ {
		chosen := random.Int() % len(unpickedIndices)
		chosenIndex := unpickedIndices[chosen]
		unpickedIndices = append(unpickedIndices[:chosen], unpickedIndices[chosen+1:]...)

		if i == iterationIndex {
			return chosenIndex
		}
	}

	panic("Should never reach here")
}

// (default) useEndLineNumber: false
func (s *Story) Error(message string) {
	panic(message)
}

func (s *Story) Warning(message string) {

	s.AddError(message, false, false)
}

// (default) isWarning: false
// (default) useEndLineNumber: false
func (s *Story) AddError(message string, isWarning bool, useEndLineNumber bool) {

	dm := s.CurrentDebugMetadata()

	errorTypeStr := "ERROR"
	if isWarning {
		errorTypeStr = "WARNING"
	}

	if dm != nil {
		lineNum := 0
		if useEndLineNumber {
			lineNum = dm.EndLineNumber
		} else {
			lineNum = dm.StartLineNumber
		}
		message = fmt.Sprintf("RUNTIME %s: '%s' line %d: %s", errorTypeStr, dm.FileName, lineNum, message)
	} else if !s.State().CurrentPointer().IsNull() {
		message = fmt.Sprintf("RUNTIME %s: (%s): %s", errorTypeStr, s.State().CurrentPointer().Path().String(), message)
	} else {
		message = "RUNTIME " + errorTypeStr + ": " + message
	}

	s.State().AddError(message, isWarning)

	// In a broken state don't need to know about any other errors.
	if !isWarning {
		s.State().ForceEnd()
	}
}

func (s *Story) CurrentDebugMetadata() *DebugMetadata {

	var dm *DebugMetadata

	// Try to get from the current path first
	pointer := s.State().CurrentPointer()
	if !pointer.IsNull() {
		dm = pointer.Resolve().DebugMetadata()
		if dm != nil {
			return dm
		}
	}

	// Move up callstack if possible
	for i := len(s.State().CallStack().Elements()) - 1; i >= 0; i-- {
		pointer = s.State().CallStack().Elements()[i].CurrentPointer
		if !pointer.IsNull() && pointer.Resolve() != nil {
			dm = pointer.Resolve().DebugMetadata()
			if dm != nil {
				return dm
			}
		}
	}

	// Current/previous path may not be valid if we've just had an error,
	// or if we've simply run out of content.
	// As a last resort, try to grab something from the output stream
	for i := len(s.State().OutputStream()) - 1; i >= 0; i-- {
		outputObj := s.State().OutputStream()[i]
		dm = outputObj.DebugMetadata()
		if dm != nil {
			return dm
		}
	}

	return nil
}

/*
   int currentLineNumber
   {
       get {
           var dm = currentDebugMetadata;
           if (dm != null) {
               return dm.startLineNumber;
           }
           return 0;
       }
   }
*/

func (s *Story) MainContentContainer() *Container {
	if s._temporaryEvaluationContainer != nil {
		return s._temporaryEvaluationContainer
	} else {
		return s._mainContentContainer
	}
}
