package main

// func makeGRPCServerAndRun(listenAddr string, svc PriceService) error {
// 	grpcPriceFetcher := NewGRPCPriceFetcherServer(svc)

// 	ln, err := net.Listen("tcp", listenAddr)
// 	if err != nil {
// 		return err
// 	}

// 	opts := []grpc.ServerOption{}
// 	server := grpc.NewServer(opts...)
// 	proto.RegisterPriceFetcherServer(server, grpcPriceFetcher)

// 	return server.Serve(ln)
// }

type HelloRequest struct{}
type HelloResponse struct{}
