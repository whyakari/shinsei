package main

import (
    "fmt"
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
    "strings"
)

func main() {
    servers := []string{"http://localhost:3000", "http://localhost:3001"}

    mux := http.NewServeMux()
    for i, server := range servers {
        target, err := url.Parse(server)
        if err != nil {
            log.Fatal("Error parsing server URL:", err)
        }
        mux.HandleFunc(fmt.Sprintf("/%d/", i), reverseProxy(target))
    }

    port := ":8080"
    log.Println("Load balancer listening on port", port)
    if err := http.ListenAndServe(port, mux); err != nil {
        log.Fatal("Error starting load balancer:", err)
    }
}

func reverseProxy(target *url.URL) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        director := func(req *http.Request) {
            req.URL.Scheme = target.Scheme
            req.URL.Host = target.Host
            req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
            req.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
            req.Header.Set("X-Forwarded-Proto", r.Header.Get("X-Forwarded-Proto"))
            req.Header.Set("X-Real-IP", r.RemoteAddr)
        }

        proxy := &httputil.ReverseProxy{Director: director}
        proxy.ServeHTTP(w, r)
    }
}

func singleJoiningSlash(a, b string) string {
    aslash := strings.HasSuffix(a, "/")
    bslash := strings.HasPrefix(b, "/")
    switch {
    case aslash && bslash:
        return a + b[1:]
    case !aslash && !bslash:
        return a + "/" + b
    }
    return a + b
}

