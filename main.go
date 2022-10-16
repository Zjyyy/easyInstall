package main

import "fmt"

func main() {
	config, userConfig, err := LoadConfigAndUserConfig("config.ini", "userConfig.yml")
	if err != nil {
		fmt.Printf("加载配置文件异常: %v \n", err)
		return
	}
	fmt.Printf("Config: %v \n UserConfig: %v \n", config, userConfig)
}
