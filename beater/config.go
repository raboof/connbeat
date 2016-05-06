package beater

type ConnConfig struct {
}

type ConfigSettings struct {
	Input    *ConnConfig `config:"input"`
	Connbeat *ConnConfig `config:"connbeat"`
}
