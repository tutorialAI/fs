package main

import (
	"net"
)

type Peer struct {
	conn net.Conn
	host string

	store Store
}

func NewPeer(conn net.Conn) *Peer {
	return &Peer{
		conn: conn,
	}
}
