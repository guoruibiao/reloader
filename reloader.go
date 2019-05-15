package reloader

import (
		"io/ioutil"
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/guoruibiao/commands"
	"fmt"
)

type Reloader struct {
	files   []string
	Watcher *fsnotify.Watcher
}

func NewReloader() (*Reloader, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	reloader := &Reloader{
		files:   nil,
		Watcher: watcher,
	}
	return reloader, nil
}


func (reloader *Reloader)dirWalks(home string) (error) {
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


// add ignore files 进行过滤
func (reloader *Reloader) AddFiles(files []string) {
	for _, filename := range files {
		if ok := reloader.ignoreFilter(filename); ok == true {
			reloader.Watcher.Add(filename)
			reloader.files = append(reloader.files, filename)
		}
	}
}

func (reloader *Reloader)EchoFiles() {
	fmt.Println(reloader.files)
}

// based on .gitignore rules and then do it.
func (reloader *Reloader) ignoreFilter(filename string) bool {

	return true
}

func (reloader *Reloader) Start() (error) {
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
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file: ", event.Name)
				}
			case err, ok := <-reloader.Watcher.Errors:
				if !ok {
					return
				}
				log.Println("error: ", err)

			}
		}
	}()
    return nil
}
