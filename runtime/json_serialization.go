package runtime

import (
	"fmt"
)

func JArrayToRuntimeObjList[T any](jArray []interface{}, skipLast bool) []T {

	count := len(jArray)
	if skipLast {
		count--
	}

	var list []T
	for i := 0; i < count; i++ {
		runtimeObj := JTokenToRuntimeObject(jArray[i]).(T)
		list = append(list, runtimeObj)
	}

	return list
}

func WriteDictionaryRuntimeObjs(writer *Writer, dictionary map[string]Object) {

	writer.WriteObjectStart()
	for k, v := range dictionary {
		writer.WritePropertyStart(k)
		WriteRuntimeObject(writer, v)
		writer.WritePropertyEnd()
	}
	writer.WriteObjectEnd()
}

func WriteListRuntimeObjs(writer *Writer, list []Object) {
	writer.WriteArrayStart()
	for _, v := range list {
		WriteRuntimeObject(writer, v)
	}
	writer.WriteArrayEnd()
}

func WriteIntDictionary(writer *Writer, dict map[string]int) {
	writer.WriteObjectStart()
	for k, v := range dict {
		writer.WriteIntProperty(k, v)
	}
	writer.WriteObjectEnd()
}

func WriteRuntimeObject(writer *Writer, obj Object) {

	container, _ := obj.(*Container)
	if container != nil {
		WriteRuntimeContainer(writer, container, false)
		return
	}

	divert, _ := obj.(*Divert)
	if divert != nil {

		divTypeKey := "->"
		if divert.IsExternal {

			divTypeKey = "x()"
		} else if divert.PushesToStack {

			if divert.StackPushType == Function {
				divTypeKey = "f()"
			} else if divert.StackPushType == Tunnel {
				divTypeKey = "->t->"
			}
		}

		targetStr := ""
		if divert.HasVariableTarget() {
			targetStr = divert.VariableDivertName
		} else {
			targetStr = divert.TargetPathString()
		}

		writer.WriteObjectStart()
		writer.WriteStringProperty(divTypeKey, targetStr)

		if divert.HasVariableTarget() {
			writer.WriteBoolProperty("var", true)
		}

		if divert.IsConditional {
			writer.WriteBoolProperty("c", true)
		}

		if divert.ExternalArgs > 0 {
			writer.WriteIntProperty("exArgs", divert.ExternalArgs)
		}

		writer.WriteObjectEnd()
		return
	}

	choicePoint, _ := obj.(*ChoicePoint)
	if choicePoint != nil {
		writer.WriteObjectStart()
		writer.WriteStringProperty("*", choicePoint.PathStringOnChoice())
		writer.WriteIntProperty("flg", choicePoint.Flags())
		writer.WriteObjectEnd()
		return
	}

	boolVal, _ := obj.(*BoolValue)
	if boolVal != nil {
		writer.WriteBool(boolVal.Value())
		return
	}

	intVal, _ := obj.(*IntValue)
	if intVal != nil {
		writer.WriteInt(intVal.Value())
		return
	}

	floatVal, _ := obj.(*FloatValue)
	if floatVal != nil {
		writer.WriteFloat(floatVal.Value())
		return
	}

	strVal, _ := obj.(*StringValue)
	if strVal != nil {
		if strVal.IsNewline() {
			writer.WriteString("\\n", false)
		} else {
			writer.WriteStringStart()
			writer.WriteStringInner("^", true)
			writer.WriteStringInner(strVal.Value(), true)
			writer.WriteStringEnd()
		}
		return
	}

	listVal, _ := obj.(*ListValue)
	if listVal != nil {
		WriteInkList(writer, listVal)
		return
	}

	divTargetVal, _ := obj.(*DivertTargetValue)
	if divTargetVal != nil {
		writer.WriteObjectStart()
		writer.WriteStringProperty("^->", divTargetVal.Value().ComponentsString())
		writer.WriteObjectEnd()
		return
	}

	varPtrVal, _ := obj.(*VariablePointerValue)
	if varPtrVal != nil {
		writer.WriteObjectStart()
		writer.WriteStringProperty("^var", varPtrVal.Value())
		writer.WriteIntProperty("ci", varPtrVal.ContextIndex())
		writer.WriteObjectEnd()
		return
	}

	glue, _ := obj.(*VariablePointerValue)
	if glue != nil {
		writer.WriteString("<>", true)
		return
	}

	controlCmd, _ := obj.(*ControlCommand)
	if controlCmd != nil {
		v, _ := controlCommandNames[controlCmd.CommandType]
		writer.WriteString(v, true)
		return
	}

	nativeFunc, _ := obj.(*NativeFunctionCall)
	if nativeFunc != nil {
		name := nativeFunc.Name()

		// Avoid collision with ^ used to indicate a string
		if name == "^" {
			name = "L^"
		}

		writer.WriteString(name, true)
		return
	}

	varRef, _ := obj.(*VariableReference)
	if varRef != nil {

		writer.WriteObjectStart()

		readCountPath, ok := varRef.PathStringForCount()
		if ok {
			writer.WriteStringProperty("CNT?", readCountPath)
		} else {
			writer.WriteStringProperty("VAR?", varRef.Name)
		}

		writer.WriteObjectEnd()
		return
	}

	varAss, _ := obj.(*VariableAssignment)
	if varAss != nil {
		writer.WriteObjectStart()
		key := "VAR="
		if !varAss.IsGlobal {
			key = "temp="
		}
		writer.WriteStringProperty(key, varAss.VariableName())

		// Reassignment?
		if !varAss.IsNewDeclaration() {
			writer.WriteBoolProperty("re", true)
		}

		writer.WriteObjectEnd()
		return
	}

	voidObj, _ := obj.(*Void)
	if voidObj != nil {
		writer.WriteString("void", true)
		return
	}

	tag, _ := obj.(*Tag)
	if tag != nil {
		writer.WriteObjectStart()
		writer.WriteStringProperty("#", tag.text)
		writer.WriteObjectEnd()
		return
	}

	// Used when serialising save state only
	choice, _ := obj.(*Choice)
	if choice != nil {
		WriteChoice(writer, choice)
		return
	}

	panic(fmt.Sprintf("Failed to write runtime object to JSON: %v", obj))
}

