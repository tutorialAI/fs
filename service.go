package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"sync"
)

type OperationType string

const (
	Download OperationType = "download"
	Upload   OperationType = "upload"
	View     OperationType = "view"
)

type Limits struct {
	download int
	upload   int
	view     int
}

type FileStorageOpts struct {
	ListenAddr  string
	StorageRoot string
	Limits      Limits
	mu          sync.Mutex
}

type FileStorage struct {
	*FileStorageOpts
	store *Store

	limiter map[int64]map[OperationType]chan struct{}
}

func NewFileStorage(opts *FileStorageOpts) *FileStorage {
	storeOpts := StoreOpts{
		Root: opts.StorageRoot,
	}

	return &FileStorage{
		FileStorageOpts: opts,
		store:           NewStore(storeOpts),
		limiter:         make(map[int64]map[OperationType]chan struct{}),
	}
}

func (s *FileStorage) Upload(data <-chan FileChunk) error {
	err := s.store.writeStream(data)

	if err != nil {
		return err
	}

	return nil
}

func (s *FileStorage) Download(ctx context.Context, clientID int64, fileName string) (chan FileChunk, error) {
	sem, err := s.getSemaphore(clientID, Download)

	if err != nil {
		return nil, err
	}

	select {
	case sem <- struct{}{}:
		<-sem
	case <-ctx.Done():
		return nil, fmt.Errorf("download request cancelled")
	}

	file, err := s.store.Download(clientID, fileName)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (s *FileStorage) View(ctx context.Context, clientID int64) ([]fs.FileInfo, error) {
	sem, err := s.getSemaphore(clientID, View)

	if err != nil {
		return nil, err
	}

	select {
	case sem <- struct{}{}:
		<-sem
	case <-ctx.Done():
		return nil, fmt.Errorf("view request cancelled")
	}

	files, err := s.store.FileList(clientID)
	if err != nil {
		return nil, err
	}

	return files, nil
}

/*
TODO: Основную мысль\идею на написание данного метода я вязл из stackoverflow -> https://shorturl.at/7HxNK,
Вопрос с вашим тестовым написал не я, посмотрите на дату
*/
func (s *FileStorage) getSemaphore(clientID int64, method OperationType) (chan struct{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем, есть ли запись для данного ClientID
	if _, ok := s.limiter[clientID]; !ok {
		s.limiter[clientID] = make(map[OperationType]chan struct{})
	}

	// Проверяем, есть ли семафор для данного метода ("View", "Download", "Upload")
	if sem, ok := s.limiter[clientID][method]; ok {
		return sem, nil
	}

	// Если семафора еще нет, создаем его
	limit, err := s.getLimits(method)
	if err != nil {
		return nil, err
	}

	sem := make(chan struct{}, limit)
	s.limiter[clientID][method] = sem
	return sem, nil
}

func (s *FileStorage) getLimits(method OperationType) (int, error) {
	switch method {
	case Download:
		return s.Limits.download, nil
	case Upload:
		return s.Limits.upload, nil
	case View:
		return s.Limits.view, nil
	}

	return -1, errors.New("undefined limit type")
}
