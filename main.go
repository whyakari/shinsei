package main

import (
    "fmt"
    "io"
    "log"
    "net"
    "net/http"
    "net/url"
    "sync/atomic"
)

var (
    maxConnections = 800
    sem            = make(chan struct{}, maxConnections)
)

type RoundRobin struct {
    addrs      []string
    reqCounter uint64
}

func (rr *RoundRobin) NextServer() string {
    counter := atomic.AddUint64(&rr.reqCounter, 1)
    return rr.addrs[counter%uint64(len(rr.addrs))]
}

type ProxyHandler struct {
    loadBalancer *RoundRobin
    client       *http.Client
}

func (ph *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    sem <- struct{}{}
    defer func() { <-sem }()

    targetAddr := ph.loadBalancer.NextServer()
    targetURL, _ := url.Parse("http://" + targetAddr)

    newRequest := &http.Request{
        Method: r.Method,
        URL:    targetURL.ResolveReference(r.URL),
        Header: r.Header,
        Body:   r.Body,
    }

    resp, err := ph.client.Do(newRequest)
    if err != nil {
        log.Printf("Error proxying request: %v", err)
        http.Error(w, "Bad Gateway", http.StatusBadGateway)
        return
    }
    defer resp.Body.Close()

    for name, values := range resp.Header {
        w.Header()[name] = values
    }
    w.WriteHeader(resp.StatusCode)

    io.Copy(w, resp.Body)
}

func handleHTTPServer() {
    addrs := []string{"api01:3000", "api02:3000"}
    roundRobin := &RoundRobin{addrs: addrs}
    client := &http.Client{}

    http.Handle("/", &ProxyHandler{loadBalancer: roundRobin, client: client})

    fmt.Println("HTTP Proxy server listening on port :9999")
    if err := http.ListenAndServe(":9999", nil); err != nil {
        log.Fatal("Error starting HTTP proxy server:", err)
    }
}

func handleTCPConnection() {
    addrs := []string{"api01:3000", "api02:3000"}
    roundRobin := &RoundRobin{addrs: addrs}

    port := ":9999"
    tcpListener, err := net.Listen("tcp", port)
    if err != nil {
        log.Fatal("Error starting TCP proxy server:", err)
    }
    defer tcpListener.Close()

    fmt.Println("TCP Proxy server listening on port", port)

    for {
        downstream, err := tcpListener.Accept()
        if err != nil {
            log.Println("Error accepting TCP connection:", err)
            continue
        }
        go func(downstream net.Conn) {
            defer downstream.Close()

            sem <- struct{}{}
            defer func() { <-sem }()

            upstreamAddr := roundRobin.NextServer()
            upstream, err := net.Dial("tcp", upstreamAddr)
            if err != nil {
                log.Println("Error connecting to upstream:", err)
                return
            }
            defer upstream.Close()

            go io.Copy(upstream, downstream)
            io.Copy(downstream, upstream)
        }(downstream)
    }
}

func main() {
    //handleHTTPServer()
    handleTCPConnection()
}