func JObjectToDictionaryRuntimeObjs(jObject map[string]interface{}) map[string]Object {

	dict := make(map[string]Object)

	for k, v := range jObject {

		if _, ok := dict[k]; ok {
			panic("key already in map")
		}

		dict[k] = JTokenToRuntimeObject(v)
	}

	return dict
}

func JObjectToIntDictionary(jObject map[string]interface{}) map[string]int {

	dict := make(map[string]int, 0)

	for k, v := range jObject {
		dict[k] = v.(int)
	}

	return dict
}

// JTokenToRuntimeObject
// ----------------------
// JSON ENCODING SCHEME
// ----------------------
//
// Glue:           "<>", "G<", "G>"
//
// ControlCommand: "ev", "out", "/ev", "du" "pop", "->->", "~ret", "str", "/str", "nop",
//                 "choiceCnt", "turns", "visit", "seq", "thread", "done", "end"
//
// NativeFunction: "+", "-", "/", "*", "%" "~", "==", ">", "<", ">=", "<=", "!=", "!"... etc
//
// Void:           "void"
//
// Value:          "^string value", "^^string value beginning with ^"
//                 5, 5.2
//                 {"^->": "path.target"}
//                 {"^var": "varname", "ci": 0}
//
// Container:      [...]
//                 [...,
//                     {
//                         "subContainerName": ...,
//                         "#f": 5,                    // flags
//                         "#n": "containerOwnName"    // only if not redundant
//                     }
//                 ]
//
// Divert:         {"->": "path.target", "c": true }
//                 {"->": "path.target", "var": true}
//                 {"f()": "path.func"}
//                 {"->t->": "path.tunnel"}
//                 {"x()": "externalFuncName", "exArgs": 5}
//
// Var Assign:     {"VAR=": "varName", "re": true}   // reassignment
//                 {"temp=": "varName"}
//
// Var ref:        {"VAR?": "varName"}
//                 {"CNT?": "stitch name"}
//
// ChoicePoint:    {"*": pathString,
//                  "flg": 18 }
//
// Choice:         Nothing too clever, it's only used in the save state,
//                 there's not likely to be many of them.
//
// Tag:            {"#": "the tag text"}
func JTokenToRuntimeObject(token interface{}) Object {

	_, isInt := token.(int)
	_, isFloat := token.(float64)
	_, isBool := token.(bool)

	if isInt || isFloat || isBool {
		fmt.Println("Create Value: ", token)
		return CreateValue(token)
	}

	if str, ok := token.(string); ok {

		// String value
		firstChar := str[0]

		if firstChar == '^' {
			fmt.Println("String value", str[1:])
			return NewStringValueFromString(str[1:])
		}

		// String value (newline)
		if firstChar == '\n' && len(str) == 1 {
			fmt.Println("String value: \\n")
			return NewStringValueFromString("\n")
		}

		// Glue
		if str == "<>" {
			fmt.Println("Glue")
			return NewGlue()
		}

		// Control commands (would looking up in a hash set be faster?)
		for i := 0; i < len(controlCommandNames); i++ {
			cmdName, isInMap := controlCommandNames[CommandType(i)]
			if str == cmdName {
				fmt.Println("CommandType: ", CommandType(i))
				return NewControlCommand(CommandType(i))
			}
			if !isInMap {
				panic("unknown command type")
			}
		}

		// Native functions
		// "^" conflicts with the way to identify strings, so now
		// we know it's not a string, we can convert back to the proper
		// symbol for the operator.
		if str == "L^" {
			str = "^"
		}
		if CallExistsWithName(str) {
			fmt.Println("Native Function: ", str)
			return NewNativeFunctionCallFromName(str)
		}

		// Pop
		if str == "->->" {
			fmt.Println("Pop: ->->")
			return NewPopFunctionCommand()
		}

		if str == "~ret" {
			fmt.Println("~ret", str)
			return NewPopFunctionCommand()
		}

		// Void
		if str == "void" {
			fmt.Println("Void")
			return NewVoid()
		}
	}

	if obj, ok := token.(map[string]interface{}); ok {

		// Divert target value to path
		if propValue, ok := obj["^->"]; ok {
			fmt.Println("Divert Target", propValue.(string))
			//path := NewPathFromString(propValue.(string))
			//fmt.Println("Path Resolve: ", path.String())
			return NewDivertTargetValueFromPath(NewPathFromString(propValue.(string)))
		}

		// VariablePointerValue
		if propValue, ok := obj["^var"]; ok {

			varPtr := NewVariablePointerValueFromValue(propValue.(string), -1)
			if propValue, ok = obj["ci"]; ok {
				varPtr.SetContextIndex(propValue.(int))
			}
			fmt.Println("VariablePointerValue: ", varPtr.String())
			return varPtr
		}

		// Divert
		isDivert := false
		var propValue interface{}
		pushesToStack := false
		divPushType := Function
		external := false

		if propValue, ok = obj["->"]; ok {
			isDivert = true
		} else if propValue, ok = obj["f()"]; ok {

			isDivert = true
			pushesToStack = true
			divPushType = Function

		} else if propValue, ok = obj["->t->"]; ok {

			isDivert = true
			pushesToStack = true
			divPushType = Tunnel

		} else if propValue, ok = obj["x()"]; ok {
			isDivert = true
			external = true
			pushesToStack = false
			divPushType = Function
		}

		if isDivert {

			divert := NewDivert()
			divert.PushesToStack = pushesToStack
			divert.StackPushType = divPushType
			divert.IsExternal = external

			target := propValue.(string)
			if propValue, ok = obj["var"]; ok {
				divert.VariableDivertName = target
			} else {
				divert.SetTargetPathString(target)
			}

			propValue, divert.IsConditional = obj["c"]

			if external {
				if propValue, ok = obj["exArgs"]; ok {
					divert.ExternalArgs = propValue.(int)
				}
			}

			return divert
		}

		// Choice
		if propValue, ok = obj["*"]; ok {
			choice := NewChoicePoint()
			choice.SetPathStringOnChoice(propValue.(string))

			if propValue, ok = obj["flg"]; ok {
				choice.SetFlags(propValue.(int))
			}

			fmt.Println("Choice: ", choice.String())
			return choice
		}

		// Variable reference
		if propValue, ok = obj["VAR?"]; ok {
			fmt.Println("Variable reference: ", propValue.(string))
			return NewVariableReferenceFromName(propValue.(string))
		} else if propValue, ok = obj["CNT?"]; ok {
			readCountVarRef := NewVariableReference()
			readCountVarRef.SetPathStringForCount(propValue.(string))

			fmt.Println("Variable reference: ", readCountVarRef.String())
			return readCountVarRef
		}

		// Variable assignment
		isVarAss := false
		isGlobalVar := false
		if propValue, ok = obj["VAR="]; ok {
			isVarAss = true
			isGlobalVar = true
		} else if propValue, ok = obj["temp="]; ok {
			isVarAss = true
			isGlobalVar = false
		}
		if isVarAss {
			varName := propValue.(string)
			_, isNewDecl := obj["re"]
			isNewDecl = !isNewDecl
			varAss := NewVariableAssignment(varName, isNewDecl)
			varAss.IsGlobal = isGlobalVar

			fmt.Println("Variable Assignment: ", varAss.String())
			return varAss
		}

		// Legacy Tag with text
		if propValue, ok = obj["#"]; ok {
			fmt.Println("Tag: ", propValue.(string))
			return NewTag(propValue.(string))
		}

		// List value
		if propValue, ok = obj["list"]; ok {
			listContent := propValue.(map[string]interface{})
			rawList := NewInkList()
			if propValue, ok = obj["origins"]; ok {
				nameAsObjs := propValue.([]interface{})
				var nameAsStr []string
				for _, v := range nameAsObjs {
					//nameAsStr.Add(v.(string))
					nameAsStr = append(nameAsStr, v.(string))
				}
				rawList.SetInitialOriginNames(nameAsStr)
				for k, v := range listContent {
					item := NewInkListFromFullname(k)
					val := v.(int)
					rawList.Add(item, val)
				}
				fmt.Println("Ink List: ", rawList)
				return NewListValueFromList(rawList)
			}
		}

		// Used when serialising save state only
		if propValue, ok = obj["originalChoicePath"]; ok {

			return JObjectToChoice(obj)
		}
	}

	// Array is always a Runtime.Container
	if obj, ok := token.([]interface{}); ok {

		fmt.Println("Container: ")
		return JArrayToContainer(obj)
	}

	if token == nil {
		fmt.Println("Nil Token")
		return nil
	}

	panic(fmt.Sprintf("Failed to convert token to runtime object: %v", token))
}

