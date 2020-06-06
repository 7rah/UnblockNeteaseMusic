package proxy

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/7rah/UnblockNeteaseMusic/common"
	"github.com/7rah/UnblockNeteaseMusic/config"
	"github.com/7rah/UnblockNeteaseMusic/network"
	"github.com/7rah/UnblockNeteaseMusic/processor"
	"github.com/7rah/UnblockNeteaseMusic/version"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type HttpHandler struct{}

var localhost = map[string]int{}

func InitProxy() {
	fmt.Println("-------------------Init Proxy-------------------")
	address := "0.0.0.0:"
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				localhost[ipnet.IP.String()] = 1
			}
			if ipnet.IP.To16() != nil {
				localhost[ipnet.IP.To16().String()] = 1
			}
		}
	}
	var localhostKey []string
	for k, _ := range localhost {
		localhostKey = append(localhostKey, k)
	}
	fmt.Println("Http Proxy:")
	fmt.Println(strings.Join(localhostKey, " , "))
	go startTlsServer(address+strconv.Itoa(*config.TLSPort), *config.CertFile, *config.KeyFile, &HttpHandler{})
	go startServer(address+strconv.Itoa(*config.Port), &HttpHandler{})
}
func (h *HttpHandler) ServeHTTP(resp http.ResponseWriter, request *http.Request) {
	requestURI := request.RequestURI
	if i := strings.Index(requestURI, "/unblockmusic/"); len(requestURI) > 0 && i != -1 {
		realMusicUrl := requestURI[i+len("/unblockmusic/"):]
		//fmt.Printf("Download:%s(%s)\n", realMusicUrl, request.Method)
		realURI, err := url.Parse(realMusicUrl)
		if err != nil {
			fmt.Println("url.Parse error:", err)
			return
		}
		response, err := network.Request(&network.ClientRequest{
			Method:               request.Method,
			RemoteUrl:            realMusicUrl,
			Host:                 realURI.Host,
			Header:               request.Header,
			Body:                 request.Body,
			Cookies:              request.Cookies(),
			ForbiddenEncodeQuery: true,
			Proxy:                false,
		})
		if err != nil {
			fmt.Println("network.Request error:", err)
			return
		}
		defer response.Body.Close()
		for name, values := range response.Header {
			resp.Header()[name] = values
		}
		//fix ios 7.0.20
		resp.Header().Del("Keep-Alive")
		if response.StatusCode == 200 && (len(request.Header.Get("range")) > 0 || len(response.Header.Get("content-range")) > 0) {
			response.StatusCode = 206
		}
		resp.WriteHeader(response.StatusCode)
		_, err = io.Copy(resp, response.Body)
		if err != nil {
			//fmt.Println("io.Copy error:", err)
			return
		}
	} else {

		path := request.URL.Path
		rawQuery := request.URL.RawQuery
		uriBytes := []byte(path)
		left := uriBytes[:(len(uriBytes) / 2)]
		right := uriBytes[len(uriBytes)/2:]
		hostStr := request.URL.Host
		//fmt.Println(request.URL.String(), ",", request.Method)
		if len(hostStr) == 0 {
			hostStr = request.Host
		}
		if len(request.URL.Port()) > 0 && strings.Contains(hostStr, ":"+request.URL.Port()) {
			hostStr = strings.Replace(hostStr, ":"+request.URL.Port(), "", 1)
		}
		scheme := "http://"
		if request.TLS != nil || request.URL.Port() == "443" {
			scheme = "https://"
		}
		if len(request.URL.Scheme) > 0 {
			scheme = request.URL.Scheme + "://"
		}
		infinite := false
		for k, _ := range localhost {
			if strings.Contains(hostStr, k) {
				infinite = true
				break
			}
		}
		if infinite || strings.Contains(hostStr, "localhost") || strings.Contains(hostStr, "127.0.0.1") || strings.Contains(hostStr, "0.0.0.0") || (len(path) > 1 && strings.Count(path, "/") > 1 && bytes.EqualFold(left, right)) {
			//cause infinite loop
			requestURI = scheme + request.Host
			if bytes.EqualFold(left, right) {
				requestURI += string(left)
			} else {
				requestURI += string(uriBytes)
			}
			fmt.Printf("Abandon:%s\n", requestURI)
			resp.WriteHeader(200)
			resp.Write([]byte(version.AppVersion()))
			return
		}
		request.Host = hostStr
		if proxyDomain, ok := common.ProxyDomain[hostStr]; ok && !strings.Contains(path, "stream") {
			if request.Method == http.MethodConnect {
				proxyConnectLocalhost(resp, request)
			} else {
				if *config.Mode != 1 {
					proxyDomain = hostStr
				} else if hostIp, ok := common.HostDomain[hostStr]; ok {
					proxyDomain = hostIp
				} else {
					proxyDomain = hostStr
				}
				if len(request.URL.Port()) > 0 {
					proxyDomain = proxyDomain + ":" + request.URL.Port()
				}
				urlString := scheme + proxyDomain + path
				if len(rawQuery) > 0 {
					urlString = urlString + "?" + rawQuery
				}
				fmt.Printf("Transport:%s(%s)(%s)\n", urlString, request.Host, request.Method)
				netease := processor.RequestBefore(request)
				//fmt.Printf("{path:%s,web:%v,encrypted:%v}\n", netease.Path, netease.Web, netease.Encrypted)
				response, err := processor.Request(request, urlString)
				if err != nil {
					fmt.Println("Request error:", urlString)
					return
				}
				defer response.Body.Close()
				processor.RequestAfter(request, response, netease)
				for name, values := range response.Header {
					resp.Header()[name] = values
					//fmt.Println(name,"=",values)
				}
				resp.WriteHeader(response.StatusCode)
				_, err = io.Copy(resp, response.Body)
				if err != nil {
					fmt.Println("io.Copy error:", err)
					return
				}
				defer response.Body.Close()
				//resp.Write(body)
			}
		} else {
			if request.Method == http.MethodConnect {
				proxyConnect(resp, request)
			} else {
				if proxyDomain, ok := common.HostDomain[hostStr]; ok {
					if len(request.URL.Port()) > 0 {
						proxyDomain = proxyDomain + ":" + request.URL.Port()
					}
					requestURI = scheme + proxyDomain + path
				} else {
					if len(request.URL.Port()) > 0 {
						hostStr = hostStr + ":" + request.URL.Port()
					}
					requestURI = scheme + hostStr + path
				}
				if len(rawQuery) > 0 {
					requestURI = requestURI + "?" + rawQuery
				}

				//proxy := httputil.NewSingleHostReverseProxy(remote)
				for hostDoman, _ := range common.HostDomain {
					if strings.Contains(request.Referer(), hostDoman) {
						request.Header.Set("referer", request.Host)
						break
					}
				}
				//for key, values := range request.Header {
				//	fmt.Println(key, "=", values)
				//}
				fmt.Printf("Direct:%s(%s)(%s)\n", requestURI, request.Host, request.Method)
				response, err := network.Request(&network.ClientRequest{
					Method:    request.Method,
					RemoteUrl: requestURI,
					Host:      request.Host,
					Header:    request.Header,
					Body:      request.Body,
					Cookies:   request.Cookies(),
					Proxy:     true,
				})
				if err != nil {
					fmt.Println("network.Request error:", err)
					return
				}
				defer response.Body.Close()
				for name, values := range response.Header {
					resp.Header()[name] = values
					//fmt.Println(name,"=",values)
				}
				resp.WriteHeader(response.StatusCode)
				_, err = io.Copy(resp, response.Body)
				if err != nil {
					fmt.Println("io.Copy error:", err)
					return
				}

				//proxy.ServeHTTP(resp, request)
			}
		}
	}
}
func proxyConnectLocalhost(rw http.ResponseWriter, req *http.Request) {
	fmt.Printf("Local Received request %s %s %s\n",
		req.Method,
		req.Host,
		req.RemoteAddr,
	)
	hij, ok := rw.(http.Hijacker)
	if !ok {
		fmt.Println("HTTP Server does not support hijacking")
	}
	client, _, err := hij.Hijack()
	if err != nil {
		fmt.Println(err)
		return
	}
	localUrl := "localhost:"
	var server net.Conn
	port := req.URL.Port()
	if port == "80" || port == strconv.Itoa(*config.Port) {
		localUrl = localUrl + strconv.Itoa(*config.Port)
		server, err = net.DialTimeout("tcp", localUrl, 15*time.Second)
	} else if port == "443" || port == strconv.Itoa(*config.TLSPort) {
		localUrl = localUrl + strconv.Itoa(*config.TLSPort)
		server, err = tls.Dial("tcp", localUrl, &tls.Config{InsecureSkipVerify: true})
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	client.Write([]byte("HTTP/1.0 200 Connection Established\r\n\r\n"))
	go io.Copy(server, client)
	io.Copy(client, server)
	defer client.Close()
	defer server.Close()
}
func proxyConnect(rw http.ResponseWriter, req *http.Request) {
	fmt.Printf("Received request %s %s %s\n",
		req.Method,
		req.Host,
		req.RemoteAddr,
	)
	if req.Method != "CONNECT" {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		rw.Write([]byte("This is a http tunnel proxy, only CONNECT method is allowed."))
		return
	}
	host := req.URL.Host
	hij, ok := rw.(http.Hijacker)
	if !ok {
		fmt.Println("HTTP Server does not support hijacking")
	}
	client, _, err := hij.Hijack()
	if err != nil {
		fmt.Println(err)
		return
	}
	server, err := net.DialTimeout("tcp", host, 15*time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}
	client.Write([]byte("HTTP/1.0 200 Connection Established\r\n\r\n"))
	go io.Copy(server, client)
	io.Copy(client, server)
	defer client.Close()
	defer server.Close()
}
func startTlsServer(addr, certFile, keyFile string, handler http.Handler) {
	fmt.Printf("starting TLS Server  %s\n", addr)
	s := &http.Server{
		Addr:           addr,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := s.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		panic(err)
	}
}
func startServer(addr string, handler http.Handler) {
	fmt.Printf("starting Server  %s\n", addr)
	s := &http.Server{
		Addr:           addr,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}

}
