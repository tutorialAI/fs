package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	pb "fs/proto"

	"google.golang.org/grpc"
)

func makeGRPCServerAndRun(svc *FileStorage) error {
	grpcFileStorage := NewGRPCFileStorageServer(svc)

	ln, err := net.Listen("tcp", svc.ListenAddr)
	if err != nil {
		return err
	}

	opts := []grpc.ServerOption{}
	server := grpc.NewServer(opts...)
	pb.RegisterFileStorageServer(server, grpcFileStorage)

	fmt.Printf("server is runnig on port: %s\n", svc.ListenAddr)
	return server.Serve(ln)
}

type GRPCFileStorageServer struct {
	svc *FileStorage
	pb.UnimplementedFileStorageServer
}

func NewGRPCFileStorageServer(svc *FileStorage) *GRPCFileStorageServer {
	return &GRPCFileStorageServer{
		svc: svc,
	}
}

func (s *GRPCFileStorageServer) Upload(stream pb.FileStorage_UploadServer) error {
	chunks := make(chan FileChunk)

	go func() {
		defer close(chunks)

		for {
			chunk, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println("Error receiving chunk:", err)
			}

			chunks <- FileChunk{
				Data:     chunk.Data,
				Chunk:    chunk.Chunk,
				FileName: chunk.FileName,
				ClientID: chunk.ClientId,
			}
		}
	}()

	// i := 1
	// for c := range chunks {
	// 	fmt.Printf("--------chunk %d--------\n", i)
	// 	fmt.Println(c)
	// 	i++
	// }

	err := s.svc.Upload(chunks)
	// wg.Wait()

	if err != nil {
		return fmt.Errorf("failed to save file: %s", err)
	}

	return nil
}

func (s *GRPCFileStorageServer) Download(in *pb.DownloadRequest, stream pb.FileStorage_DownloadServer) error {
	chunks, err := s.svc.Download(stream.Context(), in.ClientId, in.Name)

	if err != nil {
		return err
	}

	for chunk := range chunks {
		grpcChunk := &pb.FileChunk{
			Data:  chunk.Data,
			Chunk: chunk.Chunk,
		}
		if err := stream.Send(grpcChunk); err != nil {
			return err
		}
	}

	return nil
}

func (s *GRPCFileStorageServer) View(ctx context.Context, in *pb.ViewRequest) (*pb.ViewResponse, error) {
	list, err := s.svc.View(ctx, in.ClientId)

	if err != nil {
		log.Fatalf("view error: %s", err)
	}

	response := make([]*pb.FileInfo, len(list))

	for i, item := range list {
		response[i] = &pb.FileInfo{
			Name:      item.Name(),
			CreatedAt: FileCreation[item.Name()].Format("2006-01-02 15:04:05"),
			UpdatedAt: item.ModTime().Format("2006-01-02 15:04:05"),
		}
	}

	return &pb.ViewResponse{Files: response, Status: "OK"}, nil
}