func WriteRuntimeContainer(writer *Writer, container *Container, withoutName bool) {

	writer.WriteArrayStart()

	for _, c := range container.Content() {
		WriteRuntimeObject(writer, c)
	}

	// Container is always an array [...]
	// But the final element is always either:
	//  - a dictionary containing the named content, as well as possibly
	//    the key "#" with the count flags
	//  - null, if neither of the above
	namedOnlyContent := container.NamedOnlyContent()
	countFlags := container.CountFlags()
	hasNameProperty := container.Name() != "" && !withoutName

	hasTerminator := namedOnlyContent != nil || countFlags > 0 || hasNameProperty

	if hasTerminator {
		writer.WriteObjectStart()
	}

	if namedOnlyContent != nil {
		for name, namedContainerInterface := range namedOnlyContent {
			namedContainer, _ := namedContainerInterface.(*Container)
			writer.WritePropertyStart(name)
			WriteRuntimeContainer(writer, namedContainer, true)
			writer.WritePropertyEnd()
		}
	}

	if countFlags > 0 {
		writer.WriteIntProperty("#f", countFlags)
	}

	if hasNameProperty {
		writer.WriteStringProperty("#n", container.Name())
	}

	if hasTerminator {
		writer.WriteObjectEnd()
	} else {
		writer.WriteNull()
	}

	writer.WriteArrayEnd()
}

