package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"time"

	"github.com/agalue/k8s-grpc/proto/hello"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	server        = "192.168.0.40:9000"
	name          = "Alejandro"
	tls_enabled   = true
	ca_cert_file  = "pki/ca.pem"
	int_cert_file = "pki/support.pem"
	cli_cert_file = "pki/client.pem"
	cli_key_file  = "pki/clients-key.pem"
)

func main() {
	flag.StringVar(&name, "name", name, "Name")
	flag.StringVar(&server, "server", server, "GRPC Server Address and Port")
	flag.BoolVar(&tls_enabled, "tls", tls_enabled, "Enable/Disable TLS")
	flag.StringVar(&ca_cert_file, "ca-cert-file", ca_cert_file, "CA Root Certificate file (for Server Validation)")
	flag.StringVar(&int_cert_file, "int-cert-file", int_cert_file, "CA Intermediate Certificate file (for Server Validation)")
	flag.StringVar(&cli_cert_file, "cli-cert-file", cli_cert_file, "Client Certificate file")
	flag.StringVar(&cli_key_file, "cli-key-file", cli_key_file, "Client Certificate Key file")

	flag.Parse()

	options := []grpc.DialOption{grpc.WithBlock()}

	if tls_enabled {
		certPool := x509.NewCertPool()

		if root_cert, err := ioutil.ReadFile("pki/ca.pem"); err == nil {
			if ok := certPool.AppendCertsFromPEM(root_cert); !ok {
				log.Fatalf("failed to append CA cert: %v", err)
			}
		} else {
			log.Fatalf("cannot load Root CA certificate: %v", err)
		}

		if int_cert, err := ioutil.ReadFile("pki/support.pem"); err == nil {
			if ok := certPool.AppendCertsFromPEM(int_cert); !ok {
				log.Fatalf("failed to append Intermediate CA cert: %v", err)
			}
		} else {
			log.Fatalf("cannot load Intermediate CA certificate: %v", err)
		}

		cli_cert, err := tls.LoadX509KeyPair("pki/client.pem", "pki/client-key.pem")
		if err != nil {
			log.Fatalf("cannot load client certificate: %v", err)
		}

		cfg := &tls.Config{
			Certificates: []tls.Certificate{cli_cert},
			RootCAs:      certPool,
		}
		creds := credentials.NewTLS(cfg)
		options = append(options, grpc.WithTransportCredentials(creds))
	} else {
		options = append(options, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(server, options...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := hello.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.SayHello(ctx, &hello.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	log.Printf("Greeting: %s", r.GetMessage())
}
