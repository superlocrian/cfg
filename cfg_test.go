package cfg

import (
	"flag"
	"github.com/stretchr/testify/require"
	"os"
	"reflect"
	"testing"
)

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

func TestSplit(t *testing.T) {
	str := "G=12"

	dst, err := split([]byte(str), '=')
	require.NoError(t, err)
	require.Equal(t, []byte("G"), dst[0])
	require.Equal(t, []byte("12"), dst[1])
}

func TestParseCmdFlag(t *testing.T) {
	n, u := parseCmdFlagString("intval,i |integer value ")
	require.Equal(t, 2, len(n))
	require.Equal(t, "intval", n[0])
	require.Equal(t, "i", n[1])
	require.Equal(t, u, "integer value")

	n, u = parseCmdFlagString(" ")
	require.Equal(t, 0, len(n))

}
