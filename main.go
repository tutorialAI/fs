package main

import "log"

type ClientLimits struct {
	uploadLimit   int
	downloadLimit int
	viewLimit     int
}

func main() {

	fileServerOpts := FileServerOpts{
		StorageRoot:       ":3000_network",
		ListenAddr:        ":3000",
		PathTransformFunc: CASPathTransformFunc,
		// Transport:         tcpTransport,
		clientLimits: ClientLimits{
			uploadLimit:   10,
			downloadLimit: 10,
			viewLimit:     100,
		},
	}

	s := NewFileServer(fileServerOpts)

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}

	select {}
}
