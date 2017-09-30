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

	"golang.org/x/crypto/acme/autocert"
)

func main() {
	useTLS := flag.String("tls", "", "Get a LetsEncrypt cert for `name`.")
	addr := flag.String("addr", ":8443", "Listen for http(s) connections on `addr`.")
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

	m := http.NewServeMux()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "D:", http.StatusNotImplemented)
	})
	srv := &http.Server{
		Handler: m,
	}
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
