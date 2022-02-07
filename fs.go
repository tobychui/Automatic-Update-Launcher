package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	cp "github.com/otiai10/copy"
)

func restoreOldBackup() {
	fmt.Println("[aulauncher] Application unable to launch. Restoring from backup")
	if fileExists("app.old") {
		backupfiles, err := filepath.Glob("app.old/*")
		if err != nil {
			fmt.Println("[aulauncher] Unable to restore backup. Exiting.")
			os.Exit(1)
		}

		for _, thisBackupFile := range backupfiles {
			if isDir(thisBackupFile) {
				cp.Copy(thisBackupFile, "./"+filepath.Base(thisBackupFile))
			} else {
				copy(thisBackupFile, "./"+filepath.Base(thisBackupFile))
			}
		}

		os.RemoveAll("app.old")
	} else {
		fmt.Println("[aulauncher] Application backup not found. Exiting.")
		os.Exit(1)
	}

}

//Auto detect and execute the correct binary
func autoDetectExecutable(searchPattern string) (string, error) {
	files, err := filepath.Glob(searchPattern)
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", errors.New("start file not found")
	}

	return files[0], nil
}

func updateIfExists() {
	if fileExists("./updates") {
		//Backup all things to app.old
		os.MkdirAll("./app.old", 0775)
		files := []string{}
		if len(launchConfig.Backup) == 0 {
			//Backup everything except launcher files
			files, _ = filepath.Glob("./*")
		} else {
			//Backup the files required by config
			for _, wildcard := range launchConfig.Backup {
				matchings, _ := filepath.Glob(wildcard)
				for _, matching := range matchings {
					if !contains(files, matching) {
						files = append(files, matching)
					}
				}
			}
		}

		skippingList := []string{filepath.Base(os.Args[0]), "app.old", "updates"}
		for _, file := range files {
			thisFilename := filepath.Base(file)
			if !contains(skippingList, thisFilename) {
				if isDir(file) {
					cp.Copy(file, filepath.Join("app.old", filepath.Base(file)))
				} else {
					copy(file, filepath.Join("app.old", filepath.Base(file)))
				}
			}
		}
		//Move all things from update to current path
		if fileExists(filepath.Join("./updates", filepath.Base(os.Args[0]))) {
			//Now allow updating launcher itself during runtime
			os.Rename(filepath.Join("./updates", filepath.Base(os.Args[0])), "_"+filepath.Join("./updates", filepath.Base(os.Args[0])))
		}
		cp.Copy("./updates", "./")
		os.RemoveAll("./updates")
	}
}

func contains(slice []string, find string) bool {
	for _, a := range slice {
		if a == find {
			return true
		}
	}
	return false
}

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, errors.New("invalid file")
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func isDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fileInfo.IsDir()
}

func fileExists(name string) bool {
	_, err := os.Stat(name)
	if err == nil {
		return true
	}
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return false
}
