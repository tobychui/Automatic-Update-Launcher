package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"
)

/*
	Automatic Update Launcher
	For auto update and future extension purpose

	Author: tobychui

*/

type Config struct {
	Version   string   `json:"version"`
	Start     string   `json:"start"`
	Backup    []string `json:"backup"`
	MaxRetry  int      `json:"max_retry"`
	RespPort  int      `json: "resp_port"`
	CrashTime int      `json:"crash_time"`
	Verbal    bool     `json:"verbal"`
}

var (
	launchConfig Config
	norestart    bool = false
)

func main() {
	//Grab the config from json
	configs, err := ioutil.ReadFile("launcher.json")
	if err != nil {
		log.Fatal("Unable to read launcher.json")
		panic(err)
	}

	err = json.Unmarshal(configs, &launchConfig)
	if err != nil {
		log.Fatal("launcher.json parse failed")
		panic(err)
	}

	//Print basic information
	binaryName, err := autoDetectExecutable(launchConfig.Start)
	if err != nil {
		log.Fatal("unable to get start target file")
		panic(err)
	}

	if launchConfig.Verbal {
		fmt.Println("[aulauncher] Trying to start " + binaryName)
	}

	//Check if updates exists. If yes, overwrite it
	updateIfExists()

	//Check launch paramter for norestart
	for _, arg := range os.Args[1:] {
		if arg == "-h" || arg == "-help" {
			//help argument, do not restart
			norestart = true
		} else if arg == "-version" || arg == "-v" {
			//version argument, no restart
			norestart = true
		}
	}

	//Register the binary start path
	cmd := exec.Command(binaryName, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	//Register the http server to notify the services there is a launcher will handle the update
	go func() {
		http.HandleFunc("/chk", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Launcher v" + launchConfig.Version))
			if launchConfig.Verbal {
				fmt.Println("[aulauncher] Launcher check request received")
			}
		})

		http.ListenAndServe("127.0.0.1:"+strconv.Itoa(launchConfig.RespPort), nil)
	}()

	retryCounter := 0
	//Start the cmd
	for {
		startTime := time.Now().Unix()
		cmd.Run()
		endTime := time.Now().Unix()

		if norestart {
			return
		}
		if endTime-startTime < int64(launchConfig.CrashTime) {
			//Less than 3 seconds, shd be crashed. Add to retry counter
			retryCounter++
			if launchConfig.Verbal {
				fmt.Println("[aulauncher] Application crashed. Restarting in 3 seconds")
			}
		} else {
			if launchConfig.Verbal {
				fmt.Println("[aulauncher] Application exited. Restarting in 3 seconds")
			}
			retryCounter = 0
		}

		time.Sleep(3 * time.Second)

		if retryCounter > launchConfig.MaxRetry+1 {
			//Fail to start. Exit program
			log.Fatal("Unable to start application. Exiting to OS")
			return
		} else if retryCounter > launchConfig.MaxRetry {
			//Restore from old version of the binary
			if launchConfig.Verbal {
				fmt.Println("[aulauncher] Restoring old version of application")
			}
			restoreOldBackup()
		} else {
			updateIfExists()
		}

		//Rebuild the start paramters
		cmd = exec.Command(binaryName, os.Args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

}
