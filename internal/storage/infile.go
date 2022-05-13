package storage

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"strings"
)

type InFile struct {
	ims      Storage
	filename string
}

type shortedURL struct {
	ID  ID     `json:"id"`
	URL string `json:"url"`
}

func NewInFile(filename string) (Storage, error) {
	stor := NewInMemory()

	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return stor, err
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
		stor.Save(s.ID, s.URL)
	}

	return &InFile{stor, filename}, nil
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
		log.Printf("storage: infile: cannot marshal shortedURL: %v", err)
		return
	}

	file, err := os.OpenFile(ifs.filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		log.Printf("storage: infile: cannot open storage file: %v", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	if _, err := writer.Write(data); err != nil {
		log.Printf("storage: infile: cannot write to storage file: %v", err)
		return
	}
	writer.WriteByte('\n')
	writer.Flush()
}
