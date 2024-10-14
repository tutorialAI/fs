package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"time"
)

const DefaultRootFolderName = "storage"
const chunkSize = 64 * 1024

type StoreOpts struct {
	Root         string
	UploadFolder string
}

type FileChunk struct {
	Data     []byte
	Chunk    int64
	ClientID int64
	FileName string
}

var FileCreation = make(map[string]time.Time)

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if len(opts.Root) == 0 {
		opts.Root = DefaultRootFolderName
	}

	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) Has(path string) bool {
	_, err := os.Stat(path)

	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

func (s *Store) Delete(clientID int64, fileName string) error {
	path := fmt.Sprintf("%s/%d/%s", s.Root, clientID, fileName)
	defer func() {
		log.Printf("deleted [%s] from disk", fileName)
	}()

	return os.RemoveAll(path)
}

func (s *Store) FileList(clientID int64) ([]fs.FileInfo, error) {
	path := fmt.Sprintf("%s/%d/", s.Root, clientID)

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	output := make([]fs.FileInfo, len(files))

	for i, file := range files {
		if file.IsDir() {
			continue
		}

		info, _ := file.Info()
		output[i] = info
	}

	return output, nil
}

func (s *Store) Download(clientID int64, fileName string) (chan FileChunk, error) {
	var err error
	path := fmt.Sprintf("%s/%d/%s", s.Root, clientID, fileName)

	if !s.Has(path) {
		return nil, fmt.Errorf("file %s not found", fileName)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return s.readStream(file)
}

func (s *Store) readStream(file *os.File) (chan FileChunk, error) {
	chunks := make(chan FileChunk)

	go func() {
		defer file.Close()
		defer close(chunks)

		var chunkNumber int64 = 0
		buffer := make([]byte, chunkSize)
		for {
			n, err := file.Read(buffer)
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println("Error reading file:", err)
				return
			}

			chunks <- FileChunk{
				Data:  buffer[:n],
				Chunk: chunkNumber,
			}
			chunkNumber++
		}
	}()

	return chunks, nil
}

func (s *Store) writeStream(data <-chan FileChunk) error {
	isFirstChunk := true
	var firstChunk FileChunk
	if isFirstChunk {
		firstChunk = <-data
		isFirstChunk = false
	}

	path := fmt.Sprintf("%s/%d/", s.Root, firstChunk.ClientID)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	path = fmt.Sprintf("%s/%s", path, firstChunk.FileName)
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer file.Close()

	file.Write(firstChunk.Data)
	for chunk := range data {
		_, err := file.Write(chunk.Data)
		if err != nil {
			return fmt.Errorf("failed to write chunk to file: %s", err)
		}
	}

	if !s.Has(path) {
		FileCreation[firstChunk.FileName] = time.Now()
	}
	log.Printf("file %s on disk", firstChunk.FileName)

	return nil
}
