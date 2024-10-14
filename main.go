package main

func main() {

	fileServerOpts := FileStorageOpts{
		StorageRoot: "my_storage",
		ListenAddr:  ":3000",
		Limits: Limits{
			download: 10,
			upload:   10,
			view:     10,
		},
	}

	s := NewFileStorage(&fileServerOpts)

	makeGRPCServerAndRun(s)
}
