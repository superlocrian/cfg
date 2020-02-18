package cfg

import (
	"flag"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
	"testing"
)

func TestAll(t *testing.T) {
	type cc struct {
		IntVal   int    `yaml:"intval" cmd:"intval|integer value" env:"INTVAL"`
		Int64    int64  `yaml:"int64"cmd:"int64|integer 64 value " env:"INT64"`
		StrVal   string `yaml:"strval" cmd:"strval" env:"STRVAL"`
		IntSlice []int  `yaml:"int-slice" cmd:"intslice" env:"INTSLICE"`
		Data     struct {
			Fl64    float64 `yaml:"fl-64" cmd:"float|how to use float" env:"FL64"`
			BoolVal bool    `yaml:"bool-val" cmd:"bool|boolean value" env:"BOOL"`
		}
	}

	cfg := &cc{}
	fset := flag.NewFlagSet("test", flag.ContinueOnError)
	err := initCmdFlags(reflect.ValueOf(cfg).Elem(), fset)
	require.NoError(t, err)
	err = fset.Parse([]string{
		"-intval", "11",
		"-int64", "123",
		"--strval=\"some string from cmd\"",
		"-float", "12.4",
		"-intslice", "4,5,6"})
	require.NoError(t, err)

	_ = os.Setenv("P_INTVAL", "12")
	_ = os.Setenv("P_INT64", "90009")
	_ = os.Setenv("P_FL64", "123.4")
	_ = os.Setenv("P_INTSLICE", "7,8,9")
	_ = os.Setenv("P_STRVAL", "some env string")

	yamlData := []byte(`
intval: 10
int64: 11
strval: some string
int-slice: 
  - 1
  - 2
  - 3
data:
  fl-64: 12.5
  bool-val: yes
`)

	//start with yaml
	err = yaml.Unmarshal(yamlData, &cfg)
	require.NoError(t, err)
	require.Equal(t, 10, cfg.IntVal)
	require.Equal(t, int64(11), cfg.Int64)
	require.Equal(t, "some string", cfg.StrVal)
	require.Equal(t, []int{1, 2, 3}, cfg.IntSlice)
	require.Equal(t, 12.5, cfg.Data.Fl64)
	require.Equal(t, true, cfg.Data.BoolVal)

	//then rewrite with flags
	err = getFlags(reflect.ValueOf(cfg).Elem(), fset)
	require.NoError(t, err)
	require.Equal(t, 11, cfg.IntVal)
	require.Equal(t, int64(123), cfg.Int64)
	require.Equal(t, "some string from cmd", cfg.StrVal)
	require.Equal(t, []int{4, 5, 6}, cfg.IntSlice)
	require.Equal(t, 12.4, cfg.Data.Fl64)

	//then rewrite with env vars
	err = getEnv("P_", reflect.ValueOf(cfg).Elem())
	require.NoError(t, err)
	require.NoError(t, err)
	require.Equal(t, 12, cfg.IntVal)
	require.Equal(t, int64(90009), cfg.Int64)
	require.Equal(t, []int{7, 8, 9}, cfg.IntSlice)
	require.Equal(t, "some env string", cfg.StrVal)
}

func TestGenEnv(t *testing.T) {

	type cc struct {
		IntVal   int    `env:"IntVal"`
		StrVal   string `env:"StrVal"`
		IntSlice []int  `env:"IntSlice"`
	}

	cf := &cc{}

	_ = os.Setenv("IntVal", "11")
	_ = os.Setenv("StrVal", "11")
	_ = os.Setenv("IntSlice", "11,12,13")

	err := getEnv("", reflect.ValueOf(cf).Elem())
	require.NoError(t, err)
	require.Equal(t, 11, cf.IntVal)
	require.Equal(t, "11", cf.StrVal)
	require.Equal(t, []int{11, 12, 13}, cf.IntSlice)

}

func TestCmdArgs(t *testing.T) {

	type cc struct {
		IntVal   int    `cmd:"intval|integer value "`
		Int64    int64  `cmd:"int64|integer 64 value "`
		StrVal   string `cmd:"strval"`
		IntSlice []int  `cmd:"intslice"`
		Data     struct {
			Fl64    float64 `cmd:"float|how to use float"`
			BoolVal bool    `cmd:"bool|boolean value"`
		}
	}

	cf := &cc{}
	fset := flag.NewFlagSet("test", flag.ContinueOnError)
	err := initCmdFlags(reflect.ValueOf(cf).Elem(), fset)
	require.NoError(t, err)
	err = fset.Parse([]string{"-intval", "11", "-int64", "123", "--strval=12", "-float", "12.4", "-intslice", "1,2,3", "-bool", "true"})
	require.NoError(t, err)

	err = getFlags(reflect.ValueOf(cf).Elem(), fset)
	require.NoError(t, err)
	require.Equal(t, 11, cf.IntVal)
	require.Equal(t, int64(123), cf.Int64)
	require.Equal(t, "12", cf.StrVal)
	require.Equal(t, []int{1, 2, 3}, cf.IntSlice)
	require.Equal(t, 12.4, cf.Data.Fl64)
	require.Equal(t, true, cf.Data.BoolVal)

}

func TestParseCmdFlag(t *testing.T) {
	n, u := parseCmdFlagString("intval|integer value ")
	require.Equal(t, "intval", n)
	require.Equal(t, u, "integer value")

	n, u = parseCmdFlagString(" ")
	require.Equal(t, 0, len(n))

}
