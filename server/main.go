package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/agalue/k8s-grpc/proto/hello"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	port          = 9000
	tls_enabled   = true
	ca_cert_file  = "pki/ca.pem"
	int_cert_file = "pki/support.pem"
	srv_cert_file = "pki/server.pem"
	srv_key_file  = "pki/server-key.pem"
)

type server struct {
	hello.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, in *hello.HelloRequest) (*hello.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())

	host, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return &hello.HelloReply{Message: fmt.Sprintf("Hello %s from %s", in.GetName(), host)}, nil
}

func main() {
	flag.IntVar(&port, "port", port, "Listen Port")
	flag.BoolVar(&tls_enabled, "tls", tls_enabled, "Enable/Disable TLS")
	flag.StringVar(&ca_cert_file, "ca-cert-file", ca_cert_file, "CA Root Certificate file (for Client Validation)")
	flag.StringVar(&int_cert_file, "int-cert-file", int_cert_file, "CA Intermediate Certificate file (for Client Validation)")
	flag.StringVar(&srv_cert_file, "srv-cert-file", srv_cert_file, "Server Certificate file")
	flag.StringVar(&srv_key_file, "srv-key-file", srv_key_file, "Server Certificate Key file")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	options := make([]grpc.ServerOption, 0)

	if tls_enabled {
		certPool := x509.NewCertPool()

		if root_cert, err := ioutil.ReadFile(ca_cert_file); err == nil {
			if ok := certPool.AppendCertsFromPEM(root_cert); !ok {
				log.Fatalf("failed to append CA cert: %v", err)
			}
		} else {
			log.Fatalf("cannot load Root CA certificate: %v", err)
		}

		if int_cert, err := ioutil.ReadFile(int_cert_file); err == nil {
			if ok := certPool.AppendCertsFromPEM(int_cert); !ok {
				log.Fatalf("failed to append Intermediate CA cert: %v", err)
			}
		} else {
			log.Fatalf("cannot load Intermediate CA certificate: %v", err)
		}

		srv_cert, err := tls.LoadX509KeyPair(srv_cert_file, srv_key_file)
		if err != nil {
			log.Fatalf("cannot load server certificate: %v", err)
		}

		cfg := &tls.Config{
			Certificates: []tls.Certificate{srv_cert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    certPool,
		}
		creds := credentials.NewTLS(cfg)
		options = append(options, grpc.Creds(creds))
	}

	s := grpc.NewServer(options...)
	hello.RegisterGreeterServer(s, &server{})

	go func() {
		log.Printf("server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	s.GracefulStop()
	log.Println("Good bye")
}