func JArrayToContainer(jArray []interface{}) *Container {

	container := NewContainer()
	for _, c := range JArrayToRuntimeObjList[Object](jArray, true) {
		container.AddContent(c)
	}

	// Final object in the array is always a combination of
	//  - named content
	//  - a "#f" key with the countFlags
	// (if either exists at all, otherwise null)
	terminatingObj, _ := jArray[len(jArray)-1].(map[string]interface{})
	if terminatingObj != nil {

		namedOnlyContent := make(map[string]Object)

		for k, v := range terminatingObj {
			if k == "#f" {
				container.SetCountFlags(v.(int))
			} else if k == "#n" {
				container.SetName(v.(string))
			} else {
				namedContentItem := JTokenToRuntimeObject(v)
				namedSubContainer, _ := namedContentItem.(*Container)
				if namedSubContainer != nil {
					namedSubContainer.SetName(k)
				}
				namedOnlyContent[k] = namedContentItem
			}
		}

		container.SetNamedOnlyContent(namedOnlyContent)
	}

	return container
}

func JObjectToChoice(jObj map[string]interface{}) *Choice {

	choice := NewChoice()
	choice.Text = jObj["text"].(fmt.Stringer).String()
	choice.Index = jObj["index"].(int)
	choice.SourcePath = jObj["originalChoicePath"].(fmt.Stringer).String()
	choice.OriginalTheadIndex = jObj["originalThreadIndex"].(int)
	choice.SetPathStringOnChoice(jObj["targetPath"].(fmt.Stringer).String())

	fmt.Println("Choice: ", choice)

	return choice
}

