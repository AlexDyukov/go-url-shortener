package storage

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"
)

type InFile struct {
	ims      *InMemory
	filename string
}

type shortedURL struct {
	ID  ID     `json:"id"`
	URL string `json:"url"`
}

func NewInFile(filename string) (Storage, error) {
	stor := InMemory{sync.RWMutex{}, map[ID]string{}}

	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return &stor, err
	}

	var s shortedURL
	fileLines := strings.Split(string(fileBytes), "\n")
	for _, line := range fileLines {
		if line == "" {
			continue
		}
		if err := json.Unmarshal([]byte(line), &s); err != nil {
			log.Fatal(err)
		}
		err := stor.Save(s.ID, s.URL)
		if err != nil {
			log.Fatal(err)
		}
	}

	return &InFile{&stor, filename}, nil
}

func (ifs *InFile) Get(id ID) (string, bool) {
	return ifs.ims.Get(id)
}

func (ifs *InFile) Put(str string) (ID, error) {
	id, err := ifs.ims.Put(str)
	if err != nil {
		return id, err
	}

	if err := ifs.Save(shortedURL{id, str}); err != nil {
		log.Fatal(err)
	}

	return id, nil
}

func (ifs *InFile) Save(s shortedURL) error {
	data, err := json.Marshal(s)
	if err != nil {
		log.Fatalf("storage: infile: cannot marshal shortedURL: %v", err)
	}

	file, err := os.OpenFile(ifs.filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	if _, err := writer.Write(data); err != nil {
		return err
	}
	writer.WriteByte('\n')
	writer.Flush()

	return nil
}
