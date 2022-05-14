package storage

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

type InFile struct {
	ims      Storage
	filename string
	seek     int64
}

type shortedURL struct {
	ID  ID     `json:"id"`
	URL string `json:"url"`
}

func NewInFile(filename string) (Storage, error) {
	ifs := InFile{ims: NewInMemory(), filename: filename, seek: 0}

	if err := ifs.ReadUpdates(); err != nil {
		return &ifs, err
	}

	go ifs.Async()

	return &ifs, nil
}

func (ifs *InFile) Get(id ID) (string, bool) {
	return ifs.ims.Get(id)
}

func (ifs *InFile) Put(str string) (ID, error) {
	id := hash(str)

	link, exist := ifs.Get(id)
	if exist {
		if link == str {
			return id, nil

		}
		return id, ErrConflict{}
	}

	ifs.ims.Save(id, str)
	go ifs.Save(id, str)

	return id, nil
}

func (ifs *InFile) Save(id ID, str string) {
	data, err := json.Marshal(shortedURL{id, str})
	if err != nil {
		log.Println("storage: infile: Save: cannot marshal shortedURL:", err.Error())
		return
	}

	file, err := os.OpenFile(ifs.filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		log.Println("storage: infile: Save: cannot open storage file:", err.Error())
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	n, err := writer.Write(data)
	if err != nil {
		log.Println("storage: infile: Save: cannot write to storage file:", err.Error())
		return
	}
	writer.WriteByte('\n')
	ifs.seek += int64(n + 1)
	writer.Flush()
}

func (ifs *InFile) Async() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("storage: infile: Async: cannot initialize async worker, async disabled: ", err.Error())
		return
	}
	defer watcher.Close()

	if err = watcher.Add(ifs.filename); err != nil {
		log.Println("storage: infile: Async: cannot initialize notifies for storage file: ", err.Error())
		return
	}
	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write != fsnotify.Write {
				continue
			}
			if err := ifs.ReadUpdates(); err != nil {
				log.Println("storage: infile: ReadUpdates error:", err.Error())
			}
		case err := <-watcher.Errors:
			log.Println("storage: infile: Async error:", err.Error())
		}
	}
}

func (ifs *InFile) ReadUpdates() error {
	file, err := os.OpenFile(ifs.filename, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	if fileInfo.Size() < ifs.seek {
		ifs.seek = 0
	}

	if _, err = file.Seek(ifs.seek, io.SeekStart); err != nil {
		return err
	}

	buffered := bufio.NewReader(file)
	var s shortedURL
	var line []byte
	for {
		if line, err = buffered.ReadBytes('\n'); err != nil {
			break
		}
		if err := json.Unmarshal(line, &s); err != nil {
			log.Println("storage: infile: ReadUpdates: cannot Unmarshal shortedURL:", err.Error())
			continue
		}
		ifs.ims.Save(s.ID, s.URL)
	}

	if fileInfo, err = file.Stat(); err != nil {
		return err
	}

	ifs.seek = fileInfo.Size()
	return nil
}
