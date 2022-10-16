package conf

import (
	"github.com/go-ini/ini"
	"github.com/spf13/viper"
)

func LoadConfigAndUserConfig(
	configPath string,
	userConfigPath string) (Config, UserConfig, error) {
	errConfig := new(Config)
	errUserConfig := new(UserConfig)
	config, err := LoadConfigFromInI(configPath)
	if err != nil {
		return *errConfig, *errUserConfig, err
	}
	userConfig, err := LoadUserConfigFromYAML(userConfigPath)
	if err != nil {
		return *errConfig, *errUserConfig, err
	}
	return config, userConfig, nil
}

type CommonConf struct {
	Languages          string `ini:"languages"`
	SetupFileDirectory string `ini:"setupFileDirectory"`
	UserConfs          string `ini:"userConfs"`
}
type SystemConf struct {
	WindowsEnable      bool   `ini:"windows.enable"`
	WindowsDefaultPath string `ini:"windows.defaultPath"`
	LinuxEnable        bool   `ini:"linux.enable"`
	LinuxDefaultPath   string `ini:"linux.defaultPath"`
}
type TerminalConf struct {
	ConsoleEnable bool `ini:"console.enable"`
	ServerEnable  bool `ini:"server.enable"`
	ServerPort    int  `ini:"server.port"`
}
type Config struct {
	Common   CommonConf   `ini:"Common"`
	System   SystemConf   `ini:"System"`
	Terminal TerminalConf `ini:"Terminal"`
}

func LoadConfigFromInI(fileName string) (Config, error) {
	CONFIG := new(Config)
	err := ini.MapTo(CONFIG, fileName)
	if err != nil {
		return *CONFIG, nil
	}
	return *CONFIG, err
}

type UserDefinedVariableInfo struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}
type EnvironmentConf struct {
	Windows struct {
		Path        []string                  `yaml:"path"`
		UserDefined []UserDefinedVariableInfo `yaml:"userDefined"`
	}
}
type DirectoryConf struct {
	IsDirectory bool   `yaml:"isDirectory"`
	FileName    string `yaml:"fileName"`
	Source      string `yaml:"source"`
	Target      string `yaml:"target"`
}
type UserConfig struct {
	Environment EnvironmentConf `yaml:"environment"`
	Directory   []DirectoryConf `yaml:"directory"`
}

func LoadUserConfigFromYAML(fileName string) (UserConfig, error) {
	CONFIG := new(UserConfig)
	viper.SetConfigFile("yaml")
	viper.SetConfigFile(fileName)
	err := viper.ReadInConfig()
	if err != nil {
		return *CONFIG, err
	}
	err = viper.Unmarshal(&CONFIG)
	if err != nil {
		return *CONFIG, err
	}
	return *CONFIG, nil
}
