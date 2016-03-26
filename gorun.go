// Copyright 2013 The Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.
//
// Test Run: go run gorun.go
// INSTALL: go install
//
// @Author: wangxian
// @Created: 2013-07-31
//
// Project URL: https://github.com/wangxian/gorun
package main

import (
	"flag"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	// "code.google.com/p/go.exp/fsnotify"
)

var (
	cmd     *exec.Cmd
	state   sync.Mutex
	appname string

	// fixed: File change soon
	eventTime = make(map[string]time.Time)
)

func Start() {
	if appname != "" {
		log.Println("Start", appname, "...")

		cmd = exec.Command("go", "run", appname)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		go cmd.Run()
	} else {
		Rebuild()
		log.Println("Start project ...")

		curpath, _ := os.Getwd()
		// log.Println(curpath)

		cmd = exec.Command("./" + filepath.Base(curpath))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		go cmd.Run()
	}
}

func Rebuild() {
	state.Lock()
	defer state.Unlock()

	log.Println("Start rebuild project...")

	bcmd := exec.Command("go", "build")
	bcmd.Stdout = os.Stdout
	bcmd.Stderr = os.Stderr
	err := bcmd.Run()

	if err != nil {
		log.Println("============== Rebuild project failed ==============")
		return
	}
	log.Println("Rebuild project success ...")
}

func Stop() {
	defer func() {
		if e := recover(); e != nil {
			log.Println("Kill process error:", e)
		}
	}()
	if cmd != nil {
		log.Println("Kill running process:", cmd.Process.Pid)
		cmd.Process.Kill()
		// // cmd.Process.Release()

		// st, _ := cmd.Process.Wait()
		// log.Println("Kill process:", st.Exited())
	}
}

func Restart() {
	Stop()
	Start()
}

func Watch() {
	path, _ := os.Getwd()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// walk dirs
	walkFn := func(path string, info os.FileInfo, err error) error {

		// Not watch dir, has .
		if info.IsDir() && !strings.Contains(path, ".") {
			log.Println("Watch DIR:", path)

			err = watcher.Add(path)
			if err != nil {
				log.Fatal(err)
			}
		}
		return nil
	}

	if err := filepath.Walk(path, walkFn); err != nil {
		log.Println(err)
	}

	changed := false
	for {
		select {
		case event := <-watcher.Events:

			changed = true
			if t, ok := eventTime[event.Name]; ok {
				if t.Add(time.Millisecond * 3000).After(time.Now()) {
					changed = false
				}
			}
			eventTime[event.Name] = time.Now()

			log.Println("event", event, "changed=", changed)

			if event.Op == fsnotify.Write && strings.Contains(event.Name, ".go") {
				println("---------------------------------------------------")
				Restart()
			}

		case err := <-watcher.Errors:
			log.Println("error:", err)
		}
	}
}

func main() {
	flag.Parse()
	appname = flag.Arg(0)
	Start()

	Watch()
}
