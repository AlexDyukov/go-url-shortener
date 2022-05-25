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
	Surl ShortURL `json:"id"`
	Furl FullURL  `json:"url"`
	User User     `json:"user"`
}

type userURL struct {
	Surl ShortURL `json:"short_url"`
	Furl FullURL  `json:"original_url"`
}

func NewInFile(filename string) (Storage, error) {
	ifs := InFile{ims: NewInMemory(), filename: filename, seek: 0}

	if err := ifs.readUpdates(); err != nil {
		return &ifs, err
	}

	go ifs.async()

	return &ifs, nil
}

func (ifs *InFile) Get(surl ShortURL) (FullURL, bool) {
	return ifs.ims.Get(surl)
}

func (ifs *InFile) Save(surl ShortURL, furl FullURL) error {
	if err := ifs.ims.Save(surl, furl); err != nil {
		return err
	}

	go ifs.writeUpdate(shortedURL{surl, furl, DefaultUserID})

	return nil
}

func (ifs *InFile) Put(furl FullURL) (ShortURL, error) {
	surl, err := ifs.ims.Put(furl)
	if err != nil {
		return surl, err
	}

	go ifs.writeUpdate(shortedURL{surl, furl, DefaultUserID})

	return surl, nil
}

func (ifs *InFile) SetAuthor(surl ShortURL, furl FullURL, user User) error {
	if err := ifs.ims.SetAuthor(surl, furl, user); err != nil {
		return err
	}

	go ifs.writeUpdate(shortedURL{surl, furl, user})

	return nil
}

func (ifs *InFile) GetAuthorURLs(user User) (URLs, bool) {
	return ifs.ims.GetAuthorURLs(user)
}

func (ifs *InFile) async() {
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
			if err := ifs.readUpdates(); err != nil {
				log.Println("storage: infile: ReadUpdates error:", err.Error())
			}
		case err := <-watcher.Errors:
			log.Println("storage: infile: Async error:", err.Error())
		}
	}
}

func (ifs *InFile) writeUpdate(s shortedURL) {
	data, err := json.Marshal(s)
	if err != nil {
		log.Println("storage: infile: writeUpdate: cannot marshal shortedURL:", err.Error())
		return
	}

	file, err := os.OpenFile(ifs.filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Println("storage: infile: writeUpdate: cannot marshal shortedURL:", err.Error())
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	n, err := writer.Write(data)
	if err != nil {
		log.Println("storage: infile: writeUpdate: cannot write to buffer:", err.Error())
		return
	}
	writer.WriteByte('\n')
	ifs.seek += int64(n + 1)
	if err := writer.Flush(); err != nil {
		log.Println("storage: infile: writeUpdate: cannot write to file:", err.Error())
		return
	}
}

func (ifs *InFile) readUpdates() error {
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
		ifs.ims.SetAuthor(s.Surl, s.Furl, s.User)
	}

	if fileInfo, err = file.Stat(); err != nil {
		return err
	}

	ifs.seek = fileInfo.Size()
	return nil
}
