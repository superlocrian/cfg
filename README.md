# CFG

A simple configurator that parses go-tags and fills the container passed to it with the values of flags and environment variables.
The flags values overwrite the default values, contained in the transferred container. Environment variables overwrite default values and flags values.
It is useful when your application has to use different configuration sources.

Go tag format for commandline flags:
```golang
type YourConfig struct {
	Field   int    `cmd:"flagname|flag description string"`
}
```
Go tag format for environment variables:
```golang
type YourConfig struct {
	Field   int    `env:"VARNAME_WITHOUT_PREFIX"`
}
```

### limitations:

* So far, there is no way to set both normal and shortened flag names for the same field.
* It is not possible to set up default values for commandline flags using go tags (but it would be cool to add)


