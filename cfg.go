package cfg

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const envSep = '='
const envTag = "env"
const cmdRag = "cmd"

//len of array after splitting one environment variable string
const fieldLen = 2

type intSliceValue []int

func (o intSliceValue) String() string {
	return fmt.Sprintf("%#v", o)
}
func (o *intSliceValue) Set(s string) (err error) {
	si := make([]int, 0)
	var i int
	for _, v := range strings.Split(s, ",") {
		i, err = strconv.Atoi(v)
		si = append(si, i)
	}
	*o = si
	return nil
}
func (o *intSliceValue) Get() interface{} {
	return *o
}

func initCmdFlags(val reflect.Value, fs *flag.FlagSet) (err error) {
	for i := 0; i < val.NumField(); i++ {
		tf := val.Type().Field(i)
		vf := val.Field(i)
		if vf.Kind() == reflect.Struct {
			err = initCmdFlags(vf, fs)
			if err != nil {
				return
			}
			continue
		}
		name, usage := parseCmdFlagString(tf.Tag.Get(cmdRag))

		switch vkind := vf.Kind(); vkind {
		case reflect.String:
			fs.String(name, "", usage)
		case reflect.Int, reflect.Int64:
			if vkind == reflect.Int64 {
				fs.Int64(name, 0, usage)
			} else {
				fs.Int(name, 0, usage)
			}
		case reflect.Bool:
			fs.Bool(name, false, usage)
		case reflect.Slice:
			switch svk := vf.Type().Elem().Kind(); svk {
			case reflect.Int:
				fs.Var(&intSliceValue{}, name, usage)
			//TODO more cases with another types
			default:
				err = fmt.Errorf("initCmdFlags:%s: unsupported slice type: %v", name, svk)
				return
			}
		case reflect.Float64:
			fs.Float64(name, 0.0, usage)

		default:
			err = fmt.Errorf("initCmdFlags:%s: unsupported field kind: %s", name, vkind.String())
			return
		}

	}

	return
}

func parseCmdFlagString(flag string) (name string, usageString string) {
	data := strings.Split(flag, "|")
	if len(data) == 0 {
		return
	}
	if len(data) >= 1 {
		name = strings.Trim(data[0], " \n\r\t")
	}
	if len(data) >= 2 {
		usageString = strings.Trim(data[1], " \n\r\t")
	}
	return
}

func getFlags(val reflect.Value, fs *flag.FlagSet) (err error) {
	for i := 0; i < val.NumField(); i++ {
		tf := val.Type().Field(i)
		vf := val.Field(i)
		if vf.Kind() == reflect.Struct {
			err = getFlags(vf, fs)
			if err != nil {
				return
			}
			continue
		} else {
			name, _ := parseCmdFlagString(tf.Tag.Get(cmdRag))
			ff := fs.Lookup(name)
			if ff == nil {
				err = fmt.Errorf("lookup by name %s finished unsuccessfuly", name)
				return
			}
			getter, ok := ff.Value.(flag.Getter)
			if !ok {
				err = fmt.Errorf("%s can't cast to getter", name)
				return
			}
			vf.Set(reflect.ValueOf(getter.Get()))
		}
	}
	return
}

func getEnv(pref string, val reflect.Value) (err error) {
	for i := 0; i < val.NumField(); i++ {
		tf := val.Type().Field(i)
		vf := val.Field(i)
		if vf.Kind() == reflect.Struct {
			err = getEnv(pref, vf)
			if err != nil {
				return
			}
			continue
		} else {

			tagVal := tf.Tag.Get(envTag)
			if tagVal != "" {
				varname := pref + tagVal
				for _, e := range os.Environ() {
					var pair [fieldLen][]byte
					if pair, err = split([]byte(e), envSep); err != nil {
						return
					}
					if varname == string(pair[0]) && len(pair[1]) > 0 {
						stringValue := strings.ToLower(strings.Trim(string(pair[1]), " \n\t\r"))
						if err = parse(vf, stringValue, tagVal); err != nil {
							return
						}
					}
				}

			}
		}
	}
	return
}

func parse(vf reflect.Value, stringValue string, tagVal string) (err error) {
	switch vkind := vf.Kind(); vkind {
	case reflect.String:
		vf.SetString(stringValue)
	case reflect.Int, reflect.Int64:
		var intval int64
		if intval, err = strconv.ParseInt(stringValue, 10, 64); err != nil {
			return
		}
		if vkind == reflect.Int64 {
			vf.Set(reflect.ValueOf(intval))
		} else {
			vf.Set(reflect.ValueOf(int(intval)))
		}
	case reflect.Bool:
		if stringValue == "true" || stringValue == "yes" || stringValue == "on" || stringValue == "1" {
			vf.SetBool(true)
		} else {
			vf.SetBool(false)
		}
	case reflect.Slice:
		switch svk := vf.Type().Elem().Kind(); svk {
		case reflect.Int:
			var intSlice []int
			for _, si := range strings.Split(stringValue, ",") {
				var intval int64
				if intval, err = strconv.ParseInt(si, 10, 64); err != nil {
					return
				}
				intSlice = append(intSlice, int(intval))
			}
			vf.Set(reflect.ValueOf(intSlice))
			//TODO more cases with another types
		}
	case reflect.Float64:
		var fl64 float64
		fl64, err = strconv.ParseFloat(stringValue, 64)
		if err != nil {
			return
		}
		vf.SetFloat(fl64)

	default:
		err = fmt.Errorf("parse: %s: unsupported field type: %s", tagVal, vkind.String())
		return
	}
	return
}

func split(src []byte, separator byte) (dest [fieldLen][]byte, err error) {
	var cnt int

	for p := 0; p < len(src); cnt++ {
		var col []byte
		if idx := bytes.IndexByte(src[p:], separator); idx == -1 {
			col = src[p:]
			p = len(src)
		} else {
			col = src[p : p+idx]
			p += idx + 1
		}

		if cnt < len(dest) {
			dest[cnt] = col
		}
	}
	return
}
