package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"strings" // Импортируем пакет strings для конкатенации
	"time"

	pb "github.com/Part001-R/grpcs/pkg/api" // Убедитесь, что пакетный путь правильный
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/joho/godotenv"
)

type server struct {
	pb.UnimplementedServSrvServer
}

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("fault read env file")
	}

	creds, err := credentials.NewServerTLSFromFile(os.Getenv("PATH_PUBLIC_KEY"), os.Getenv("PATH_PRIVATE_KEY"))
	if err != nil {
		log.Fatalf("Fault load sertificats TLS: %v", err)
	}

	ipAndPort := os.Getenv("PORT")
	listener, err := net.Listen("tcp", ipAndPort)
	if err != nil {
		log.Fatalf("Fault create listener tcp port %s: %v", ipAndPort, err)
	}

	srv := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterServSrvServer(srv, &server{})
	log.Println("Start up server:", ipAndPort)

	err = srv.Serve(listener)
	if err != nil {
		log.Fatalf("Fault start up server: %v ", err)
	}
}

// Handler - Unary Current time server
func (s *server) CurTime(ctx context.Context, req *pb.TimeRequest) (*pb.TimeResponse, error) {
	rxStr := req.GetStrReq()
	log.Printf("Recive request time server: %s", rxStr)

	if rxStr == "time" {
		tn := time.Now().Format("01-02-2006 15:04:05")
		return &pb.TimeResponse{StrResp: tn}, nil
	}

	return nil, errors.New("fault request")
}

// Handler - Client-streaming method for string concatenation
func (s *server) ConcatStr(stream pb.ServSrv_ConcatStrServer) error {
	var receivedStrings []string

	for {
		req, err := stream.Recv()

		if err == io.EOF {
			concatenatedString := strings.Join(receivedStrings, " ")
			log.Printf("Recive slice: %v\n", receivedStrings)
			log.Printf("Result concatenation: %s\n", concatenatedString)
			return stream.SendAndClose(&pb.ConcatStrResponse{StrReq: concatenatedString})
		}
		if err != nil {
			log.Printf("Fault recieve stream message for concatenation: %v", err)
			return err
		}

		rxStr := req.GetStrResp()
		receivedStrings = append(receivedStrings, rxStr)
	}
}
