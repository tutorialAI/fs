package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	pb "fs/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	fmt.Println("sending data ...")
	conn, err := grpc.NewClient("localhost:3000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := pb.NewFileStorageClient(conn)
	clientID := int64(123)

	// for i := 0; i < 20; i++ {
	// 	go func() {
	// 		data := []byte("some image bytes")
	// 		fileName := fmt.Sprintf("file_%d.jpg", i)
	// 		// fileName := "file_1.jpg"

	// 		// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// 		// defer cancel()

	// 		r, err := c.Upload(context.Background(), &pb.UploadRequest{ClientId: cleintId, Payload: data, FileName: fileName})
	// 		if err != nil {
	// 			log.Fatalf("could not upload: %s", err)
	// 		}

	// 		// time.Sleep(time.Second * 2)
	// 		log.Printf("Status: %s", r.Status)
	// 	}()
	// }

	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// defer cancel()

	upload(context.Background(), c, clientID)
	time.Sleep(time.Second * 3)
	upload(context.Background(), c, clientID)
	view(c, clientID)
}

func download(ctx context.Context, c pb.FileStorageClient, clientID int64) {
	request := &pb.DownloadRequest{
		Name:     "file_0.jpg",
		ClientId: clientID,
	}

	stream, err := c.Download(ctx, request)
	if err != nil {
		log.Fatalf("error while calling DownloadFile: %s", err)
	}

	outFile, err := os.Create("downloaded_file.jpg")
	if err != nil {
		log.Fatalf("failed to create file: %s", err)
	}
	defer outFile.Close()

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error while receiving chunk: %s", err)
		}

		if _, err := outFile.Write(chunk.Data); err != nil {
			log.Fatalf("failed to write chunk to file: %s", err)
		}

		fmt.Printf("Received chunk number: %d\n", chunk.Chunk)
	}

	fmt.Println("File downloaded successfully")
}

func upload(ctx context.Context, c pb.FileStorageClient, clientID int64) error {
	fileName := "file_to_upload.jpg"
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("failed to open file: %s", err)
	}
	defer file.Close()

	stream, err := c.Upload(ctx)
	const chunkSize = 1 * 1024 // 64 KB
	buffer := make([]byte, chunkSize)
	var chunkNumber int64 = 0

	for {
		n, err := file.Read(buffer)

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read file: %s", err)
		}

		err = stream.Send(&pb.UploadRequest{
			Data:     buffer[:n],
			Chunk:    chunkNumber,
			FileName: fileName,
			ClientId: clientID,
		})
		if err != nil {
			return fmt.Errorf("failed to send chunk: %s", err)
		}

		chunkNumber++
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("failed to receive upload status: %s", err)
	}

	fmt.Printf("Upload status: %s\n", res.Status)
	return nil
}

func view(c pb.FileStorageClient, clientID int64) {
	r, err := c.View(context.Background(), &pb.ViewRequest{ClientId: clientID})
	if err != nil {
		log.Fatalf("Could not view: %s", err)
	}

	for _, file := range r.Files {
		fmt.Printf("Name: %s | Created: %s | Updated: %s\n", file.Name, file.CreatedAt, file.UpdatedAt)
	}
}
