package main

import (
	"easyInstall/conf"
	"easyInstall/migration"
	"fmt"
)

func main() {
	config, userConfig, err := conf.LoadConfigAndUserConfig("config.ini", "userConfig.yml")
	if err != nil {
		fmt.Printf("加载配置文件异常: %v \n", err)
		return
	}
	// fmt.Printf("Config: %v \n UserConfig: %v \n", config, userConfig)
	fileMigration := migration.InitializeFileMigration(&config, &userConfig)
	fmt.Printf("%v \n", fileMigration)
	err = fileMigration.CopyMigration()
	if err != nil {
		fmt.Printf("拷贝文件错误:%v \n", err)
	}
}
