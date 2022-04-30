package runtime

func JArrayToRuntimeObjList[T Object](jArray []interface{}, skipLast bool) []T {
	count := len(jArray)

	if skipLast {
		count--
	}

	var list []T

	for i := 0; i < count; i++ {
		jTok := jArray[i]
		runtimeObj := JTokenToRuntimeObject(jTok).(T)
		list = append(list, runtimeObj)
	}

	return list
}

func WriteMapRuntimeObjs(writer *Writer, collection map[string]Object) {

	writer.WriteObjectStart()
	for key, value := range collection {
		writer.WritePropertyStart(key)
		WriteRuntimeObject(writer, value)
		writer.WritePropertyEnd()
	}
	writer.WriteObjectEnd()
}

func WriteListRuntimeObjs(writer *Writer, list []Object) {
	writer.WriteArrayStart()
	for _, val := range list {
		WriteRuntimeObject(writer, val)
	}
	writer.WriteArrayEnd()
}

func WriteIntMap(writer *Writer, collection map[string]int) {

	writer.WriteObjectStart()
	for key, value := range collection {
		writer.WriteIntProperty(key, value)
	}
	writer.WriteObjectEnd()
}

func WriteRuntimeObject(writer *Writer, obj Object) {

	container, ok := obj.(*Container)
	if ok {
		WriteRuntimeContainer(writer, container)
		return
	}

	divert, ok := obj.(*Divert)
	if ok {

		divTypeKey = "->"
	}
}
