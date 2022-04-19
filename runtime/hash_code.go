package ink

import "hash/fnv"

func hashCodeFromString(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}
