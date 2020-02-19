package cfg

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func Unmarshal(container interface{}, envPrefix string, flagSet *flag.FlagSet, args []string) (err error) {
	if err = UnmarshalFromFlags(container, flagSet, args); err != nil {
		return
	}
	if err = UnmarshalFromEnvironment(container, envPrefix); err != nil {
		return
	}
	return
}

func UnmarshalFromEnvironment(container interface{}, envPrefix string) (err error) {
	return getEnv(envPrefix, reflect.ValueOf(container).Elem())
}

//args: command line arguments which should not include the command name
func UnmarshalFromFlags(container interface{}, flagSet *flag.FlagSet, args []string) (err error) {
	if err = initCmdFlags(reflect.ValueOf(container).Elem(), flagSet); err != nil {
		return
	}
	if err = flagSet.Parse(args); err != nil {
		return
	}
	if err = getFlags(reflect.ValueOf(container).Elem(), flagSet); err != nil {
		return
	}
	return
}

const envSep = "="
const envTag = "env"
const cmdRag = "cmd"

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
		if name == "" {
			continue
		}
		switch vkind := vf.Kind(); vkind {
		case reflect.String:
			fs.String(name, vf.String(), usage)
		case reflect.Int, reflect.Int64:
			if vkind == reflect.Int64 {
				fs.Int64(name, vf.Int(), usage)
			} else {
				fs.Int(name, int(vf.Int()), usage)
			}
		case reflect.Bool:
			fs.Bool(name, vf.Bool(), usage)
		case reflect.Slice:
			switch svk := vf.Type().Elem().Kind(); svk {
			case reflect.Int:
				isv := intSliceValue(vf.Interface().([]int))
				fs.Var(&isv, name, usage)
			//TODO more cases with another types
			default:
				err = fmt.Errorf("initCmdFlags:%s: unsupported slice type: %v", name, svk)
				return
			}
		case reflect.Float64:
			fs.Float64(name, vf.Float(), usage)
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
			if name == "" {
				continue
			}
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
			if getter.Get() == nil {
				continue
			}
			if str, ok := getter.Get().(string); ok {
				vf.Set(reflect.ValueOf(strings.Trim(str, " \t\n\r'\"`")))
			} else {
				vf.Set(reflect.ValueOf(getter.Get()))
			}
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
			tagVal := strings.Trim(tf.Tag.Get(envTag), " \n\t\r")
			if tagVal == "" {
				continue
			}
			varname := pref + tagVal
			for _, e := range os.Environ() {
				pair := strings.SplitN(e, envSep, 2)
				if varname == pair[0] && len(pair[1]) > 0 {
					stringValue := strings.ToLower(strings.Trim(pair[1], " \n\t\r"))
					if err = parse(vf, stringValue, tagVal); err != nil {
						return
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
		vf.SetBool(false)
		if stringValue == "true" || stringValue == "yes" || stringValue == "on" || stringValue == "1" {
			vf.SetBool(true)
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
