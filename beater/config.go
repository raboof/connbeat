package beater

type ConnConfig struct {
	ExposeCmdline bool `config:"expose_cmdline"`
	ExposeEnviron bool `config:"expose_environ"`
}

type ConfigSettings struct {
	Connbeat ConnConfig `config:"connbeat"`
}

func (c *ConnConfig) Validate() error {
	return nil
}

var (
	defaultConfig = ConnConfig{
		ExposeCmdline: true,
		ExposeEnviron: false,
	}
)
