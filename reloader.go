package reloader

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/guoruibiao/commands"
)

// Reloader auto reload programs by checking file modified or not based on EPOOL
type Reloader struct {
	files    []string          // files in `PWD` with recursion way
	Watcher  *fsnotify.Watcher // based on epool of system supported
	Duration int               // duration to check if files whether be modified, seconds
	Command  []string          // actually command to be run, such as `go run main.go`
}

// NewReloader get the instance of Reloader
func NewReloader(commands []string, duration int) (*Reloader, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	reloader := &Reloader{
		files:    nil,
		Watcher:  watcher,
		Duration: duration,
		Command:  commands,
	}
	return reloader, nil
}

// dirWalks get all the files within `PWD`
func (reloader *Reloader) dirWalks(home string) error {
	fileinfos, err := ioutil.ReadDir(home)
	if err != nil {
		return err
	}
	pathSeparator := string(filepath.Separator)
	for _, fileinfo := range fileinfos {
		if fileinfo.IsDir() {
			home = home + pathSeparator + fileinfo.Name()
			reloader.dirWalks(home)
		} else {
			fullname := home + pathSeparator + fileinfo.Name()
			reloader.files = append(reloader.files, fullname)
		}
	}
	return nil
}

// AddFiles add ignore files to checking list
func (reloader *Reloader) AddFiles(files []string) (err error) {
	for _, filename := range files {
		ok := reloader.ignoreFilter(filename)
		// fmt.Println("filename:`"+filename, "`ok: ", ok)
		if ok == true {
			err = reloader.Watcher.Add(filename)
			if err != nil {
				// something mistakes
				fmt.Println("adding " + filename + " failed, with error:" + err.Error())
				continue
			} else {
				fmt.Println("added filename " + filename + " to filewatcher.")
			}
			reloader.files = append(reloader.files, filename)
		}
	}
	return nil
}

// EchoFiles get all the files being checked.
func (reloader *Reloader) EchoFiles() {
	fmt.Println(reloader.files)
}

// ignoreFilter filter source files if needed check.
func (reloader *Reloader) ignoreFilter(filename string) bool {
	// TODO file extension name, has has not
	// regex syntax check
	// temporary just check for golang source code.
	filename = strings.Trim(filename, "\n")
	splits := strings.Split(filename, ".")
	if splits[len(splits)-1] == "go" {
		return true
	}
	return false
}

// Start start reloading
func (reloader *Reloader) Start() error {
	commander := commands.New()
	currentPath := commander.GetOutput("pwd")
	fmt.Println("current workspace:", currentPath)
	err := reloader.dirWalks(currentPath)
	if err != nil {
		return err
	}
	// fmt.Println("files:", reloader.files)
	err = reloader.AddFiles(reloader.files)
	if err != nil {
		return err
	}
	go func() {
		for {
			fmt.Println("start method beginning...")
			time.Sleep(time.Second * 2)
			select {
			case event, ok := <-reloader.Watcher.Events:
				log.Println("event:", event)
				if !ok {
					break
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file: ", event.Name)
				}
			case err, ok := <-reloader.Watcher.Errors:
				if !ok {
					return
				}
				log.Println("error: ", err)
				// add the checking duration
				fmt.Println("looping...")
				log.Println("each loop end.")
			}
		}
	}()
	// run the main command
	fmt.Println("commands: ", reloader.Command)
	// commander.Run(reloader.Command[0], reloader.Command[1:]...)
	return nil
}
