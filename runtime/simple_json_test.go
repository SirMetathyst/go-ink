package runtime_test

import (
	"fmt"
	"github.com/SirMetathyst/go-ink/runtime/simple_json"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

var (
	TextToDictionary = []string{
		"KEY:`inkVersion`",
		"INT:`11`",
		"KEY:`root`",
		"INDEX:`0`",
		"INDEX:`0`",
		"INDEX:`0`",
		"STRING:`G>`",
		"INDEX:`1`",
		"INDEX:`0`",
		"INDEX:`0`",
		"STRING:`ev`",
		"INDEX:`1`",
		"KEY:`VAR?`",
		"STRING:`DEBUG`",
		"INDEX:`2`",
		"STRING:`/ev`",
		"INDEX:`3`",
		"NULL",
		"INDEX:`1`",
		"INDEX:`0`",
		"STRING:`du`",
		"INDEX:`1`",
		"KEY:`t?`",
		"KEY:`->`",
		"STRING:`0.0.1.1.b`",
		"INDEX:`2`",
		"KEY:`b`",
		"INDEX:`0`",
		"STRING:`pop`",
		"INDEX:`1`",
		"STRING:`^IN DEBUG MODE!`",
		"INDEX:`2`",
		"STRING:\n",
		"INDEX:`3`",
		"INDEX:`0`",
		"STRING:`ev`",
		"INDEX:`1`",
		"STRING:`str`",
		"INDEX:`2`",
		"STRING:`^Beginning...`",
		"INDEX:`3`",
		"STRING:`/str`",
		"INDEX:`4`",
		"STRING:`/ev`",
		"INDEX:`5`",
		"KEY:`*`",
		"STRING:`.^.c`",
		"KEY:`flg`",
		"INT:`20`",
		"INDEX:`6`",
		"KEY:`c`",
		"INDEX:`0`",
		"KEY:`->`",
		"STRING:`start`",
		"INDEX:`1`",
		"KEY:`#f`",
		"INT:`7`",
		"INDEX:`4`",
		"INDEX:`0`",
		"STRING:`ev`",
		"INDEX:`1`",
		"STRING:`str`",
		"INDEX:`2`",
		"STRING:`^Framing Hooper...`",
		"INDEX:`3`",
		"STRING:`/str`",
		"INDEX:`4`",
		"STRING:`/ev`",
		"INDEX:`5`",
		"KEY:`*`",
		"STRING:`.^.c`",
		"KEY:`flg`",
		"INT:`20`",
		"INDEX:`6`",
		"KEY:`c`",
		"INDEX:`0`",
		"KEY:`->`",
		"STRING:`claim_hooper_took_component`",
		"INDEX:`1`",
		"KEY:`#f`",
		"INT:`7`",
		"INDEX:`5`",
		"INDEX:`0`",
		"STRING:`ev`",
		"INDEX:`1`",
		"STRING:`str`",
		"INDEX:`2`",
		"STRING:`^In with Hooper...`",
		"INDEX:`3`",
		"STRING:`/str`",
		"INDEX:`4`",
		"STRING:`/ev`",
		"INDEX:`5`",
		"KEY:`*`",
		"STRING:`.^.c`",
		"KEY:`flg`",
		"INT:`20`",
		"INDEX:`6`",
		"KEY:`c`",
		"INDEX:`0`",
		"KEY:`->`",
		"STRING:`inside_hoopers_hut`",
		"INDEX:`1`",
		"KEY:`#f`",
		"INT:`7`",
		"INDEX:`6`",
		"KEY:`->`",
		"STRING:`0.0.1.4`",
		"INDEX:`7`",
		"NULL",
		"INDEX:`2`",
		"INDEX:`0`",
		"KEY:`f?`",
		"KEY:`->`",
		"STRING:`0.0.1.2.b`",
		"INDEX:`1`",
		"KEY:`b`",
		"INDEX:`0`",
		"KEY:`->`",
		"STRING:`start`",
		"INDEX:`1`",
		"KEY:`->`",
		"STRING:`0.0.1.4`",
		"INDEX:`2`",
		"NULL",
		"INDEX:`3`",
		"STRING:`pop`",
		"INDEX:`4`",
		"STRING:`nop`",
		"INDEX:`5`",
		"NULL",
		"INDEX:`2`",
		"STRING:`G<`",
		"INDEX:`3`",
		"NULL",
		"INDEX:`1`",
		"STRING:\n",
		"INDEX:`2`",
		"NULL",
		"INDEX:`1`",
		"STRING:`done`",
		"INDEX:`2`",
	}
)

func TestTextToDictionary(t *testing.T) {

	jsonBytes, err := os.ReadFile("TheIntercept.json")
	assert.Nil(t, err)

	dict := simple_json.TextToDictionary(string(jsonBytes))
	Walk(&Walker{t: t, node: dict, expect: TextToDictionary})
}

var (
	TextToArray = []string{
		"INDEX:`0`",
		"INT:`100`",
		"INDEX:`1`",
		"INT:`200`",
		"INDEX:`2`",
		"INT:`300`",
		"INDEX:`3`",
		"INT:`400`",
	}
)

func TestTextToArray(t *testing.T) {

	arr := simple_json.TextToArray("[100,200,300,400]")
	Walk(&Walker{t: t, node: arr, expect: TextToArray})
}

type Walker struct {
	t      *testing.T
	node   interface{}
	expect []string
	index  int
}

func Walk(w *Walker) {

	if w.index >= len(w.expect) {
		return
	}

	switch v := w.node.(type) {

	case map[string]interface{}:
		for k, v := range v {

			//fmt.Printf("KEY:`%s`\n", k)
			assert.Equal(w.t, w.expect[w.index], fmt.Sprintf("KEY:`%s`", k))
			w.index++
			w.node = v
			Walk(w)

			if w.index >= len(w.expect) {
				return
			}
		}

	case []interface{}:
		for index, v := range v {

			//fmt.Printf("INDEX:`%d`\n", index)
			assert.Equal(w.t, w.expect[w.index], fmt.Sprintf("INDEX:`%d`", index))
			w.index++
			w.node = v
			Walk(w)

			if w.index >= len(w.expect) {
				return
			}
		}

	case string:
		if strings.Contains(v, "\n") {

			//fmt.Printf("STRING:`%s`\n", strings.ReplaceAll(v, "\n", "\\n"))
			assert.Equal(w.t, w.expect[w.index], "STRING:\n")
			w.index++
			break
		}

		//fmt.Printf("STRING:`%s`\n", v)
		assert.Equal(w.t, w.expect[w.index], fmt.Sprintf("STRING:`%s`", v))
		w.index++

	case int:
		//fmt.Printf("INT:`%d`\n", v)
		assert.Equal(w.t, w.expect[w.index], fmt.Sprintf("INT:`%d`", v))
		w.index++

	case bool:
		//fmt.Printf("BOOL:`%v`\n", v)
		assert.Equal(w.t, w.expect[w.index], fmt.Sprintf("BOOL:`%v`", v))
		w.index++

	default:
		if v != nil {
			panic(fmt.Sprintf("Unknown %v", v))
		}

		//fmt.Println("NULL")
		assert.Equal(w.t, w.expect[w.index], "NULL")
		w.index++
	}
}
