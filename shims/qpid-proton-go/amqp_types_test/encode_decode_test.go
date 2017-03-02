// Quick brokerless test to check that sender and receiver can
// Mostly useful for initial shim development.

package main

import (
	"encoding/json"
	"testing"
	"log"
	"fmt"
	"qpid.apache.org/amqp"
)

func TestMessageDecodeEncode(t *testing.T) {
	for _, in := range [][]string{
		[]string{"string", `["aString"]`},
		[]string{"string", `["aString","secondString"]`},

		[]string{"null", `["None"]`},
		[]string{"boolean", `["True"]`},
		[]string{"ubyte", `["0xff"]`},
		[]string{"ushort", `["0xffff"]`},
		[]string{"uint", `["0xffffffff"]`},
		[]string{"ulong", `["0xffffffffffffffff"]`},
		[]string{"byte", `["-0x80"]`},
		[]string{"short", `["-0x8000"]`},
		[]string{"int", `["-0x80000000"]`},
		[]string{"long", `["-0x8000000000000000"]`},

		[]string{"float", `["0xff7fffff"]`},
		[]string{"float", `["3.14"]`},
		[]string{"double", `["0xffefffffffffffff"]`},
		[]string{"double", `["3.14"]`},

		//[]string{"decimal32", `["0xff7fffff"]`},
		//[]string{"decimal64", `["0xffefffffffffffff"]`},
		//[]string{"decimal128", `["0xff0102030405060708090a0b0c0d0e0f"]`},
		[]string{"char", `["G"]`},
		//[]string{"char", `["紅"]`}, // this gets printed back in hex, so I cannot enable it
		[]string{"char", `["0x16b5"]`},
		//[]string{"timestamp", `["0xdc6acfac00"]`},
		//[]string{"uuid", `["00010203-0405-0607-0809-0a0b0c0d0e0f"]`},
		//[]string{"binary", `["\\x01\\x02\\x03\\x04\\x05abcde\\x80\\x81\\xfe\\xff"]`},
		[]string{"string", `["Hello, World!"]`},
		[]string{"symbol", `["myDomain.123"]`},
		[]string{"binary", `["someData"]`},

		[]string{"string", `[]`},
		[]string{"list", `[[]]`},
		[]string{"map", `[{}]`},

		[]string{"list", `[[[],[]]]`},
		[]string{"list", `[["string:v"]]`},
		[]string{"map", `[{"string:k":"string:v"}]`},
		[]string{"map", `[{"string:k":[]}]`},
		[]string{"map", `[{"string:k":{"string:k":"string:v"}}]`},

		[]string{"list", `[[[],[[],[[],[],[]],[]],[]]]`},
		[]string{"list", `[["ubyte:1"]]`},
		[]string{"list", `[["int:-2"]]`},
		[]string{"list", `[["float:3.14"]]`},
		[]string{"list", `[["string:a"]]`},
		[]string{"list", `[["ulong:12345"]]`},
		[]string{"list", `[["short:-2500"]]`},
		[]string{"list", `[["symbol:a.b.c"]]`},
		[]string{"list", `[["none:"]]`},
		//[]string{"list", `[["none"]]`},  // FIXME: which none is correct?
		[]string{"list", `[["boolean:True"]]`},
		[]string{"list", `[["long:1234"]]`},
		//[]string{"list", `[["char:A"]]`},  // TODO: in Go I cannot distinguish between int32 and rune

                  //'timestamp:%d' % (time()*1000),
                  //'uuid:%s' % uuid4(),
                  //'decimal64:0x400921fb54442eea'

		  //[]string{"map", `[{"string:one":"ubyte:1"}]`},
		  //[]string{"map", `[{["string:AAA","ushort:5951"]:"string:list value"}]"]`},
		  []string{"map", `[{"string:None":"none:"}]`},
		  []string{"map", `[{"none:":"string:None"}]`},

//		  []string{"map",
//`[
//{},
//{"string:one":"ubyte:1","string:two":"ushort:2"},
//{
//"boolean:True":"string:True",
//
//"short:2":"int:2",
//"string:One":"long:-1234567890",
//"string:False":"boolean:False",
//"string:None":"none:"
//"string:map":{"char:A":"int:1","char:B":"int:2"},
//}]`},

		  } {
		log.Println(in) // useful for debugging
		through := make([]interface{}, 0)
		var j []interface{}
		err := json.Unmarshal([]byte(in[1]), &j)
		must(err)
		for _, jj := range j {
			must(err)
			result := parse(in[0], jj)

			m := amqp.NewMessage()
			m.Marshal(result)

			type_, value := load(in[0], m.Body())
			if type_ != in[0] {
				t.Errorf("Detected type %v does not match expected type '%v'", type_, in[0])
			}
			through = append(through, value)
		}
		fmt.Printf("%+v\n", through) // useful for debugging
		resultstring := toString(through)
		if in[1] != resultstring {
			t.Errorf("Expected `%s`, actual `%s`.", in[1], resultstring)
			//panic("")
		}
		//text = `[[], {}, {"string:one": "ubyte:1", "string:two": "ushort:2"}, {"string:One": "long:-1234567890", "none:": "string:None", "short:2": "int:2", "string:map": {"char:A": "int:1", "char:B": "int:2"}, "boolean:True": "string:True", "string:False": "boolean:False", "string:None": "none:"}]]'`

	}
}
