package main

import (
	"errors"
	"fmt"
	"log"
	"net"
)

type FileServerOpts struct {
	ListenAddr        string
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	// Transport         p2p.Transport
	clientLimits ClientLimits
}

type FileServer struct {
	FileServerOpts
	store     *Store
	listenner net.Listener
	quitCh    chan struct{}
}

type RequestCount struct {
	views    int
	download int
	upload   int
}

var cleintRequestCount = make(map[string]RequestCount)

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}

	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		quitCh:         make(chan struct{}),
	}
}

func (s *FileServer) Start() error {
	if err := ListenAndAccept(s); err != nil {
		return err
	}

	// s.loop()

	return nil
}

func ListenAndAccept(s *FileServer) error {
	var err error
	s.listenner, err = net.Listen("tcp", s.FileServerOpts.ListenAddr)

	if err != nil {
		return err
	}

	go s.startAcceptLoop()
	log.Printf("TCP listining on port: %s \n", s.FileServerOpts.ListenAddr)

	return err
}

func (s *FileServer) startAcceptLoop() {
	for {
		conn, err := s.listenner.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}

		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
			continue
		}

		fmt.Printf("new incoming connection %+v\n", conn)

		go s.hanndleConn(conn)
	}
}

func (s *FileServer) hanndleConn(conn net.Conn) {
	var err error

	defer func() {
		fmt.Printf("Droping the connection %s\n", err)
		conn.Close()
	}()

	//peer := NewPeer(conn)

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("TCP read error %s", err)
			continue
		}

		host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
		if err != nil {
			fmt.Printf("error while spliting host and port %s", err)
			continue
		}
		fmt.Printf("message %s, from %s\n", buf[:n], host)

	}

	// Read loop
	// rpc := RPC{}
	// for {
	// 	err := t.Decoder.Decode(conn, &rpc)
	// 	if err == net.ErrClosed {
	// 		return
	// 	}

	// 	if err != nil {
	// 		fmt.Printf("TCP read error %s", err)
	// 		continue
	// 	}

	// 	rpc.From = conn.RemoteAddr()
	// 	t.rpcch <- rpc

	// 	// fmt.Printf("message %+v, from %s\n", buf[:n], msg.From.String())
	// }
}
