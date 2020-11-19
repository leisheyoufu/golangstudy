package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func timespecToTime(ts syscall.Timespec) time.Time {
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}

func GetFileCreateTime(path string) {
	finfo, _ := os.Stat(path)
	// Sys() returns interface{}, so type assertion is required. Different platforms require different types. *syscall.Stat_t on linux
	stat_t := finfo.Sys().(*syscall.Stat_t)
	fmt.Println(stat_t)
	// atime, ctime, and mtime are access time, creation time, and modification time respectively. For details, see man 2 stat.
	// fmt.Println(timespecToTime(stat_t.Atim))
	// fmt.Println(timespecToTime(stat_t.Ctim))
	// fmt.Println(timespecToTime(stat_t.Mtim))
}

func DelFiles(path string, interval time.Duration) error {
	fileInfoList, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	now := time.Now()
	for _, file := range fileInfoList {
		if now.Sub(file.ModTime()) > interval {
			fmt.Println(file.Name())
			os.Remove(filepath.Join(path, file.Name()))
		}
	}
	return nil
}

func BackupFile(dir, prefix string, b []byte) error {
	t := time.Now()
	s := prefix + "_" + strings.ReplaceAll(t.Format("2006-01-02 15:04:05"), " ", "_") + ".json"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
	}
	filepath := filepath.Join(dir, s)
	err := ioutil.WriteFile(filepath, b, 0644)
	if err != nil {
		return err
	}
	fmt.Println("Backup file successfully")
	return nil
}

func main() {
	BackupFile("/tmp/backup", "backup", []byte("backup"))
	//DelFiles("/tmp/backup", time.Hour*1)
}
