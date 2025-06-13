package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"strings" // Импортируем пакет strings для конкатенации
	"time"

	pb "github.com/Part001-R/grpcs/pkg/api" // Убедитесь, что пакетный путь правильный
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	ipAndPort = ":50100"
)

type server struct {
	pb.UnimplementedServSrvServer
}

func main() {

	creds, err := credentials.NewServerTLSFromFile("certs/server.crt", "certs/server.key")
	if err != nil {
		log.Fatalf("Fault load sertofacats TLS: %v", err)
	}

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

	if rxStr == "time" {
		tn := time.Now().Format("01-02-2006 15:04:05")
		return &pb.TimeResponse{StrResp: tn}, nil
	}

	return nil, errors.New("fault request")
}

// Handler - Client-streaming method for string concatenation
func (s *server) ConcatStr(stream pb.ServSrv_ConcatStrServer) error {
	log.Println("Recieve client stream for string concatenation")
	var receivedStrings []string

	for {
		req, err := stream.Recv()

		if err == io.EOF {
			log.Println("Client stream for concatenation finished")
			concatenatedString := strings.Join(receivedStrings, " ")
			return stream.SendAndClose(&pb.ConcatStrResponse{StrReq: concatenatedString})
		}

		if err != nil {
			log.Printf("Fault recieve stream message for concatenation: %v", err)
			return err
		}
		receivedStrings = append(receivedStrings, req.GetStrResp())
		log.Printf("Recieve string part from ConcatStrResponse: %s", req.GetStrResp())
	}
}
