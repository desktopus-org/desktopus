package config

// AppConfig is the application-level configuration (~/.desktopus/config.yaml)
type AppConfig struct {
	Docker DockerConfig `yaml:"docker"`
	Server ServerConfig `yaml:"server"`
	Build  BuildConfig  `yaml:"build"`
	Log    LogConfig    `yaml:"log"`
	Store  StoreConfig  `yaml:"store"`
}

type DockerConfig struct {
	Host string `yaml:"host"`
}

type ServerConfig struct {
	Listen string `yaml:"listen"`
	Port   int    `yaml:"port"`
}

type BuildConfig struct {
	CacheDir         string `yaml:"cache_dir"`
	Parallel         int    `yaml:"parallel"`
	AnsibleVerbosity int    `yaml:"ansible_verbosity"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type StoreConfig struct {
	Path string `yaml:"path"`
}

// DefaultAppConfig returns an AppConfig with sensible defaults
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		Docker: DockerConfig{
			Host: "unix:///var/run/docker.sock",
		},
		Server: ServerConfig{
			Listen: "127.0.0.1",
			Port:   7575,
		},
		Build: BuildConfig{
			CacheDir:         "~/.desktopus/cache",
			Parallel:         2,
			AnsibleVerbosity: 0,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "text",
		},
		Store: StoreConfig{
			Path: "~/.desktopus/desktopus.db",
		},
	}
}
