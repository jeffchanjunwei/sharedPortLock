package sharedPortLock

import (
	"bufio"
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"syscall"
)

func main() {

	beginPort := os.Args[0]
	engPort := os.Args[1]

	const (
		dubboQosSharedPortFile = "dubboQosSharedPorts"
	)

	homePath, err := Home()
	existed, err := PathExists(homePath + "/" + dubboQosSharedPortFile)
	if err != nil {
		panic("dubbo qos shared ports file detect failed")
	}
	if existed == false {
		_, err := os.Create(homePath + "/" + dubboQosSharedPortFile)
		if err != nil {
			panic("dubbo qos shared ports file create failed")
		}
	}

	var f *os.File
	f, err = os.Open(homePath + "/" + dubboQosSharedPortFile)
	if err != nil {
		panic("dubbo qos shared ports file open failed")
	}

	defer f.Close()
	// 阻塞模式下，加排它锁
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		panic("dubbo qos shared ports file lock failed")
	}
	// 这里进行业务逻辑
	// TODO
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		scanner.Text()
	}

	// 解锁
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_UN); err != nil {
		panic("dubbo qos shared ports file unlock failed")
	}

	return
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	//isnotexist来判断，是不是不存在的错误
	if os.IsNotExist(err) { //如果返回的错误类型使用os.isNotExist()判断为true，说明文件或者文件夹不存在
		return false, nil
	}
	return false, err //如果有错误了，但是不是不存在的错误，所以把这个错误原封不动的返回
}

func Home() (string, error) {
	user, err := user.Current()
	if nil == err {
		return user.HomeDir, nil
	}

	// cross compile support

	if "windows" == runtime.GOOS {
		return homeWindows()
	}

	// Unix-like system, so just assume Unix
	return homeUnix()
}

func homeUnix() (string, error) {
	// First prefer the HOME environmental variable
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	// If that fails, try the shell
	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", errors.New("blank output when reading home directory")
	}

	return result, nil
}

func homeWindows() (string, error) {
	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	if drive == "" || path == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		return "", errors.New("HOMEDRIVE, HOMEPATH, and USERPROFILE are blank")
	}

	return home, nil
}
