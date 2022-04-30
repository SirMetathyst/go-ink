package runtime

func TextToMap(text string) map[string]interface{} {
	return NewReader(text).ToMap()
}

func TextToArray(text string) []interface{} {
	return NewReader(text).ToArray()
}
