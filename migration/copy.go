package migration

import (
	"easyInstall/conf"
	"errors"
	"io"
	"log"
	"os"
	"strings"
)

type FileMigration struct {
	*fileMigration
}
type fileMigration struct {
	enableWindows      bool
	enableLinux        bool
	setupFileDirectory string
	linuxDefaultPath   string
	windowsDefaultPath string
	filesPath          []conf.DirectoryConf
}

func InitializeFileMigration(config *conf.Config, userConfig *conf.UserConfig) *FileMigration {
	return &FileMigration{
		&fileMigration{
			enableWindows:      config.System.WindowsEnable,
			enableLinux:        config.System.LinuxEnable,
			setupFileDirectory: config.Common.SetupFileDirectory,
			windowsDefaultPath: config.System.WindowsDefaultPath,
			linuxDefaultPath:   config.System.LinuxDefaultPath,
			filesPath:          userConfig.Directory, // TODO: 需要改成深拷贝
		},
	}
}

// TODO: 拷贝目录的部分还没实现
func (self *fileMigration) CopyMigration() error {
	if len(self.filesPath) <= 0 {
		return errors.New("文件路径配置为空")
	}
	for _, item := range self.filesPath {
		if !self.validityPathConfig(&item) {
			return errors.New("配置文件中路径校验不合法")
		}
		if !item.IsDirectory {
			self.copyFile(item.Source, item.Target+"/"+item.FileName)
		}
	}
	return nil
}
func (self fileMigration) validityPathConfig(conf *conf.DirectoryConf) bool {
	isSetupFileExist := self.ExistFileOrDirectory(conf.Source)
	if !isSetupFileExist {
		log.Fatalf("安装源文件或目录不存在：%v \n", conf.Source)
		return false
	}
	isTargetDirectortExist := self.ExistFileOrDirectory(conf.Target)
	if !isTargetDirectortExist {
		os.MkdirAll(conf.Target, 0777)
		os.Chmod(conf.Target, 0777)
		// 创建目录后，再次校验目标目录是否存在
		if !self.ExistFileOrDirectory(conf.Target) {
			log.Fatalf("目标目录不存在: %v \n", conf.Target)
			return false
		}
	}

	// 校验配置文件中文件名和目录中文件名是否一样
	if !conf.IsDirectory {
		srcFileSplit := strings.Split(conf.Source, "/")
		if conf.FileName != srcFileSplit[len(srcFileSplit)-1] {
			log.Fatalf("配置文件中文件名和路径中文件名不一致 文件名: %v 路径: %v \n",
				conf.FileName, conf.Source)
			return false
		}
	}
	return true
}
func (self fileMigration) copyDirectory(srcDirectoryPath string, destDirectoryPath string) error {
	return nil
}
func (self fileMigration) copyFile(srcFilePath string, destPath string) error {
	srcFile, err := os.Open(srcFilePath)
	if err != nil {
		log.Fatalf("读取文件错误：%v \n", err.Error())
		return err
	}
	defer srcFile.Close() // close after checking err

	// 创建目标文件，稍后会向这个目标文件写入拷贝内容
	distFile, err := os.Create(destPath)
	if err != nil {
		log.Fatalf("目标文件创建失败，原因是: %v \n", err)
	}
	defer distFile.Close()

	srcFileStat, err := srcFile.Stat()
	if err != nil {
		log.Fatalf("获取原始文件状态失败: %v \n", err)
	}
	srcFileSize := srcFileStat.Size()
	log.Printf("srcFileSize: %v \n", srcFileSize)

	var tmp = make([]byte, 1024*4)
	for {
		n, err := srcFile.Read(tmp)
		n, _ = distFile.Write(tmp[:n])
		if err != nil {
			// 读到了文件末尾，并且写入完毕，任务完成返回
			if err == io.EOF {
				break
			} else {
				log.Fatalf("拷贝过程中发生错误，错误原因是: %v \n", err)
			}
		}
	}
	return nil
}

func (self fileMigration) ExistFileOrDirectory(filePath string) bool {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func (self fileMigration) IsDirectory(directoryPath string) bool {
	s, err := os.Stat(directoryPath)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func (self fileMigration) IsFile(filePath string) bool {
	return !self.IsDirectory(filePath)
}
