package main

import (
	"flag"
	fn "github.com/howeyc/fsnotify"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	runningApp *exec.Cmd
	appName    string
	// 由于 fsnotify 的文件产生的事件与期望的不一样, 所以只能使用 time 来确定
	modifyUnixTimes = make(map[string]int64)
)

func main() {
	args := flag.Args()

	if len(args) == 1 {
		appName = args[0]
	} else if len(args) > 1 {
		log.Fatalln("只允许一个参数")
	} else {
		abs, err := filepath.Abs("./")
		if err != nil {
			log.Fatalln(err)
		}
		appName = filepath.Base(abs)
	}

	paths, err := Walk("./")
	if err != nil {
		log.Fatalln(err)
	}

	Build()
	go Start()
	Watch(paths)
}

func Walk(rootDir string) (paths []string, err error) {
	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() || strings.Contains(path, ".git") {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return
	}
	return
}

func Watch(paths []string) {
	watcher, err := fn.NewWatcher()
	if err != nil {
		log.Fatalln(err)
	}

	done := make(chan bool)

	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				if filepath.Ext(ev.Name) == ".go" {
					reBuild := false
					t, ok := modifyUnixTimes[ev.Name]
					if !ok {
						modifyUnixTimes[ev.Name] = time.Now().Unix()
						reBuild = true
					} else {
						nt := time.Now().Unix()
						reBuild = (nt - t) > 2
						modifyUnixTimes[ev.Name] = nt
					}
					if reBuild {
						Rebuild()
					}
				}
			case err := <-watcher.Error:
				log.Println("error:", err)
			}
		}
	}()

	for _, path := range paths {
		err = watcher.Watch(path)
		if err != nil {
			log.Fatalln(err)
		}
	}

	log.Println("Begin to watch app:", appName)
	<-done
	watcher.Close()
}

func Build() (err error) {
	begin := time.Now().UnixNano()
	cmd := exec.Command("go", "build")
	// 将执行的错误和输出都显示到当前的 标准输出, 标准错误 设备中
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run() // Wait for build
	log.Println("Build passed:", (time.Now().UnixNano()-begin)/1000/1000, "ms")
	return
}

// Golang 的应用使用 build, 然后再 run 避免直接使用 run 会出现文件缺少引入的问题
func Rebuild() {
	err := Build()
	if err != nil {
		log.Println(err)
	} else {
		ReStart()
	}
}

func ReStart() {
	if runningApp != nil {
		log.Println("Kill old running app:", appName)
		runningApp.Process.Kill()
	}
	Start()
}

func Start() {
	runningApp = exec.Command("./" + appName)
	runningApp.Stdout = os.Stdout
	runningApp.Stderr = os.Stderr
	log.Println("Start running app:", appName)
	go runningApp.Run()
}
