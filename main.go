package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/peixian/gooder-reads/app"

	"golang.org/x/crypto/acme/autocert"
)

func main() {
	useTLS := flag.String("tls", "", "Get a LetsEncrypt cert for `name`.")
	addr := flag.String("addr", ":8443", "Listen for http(s) connections on `addr`.")
	dsn := flag.String("dsn", "postgres:///gooder-reads?sslmode=disable", "Connect to database at `spec`.")
	flag.Parse()

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	if *useTLS != "" {
		mgr := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			Cache:      autocert.DirCache("certcache"),
			HostPolicy: autocert.HostWhitelist(*useTLS),
		}
		cfg := &tls.Config{
			GetCertificate: mgr.GetCertificate,
		}
		l = tls.NewListener(l, cfg)
	}

	a, err := app.New(*dsn)
	if err != nil {
		log.Fatal(err)
	}

	srv := &http.Server{Handler: a}
	go func() {
		if *useTLS != "" {
			log.Printf("application (via https) on %q", *addr)
		} else {
			log.Printf("application (via http) on %q", *addr)
		}
		if err := srv.Serve(l); err != nil {
			log.Println(err)
		}
	}()
	go func() {
		addr := "[::1]:6060"
		log.Printf("debug (via http) on %q", addr)
		if err := http.ListenAndServe(addr, http.DefaultServeMux); err != nil {
			log.Println(err)
		}
	}()

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	<-sig

	ctx, free := context.WithTimeout(context.Background(), time.Second*5)
	defer free()
	if err := srv.Shutdown(ctx); err != nil {
		log.Println(err)
	}
}
