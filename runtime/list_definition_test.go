package runtime_test

import (
	"fmt"
	"github.com/SirMetathyst/go-ink/runtime"
	"testing"
)

func TestListDefinition(t *testing.T) {

	items := make(map[string]int, 0)
	items["item1"] = 10
	items["item2"] = 20
	items["item3"] = 30
	items["item4"] = 40

	listDef := runtime.NewListDefinition("testName", items)

	fmt.Println(listDef.Name())

	for k, v := range listDef.Items() {
		fmt.Println(k)
		fmt.Println(v)
	}

}
