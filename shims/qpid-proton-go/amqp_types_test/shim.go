package main

import (
	"log"
	"net"
	"net/url"
	"time"

	"qpid.apache.org/electron"
	"encoding/json"
	"strings"
	"strconv"
	"qpid.apache.org/amqp"
	"fmt"
	"unicode/utf8"
	"unicode"
	"math"
)

func must(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func connect(url *url.URL) (c electron.Connection, err error) {
	timeout := 15 * time.Second
	tcp, err := net.DialTimeout("tcp", url.Host, timeout)
	if err != nil {
		return
	}

	container := electron.NewContainer("")

	opts := make([]electron.ConnectionOption, 0)
	opts = append(opts, electron.SASLAllowInsecure(true))
	if url.User != nil {
		if u := url.User.Username(); u != "" {
			//glog.V(2).Infof("User: %v", u)
			opts = append(opts, electron.User(u))
		}
		if p, ok := url.User.Password(); ok {
			//glog.V(2).Infof("Password: %v", p)
			opts = append(opts, electron.Password([]byte(p)))
		}
	}

	c, err = container.Connection(tcp, opts...)
	return
}

func parse(type_ string, value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		switch type_ {
		case "binary":
			return amqp.Binary(v)
		case "boolean":
			b, err := strconv.ParseBool(v)
			must(err)
			return b
		case "byte":
			b, err := strconv.ParseInt(v, 0, 8)
			must(err)
			return int8(b) // AMQP byte is not a Go byte
		case "decimal64", "decimal128":
			// TODO: no idea how to send this, let's send nothing
			return nil
		case "int":
			i, err := strconv.ParseInt(v, 0, 32)
			must(err)
			return int32(i)
		case "long":
			i, err := strconv.ParseInt(v, 0, 64)
			must(err)
			return int64(i)
		case "null", "none":  // TODO: this is ugly
			return nil
		case "short":
			i, err := strconv.ParseInt(v, 0, 16)
			must(err)
			return int16(i)
		case "string":
			return string(v)
		case "symbol":
			return amqp.Symbol(v)
		case "timestamp":
			// TODO: no idea how to send this, let's send nothing
			return nil
		case "ubyte":
			i, err := strconv.ParseUint(v, 0, 8)
			must(err)
			return uint8(i)
		case "uint":
			i, err := strconv.ParseUint(v, 0, 32)
			must(err)
			return uint32(i)
		case "ulong":
			i, err := strconv.ParseUint(v, 0, 64)
			must(err)
			return uint64(i)
		case "ushort":
			i, err := strconv.ParseUint(v, 0, 16)
			must(err)
			return uint16(i)
		case "uuid":
			// TODO: no idea how to send this, let's send nothing
			return nil
		case "float":
			f, err := strconv.ParseFloat(v, 32)
			if err != nil {
				// hex bytes
				f, err := strconv.ParseUint(v, 0, 32)
				must(err)
				return math.Float32frombits(uint32(f))
			}
			return float32(f)
		case "double":
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				// hex bytes
				f, err := strconv.ParseUint(v, 0, 64)
				must(err)
				return math.Float64frombits(f)
			}
			return float64(f)
		case "char":
			// hex bytes
			f, err := strconv.ParseUint(v, 0, 32)
			if err != nil {
				// ascii
				if utf8.RuneCountInString(v) != 1 {
					log.Fatal("char contains more or less than one utf-8 char")
				}
				return rune(v[0])
			}
			return rune(f)
		default:
			log.Fatalf("Wrong type %v given string value %v\n", type_, v)
		}
	case []interface{}:
		// assume it is []string
		switch type_ {
		case "list":
			l := make([]interface{}, 0)
			for _, i := range v {
				switch s := i.(type) {
				case string:
					tuple := strings.SplitN(s, ":", 2)
					if len(tuple) == 2 {
						l = append(l, parse(tuple[0], tuple[1]))
					} else {
						l = append(l, s)
					}
				case []interface{}:
					//fmt.Printf("parsing list in a list %+v\n", s)
					// TODO: it it is always empty lists, I can just throw it there directly;... better not do that
					l = append(l, parse("list", s))
				case map[string]interface{}:
					// there is a map[string]interface{}, but lets be general; actually, that was stupid idea, this is json....
					l = append(l, parse("map", s))
				default:
					// TODO: it can be a map? likely not!
					log.Fatalf("neither value nor nested list in a list, the type is %T", i)
				}
			}
			return l
		}
	case map[string]interface{}:
		// assume it is map[string]string
		switch type_ {
		case "map":
			m := make(map[interface{}]interface{})
			for k, v := range v {
				var kk interface{}
				var vv interface{}
				// FIXME: duplicates the list case
				tuple := strings.SplitN(k, ":", 2)
				if len(tuple) == 2 {
					kk = parse(tuple[0], tuple[1])
				} else {
					kk = k
				}
				switch v := v.(type) {
				case string:
					tuple := strings.SplitN(v, ":", 2)
					if len(tuple) == 2 {
						vv = parse(tuple[0], tuple[1])
					} else {
						vv = v
					}
				case []interface{}:
					vv = parse("list", v)
				case map[string]interface{}:
					vv = parse("map", v)
				default:
					log.Fatalf("neither value nor nested map a map key, value %v type is %T", v, v)
				}
				m[kk] = vv
			}
			return m
		}
	}
	log.Fatalf("Wrong type %v given value %+v of type %T\n", type_, value, value)
	return nil
}