func WriteChoice(writer *Writer, choice *Choice) {
	writer.WriteObjectStart()
	writer.WriteStringProperty("text", choice.Text)
	writer.WriteIntProperty("index", choice.Index)
	writer.WriteStringProperty("originalChoicePath", choice.SourcePath)
	writer.WriteIntProperty("originalThreadIndex", choice.OriginalTheadIndex)
	writer.WriteStringProperty("targetPath", choice.PathStringOnChoice())
	writer.WriteObjectEnd()
}

func WriteInkList(writer *Writer, listVal *ListValue) {

	rawList := listVal.Value()

	writer.WriteObjectStart()
	writer.WritePropertyStart("list")
	writer.WriteObjectStart()

	for item, itemVal := range rawList._items {

		writer.WritePropertyNameStart()

		x := item.OriginName()
		if x == "" {
			x = "?"
		}

		writer.WritePropertyNameInner(x)
		writer.WritePropertyNameInner(".")
		writer.WritePropertyNameInner(item.ItemName())
		writer.WritePropertyNameEnd()

		writer.WriteInt(itemVal)
		writer.WritePropertyEnd()
	}

	writer.WriteObjectEnd()
	writer.WritePropertyEnd()

	if rawList.Count() == 0 && rawList.OriginNames() != nil && len(rawList.OriginNames()) > 0 {
		writer.WritePropertyStart("origins")
		writer.WriteArrayStart()
		for _, name := range rawList.OriginNames() {
			writer.WriteString(name, true)
		}
		writer.WriteArrayEnd()
		writer.WritePropertyEnd()
	}

	writer.WriteObjectEnd()
}

func JTokenToListDefinitions(obj interface{}) *ListDefinitionsOrigin {

	defsObj := obj.(map[string]interface{})
	allDefs := []*ListDefinition{}

	for name, listDefJsonInterface := range defsObj {

		listDefJson := listDefJsonInterface.(map[string]interface{})

		items := make(map[string]int)
		for k, v := range listDefJson {

			if _, ok := items[k]; ok {
				panic("key already in map")
			}

			items[k] = v.(int)
		}

		def := NewListDefinition(name, items)
		//allDefs.Add(def)
		allDefs = append(allDefs, def)
	}

	return NewListDefinitionsOrigin(allDefs)
}

var (
	controlCommandNames = map[CommandType]string{
		CommandTypeEvalStart:            "ev",
		CommandTypeEvalOutput:           "out",
		CommandTypeEvalEnd:              "/ev",
		CommandTypeDuplicate:            "du",
		CommandTypePopEvaluatedValue:    "pop",
		CommandTypePopFunction:          "~ret",
		CommandTypePopTunnel:            "->->",
		CommandTypeBeginString:          "str",
		CommandTypeEndString:            "/str",
		CommandTypeNoOp:                 "nop",
		CommandTypeChoiceCount:          "choiceCnt",
		CommandTypeTurns:                "turn",
		CommandTypeTurnsSince:           "turns",
		CommandTypeReadCount:            "readc",
		CommandTypeRandom:               "rnd",
		CommandTypeSeedRandom:           "srnd",
		CommandTypeVisitIndex:           "visit",
		CommandTypeSequenceShuffleIndex: "seq",
		CommandTypeStartThread:          "thread",
		CommandTypeDone:                 "done",
		CommandTypeEnd:                  "end",
		CommandTypeListFromInt:          "listInt",
		CommandTypeListRange:            "range",
		CommandTypeListRandom:           "lrnd",
		CommandTypeBeginTag:             "#",
		CommandTypeEndTag:               "/#",
	}
)
