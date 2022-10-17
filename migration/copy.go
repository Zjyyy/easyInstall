package migration

import (
	"archive/zip"
	"easyInstall/conf"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
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
		} else {
			self.copyDirectory(item.Source, item.Target)
		}
	}
	return nil
}

// 验证配置文件中的路径是否合法有效
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
func (self fileMigration) zipDir(dirPath string, zipFilePath string) {
	zipFile := filepath.Join(zipFilePath, "/", filepath.Base(zipFilePath)+".zip")
	log.Println(">>>>> " + zipFile)

	fz, err := os.Create(zipFile)
	if err != nil {
		log.Fatalf("创建zip文件失败：%v \n", err.Error())
	}
	defer fz.Close()

	w := zip.NewWriter(fz)
	defer w.Close()

	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			fDest, err := w.Create(path[len(dirPath)+1:])
			if err != nil {
				log.Fatalf("创建zip目标文件失败: %v", err.Error())
				return nil
			}

			fSrc, err := os.Open(path)
			if err != nil {
				log.Fatalf("创建Zip源头文件失败: %v", err.Error())
				return nil
			}
			defer fSrc.Close()
			_, err = io.Copy(fDest, fSrc)
			if err != nil {
				log.Fatalf("Zip复制失败: %v", err.Error())
				return nil
			}
		}
		return nil
	})
}
func (self fileMigration) unzipDir(zipFile string, dirPath string) {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		log.Fatalf("Open zip file failed: %s\n", err.Error())
	}
	defer r.Close()

	for _, f := range r.File {
		func() {
			path := dirPath + string(filepath.Separator) + f.Name
			os.MkdirAll(filepath.Dir(path), 0755)
			fDest, err := os.Create(path)
			if err != nil {
				log.Printf("Create failed: %s\n", err.Error())
				return
			}
			defer fDest.Close()

			fSrc, err := f.Open()
			if err != nil {
				log.Printf("Open failed: %s\n", err.Error())
				return
			}
			defer fSrc.Close()

			_, err = io.Copy(fDest, fSrc)
			if err != nil {
				log.Printf("Copy failed: %s\n", err.Error())
				return
			}
		}()
	}
}
func (self fileMigration) copyDirectory(srcDirectoryPath string, destDirectoryPath string) error {
	filepath.Walk(srcDirectoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf("遍历文件目录错误: %v", err)
			return nil
		}
		if !info.IsDir() {
			destPath := strings.Replace(path, filepath.Clean(srcDirectoryPath), filepath.Clean(destDirectoryPath), -1)
			self.copyFile(path, destPath)
		}
		return nil
	})
	return nil
}

// @description 分段拷贝文件
// @param srcFilePath "源头文件的路径"
// @param destFilePath "目标文件的路径(具体到文件名)"
func (self fileMigration) copyFile(srcFilePath string, destFilePath string) error {
	targetPath := filepath.Dir(destFilePath)
	// 如果文件的路径不存在，创建路径
	if !self.ExistFileOrDirectory(targetPath) {
		os.MkdirAll(targetPath, os.ModePerm)
	}

	srcFile, err := os.Open(srcFilePath)
	if err != nil {
		log.Fatalf("读取文件错误：%v \n", err.Error())
		return err
	}
	defer srcFile.Close() // close after checking err

	// 创建目标文件，稍后会向这个目标文件写入拷贝内容
	distFile, err := os.Create(destFilePath)
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

// @description 判断文件或者目录是否存在
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

// @description 判断是否是目录
func (self fileMigration) IsDirectory(directoryPath string) bool {
	s, err := os.Stat(directoryPath)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// @description 判断是否是文件
func (self fileMigration) IsFile(filePath string) bool {
	return !self.IsDirectory(filePath)
}