func formatUnsigned(v int64) string {
	sign := ""
	if v < 0 {
		sign = "-"
	}
	return fmt.Sprintf("%s0x%x", sign, v)
}

// load is inverse function to parse
func load(type_ string, body interface{}) interface{} {
	switch type_ {
	case "ubyte", "ushort", "uint", "ulong":
		return fmt.Sprintf("%#x", body)
	case "byte", "short", "int", "long":
		return fmt.Sprintf("%#x", body)
	case "float":
		// TODO: what's more reasonable criterion?
		f := body.(float32)
		if (-10 <= f) && (f <= 10) {
			return fmt.Sprint(body)
		}
		return fmt.Sprintf("%#x", math.Float32bits(f))
	case "double":
		// TODO: what's more reasonable criterion?
		f := body.(float64)
		if (-10 <= f) && (f <= 10) {
			return fmt.Sprint(body)
		}
		return fmt.Sprintf("%#x", math.Float64bits(f))
	case "char":
		r := body.(rune)
		if r <= unicode.MaxASCII && (unicode.IsDigit(r) || unicode.IsLetter(r)) || unicode.IsSpace(r) {
			return fmt.Sprintf("%c", body)
		}
		return fmt.Sprintf("%#x", body)
	case "boolean":
		return strings.Title(fmt.Sprint(body))
	case "string":
		return body
	case "null":
		return "None"
	case "symbol":
		return fmt.Sprint(body)
	case "list":
		l := make([]interface{}, 0)
		for _, v := range body.([]interface{}) {
			loadedt, loadedv := loadValue(v)
			switch loadedt {
			case "list", "map":
				l = append(l, loadedv)
			default:
				l = append(l, fmt.Sprintf("%s:%s", loadedt, loadedv))
			}
		}
		return l
	case "map":
		m := make(map[string]interface{})
		for k, v := range body.(map[interface{}]interface{}) {
			var kk string
			keyt, keyv := loadValue(k)
			switch keyt {
			case "list", "map":
				//todo: not supported
				log.Fatalln("load: cannot hancle list or map as map key yet")
				//kk = keyv
			default:
				kk = fmt.Sprintf("%s:%s", keyt, keyv.(string))
			}

			loadedt, loadedv := loadValue(v)
			switch loadedt {
			case "list", "map":
				m[kk] = loadedv
			default:
				m[kk] = fmt.Sprintf("%s:%s", loadedt, loadedv)
			}
		}
		return m
	}
	log.Panicf("cannot decode %v", body)
	return nil
}

func parseValue(v interface{}) interface{} {
	return nil
}

var GoToAMQPMapping = map[string]string {
	"uint8": "ubyte",
	"uint16": "ushort",
	"uint32": "uint",
	"uint64": "ulong",
	"int8": "byte",
	"int16": "short",
	"int32": "int",
	"int64": "long",
}

// loadValue loads a list item or map key
func loadValue(v interface{}) (string, interface{}) {
	switch body := v.(type) {
	case nil:
		return "none", ""
	case string:
		return "string", body
	case uint8, uint16, uint32, uint64, int8, int16, int32, int64:
		return GoToAMQPMapping[fmt.Sprintf("%T", body)], fmt.Sprintf("%d", body) // d instead of x
	case float32:
		// TODO: what's more reasonable criterion?
		if (-10 <= body) && (body <= 10) {
			return "float", fmt.Sprint(body)
		}
		return "float", fmt.Sprintf("%#x", math.Float32bits(body))
	case float64:
		// TODO: what's more reasonable criterion?
		if (-10 <= body) && (body <= 10) {
			return "double", fmt.Sprint(body)
		}
		return "double", fmt.Sprintf("%#x", math.Float64bits(body))
	// TODO: i cannot distinguish int32 and rune in Go
	//case rune:
	//	r := body
	//	if r <= unicode.MaxASCII && (unicode.IsDigit(r) || unicode.IsLetter(r)) || unicode.IsSpace(r) {
	//		return "char", fmt.Sprintf("%c", body)
	//	}
	//	return "char", fmt.Sprintf("%#x", body)
	case bool:
		return "boolean", strings.Title(fmt.Sprint(body))
	//case string:
	//	return "string", body
	//case nil:
	//	return "null", "None"
	case amqp.Symbol:
		return "symbol", fmt.Sprint(body)
	case []interface{}:
		return "list", load("list", v)
	case map[interface{}]interface{}:
		return "map", load("map", v)
	default:
		log.Panicf("loadValue: cannot decode %v of type %T", v, v)
		return "", nil
	}
}

func toString(bodies []interface{}) string {
	j, err := json.Marshal(bodies)
	must(err)
	return string(j)
}

