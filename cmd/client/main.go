package main

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/Part001-R/grpcs/pkg/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/joho/godotenv"
)

const (
	address = "localhost:50100"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("fault read env file")
	}

	// Read public certificat
	caCert, err := os.ReadFile(os.Getenv("PATH_PUBLIC_KEY"))
	if err != nil {
		log.Fatalf("Fault read the public certificat: %v", err)
	}
	// Add sertificat to the pool
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCert)
	creds := credentials.NewClientTLSFromCert(pool, "")

	// Connecting server
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Fault connect server: %v", err)
	}
	defer conn.Close()

	// Creating a client stub
	cs := pb.NewServSrvClient(conn)

	// Call Client-streaming method
	wordsToConcat := []string{"Education", "Client-stream", "Method", "by", "gRPCS"}
	rxStr, err := reqConcatStr(cs, wordsToConcat)
	if err != nil {
		fmt.Println("Fault request concatenation:", err)
	} else {
		fmt.Println("Result client-stream -> Concatenation:", rxStr)
	}

	// Call unary method
	txStr := "time"
	rxStr, err = reqServerTime(cs, txStr)
	if err != nil {
		fmt.Println("Fault request server time:", err)
	} else {
		fmt.Println("Result unary -> Server time is:", rxStr)
	}

}

// Conatenation string by server. Return string and error.
func reqConcatStr(c pb.ServSrvClient, str []string) (string, error) {

	if c == nil {
		return "", errors.New("not recieve interface")
	}
	if len(str) < 2 {
		return "", errors.New("length slice str must be more then 1")
	}

	rxSl := make([]string, len(str), cap(str))
	copy(rxSl, str)

	ctxConcat, cancelConcat := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelConcat()

	streamConcat, err := c.ConcatStr(ctxConcat)
	if err != nil {
		return "", fmt.Errorf("fault establish client stream for concatenation: %v", err)
	}

	for _, word := range rxSl {

		req := &pb.TimeResponse{StrResp: word}
		if err := streamConcat.Send(req); err != nil {
			return "", fmt.Errorf("fault send word to stream: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	respConcat, err := streamConcat.CloseAndRecv()
	if err != nil {
		return "", fmt.Errorf("fault close stream and recieve concatenated string: %v", err)
	}

	rxStr := respConcat.GetStrReq()

	return rxStr, nil
}

// Request current time server. Return string and error.
func reqServerTime(c pb.ServSrvClient, str string) (string, error) {

	if c == nil {
		return "", errors.New("not recieve interface")
	}

	ctxUnary, cancelUnary := context.WithTimeout(context.Background(), time.Second)
	defer cancelUnary()

	reqUnary := &pb.TimeRequest{StrReq: str}
	respUnary, err := c.CurTime(ctxUnary, reqUnary)
	if err != nil {
		return "", fmt.Errorf("fault call the method CurTime: %v", err)
	}

	rxStr := respUnary.GetStrResp()
	return rxStr, nil
}
