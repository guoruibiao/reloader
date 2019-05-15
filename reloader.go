package reloader

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
		"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/guoruibiao/commands"
	"time"
)

// Reloader auto reload programs by checking file modified or not based on EPOOL
type Reloader struct {
	files    []string          // files in `PWD` with recursion way
	Watcher  *fsnotify.Watcher // based on epool of system supported
	Duration int     // duration to check if files whether be modified, seconds
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
func (reloader *Reloader) AddFiles(files []string) {
	for _, filename := range files {
		if ok := reloader.ignoreFilter(filename); ok == true {
			reloader.Watcher.Add(filename)
			reloader.files = append(reloader.files, filename)
		}
	}
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
	splits := strings.Split(filename, "\n")
	if splits[len(splits)-1] == "go" {
		return true
	}
	return false
}

// Start start reloading
func (reloader *Reloader) Start() error {
	commander := commands.New()
	currentPath := commander.GetOutput("pwd")
	err := reloader.dirWalks(currentPath)
	if err != nil {
		return err
	}
	reloader.AddFiles(reloader.files)
	go func() {
		for {
			select {
			case event, ok := <-reloader.Watcher.Events:
				if !ok {
					break
				}
				// log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file: ", event.Name)
				}
			case err, ok := <-reloader.Watcher.Errors:
				if !ok {
					return
				}
				log.Println("error: ", err)

			}
			// add the checking duration
			time.Sleep(time.Duration(reloader.Duration))
		}
	}()
	// run the main command
	commander.Run(reloader.Command[0], reloader.Command[1:]...)
	return nil
}
