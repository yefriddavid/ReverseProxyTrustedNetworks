package main

//import "github.com/tomasen/realip"

import (
	"flag"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"time"

	tcpProxy "github.com/jpillora/go-tcp-proxy"
	utils "github.com/jpillora/go-tcp-proxy/utils"
)

var (
	matchid = uint64(0)
	connid  = uint64(0)
	logger  tcpProxy.ColorLogger

	localAddr   = flag.String("l", ":9999", "local address")
	remoteAddr  = flag.String("r", "localhost:80", "remote address")
	verbose     = flag.Bool("v", false, "display server actions")
	veryverbose = flag.Bool("vv", false, "display server actions and all tcp data")
	nagles      = flag.Bool("n", false, "disable nagles algorithm")
	hex         = flag.Bool("h", false, "output hex")
	colors      = flag.Bool("c", false, "output ansi colors")
	unwrapTLS   = flag.Bool("unwrap-tls", false, "remote connection with TLS exposed unencrypted locally")
	match       = flag.String("match", "", "match regex (in the form 'regex')")
	replace     = flag.String("replace", "", "replace regex (in the form 'regex~replacer')")
)

func main() {
	flag.Parse()

	logger := tcpProxy.ColorLogger{
		Verbose: *verbose,
		Color:   *colors,
	}

	logger.Info("tcpProxying from %v to %v", *localAddr, *remoteAddr)

	// _, err := net.ResolveTCPAddr("tcp", *localAddr)
	laddr, err := net.ResolveTCPAddr("tcp", *localAddr)
	if err != nil {
		logger.Warn("Failed to resolve local address: %s", err)
		os.Exit(1)
	}
	raddr, err := net.ResolveTCPAddr("tcp", *remoteAddr)
	if err != nil {
		logger.Warn("Failed to resolve remote address: %s", err)
		os.Exit(1)
	}
	// listener, err := net.Listen("tcp", "localhost:7777")
	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		logger.Warn("Failed to open local port to listen: %s", err)
		os.Exit(1)
	}

	matcher := createMatcher(*match)
	replacer := createReplacer(*replace)

	if *veryverbose {
		*verbose = true
	}

	for {
		conn, _ := listener.AcceptTCP()

    fmt.Print("==================new connection==============================")
    clientIp := utils.RemoteIpConn(conn)

    fmt.Print("==================new connection==============================")


		if utils.CheckIpAccess(clientIp, "1248") == false {
			go utils.ResponseHandle(conn)
		} else {

			connid++

			var p *tcpProxy.Proxy
			if *unwrapTLS {
				logger.Info("Unwrapping TLS")
				p = tcpProxy.NewTLSUnwrapped(conn, laddr, raddr, *remoteAddr)
			} else {
				p = tcpProxy.New(conn, laddr, raddr)
			}

			p.Matcher = matcher
			p.Replacer = replacer

			p.Nagles = *nagles
			p.OutputHex = *hex
			p.Log = tcpProxy.ColorLogger{
				Verbose:     *verbose,
				VeryVerbose: *veryverbose,
				Prefix:      fmt.Sprintf("Connection #%03d ", connid),
				Color:       *colors,
			}

			fmt.Println("timeout outside")
			c1 := make(chan string, 1)
			go func() {
				time.Sleep(5 * time.Second)
				//go p.Close();
				fmt.Println("timeout inside 1")
				c1 <- "result 1"
			}()

			select {
			case res := <-c1:
				fmt.Println(res)
			case <-time.After(1 * time.Second):
				fmt.Println("timeout 1")
			}

			c2 := make(chan string, 1)
			go func() {
				time.Sleep(2 * time.Second)
				c2 <- "result 2"
			}()

			go p.StartProxy()

		}
	}
}

func handleRequest(conn net.Conn) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	// Send a response back to person contacting us.
	/*monkey := `GET / HTTP/1.0
	Host: example.com
	Connection: keep-alive`
	*/

	// conn.Write([]byte("Message received."))
	//conn.Write([]byte(monkey))
	fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")

	// Close the connection when you're done with it.
	//conn.Close()
}

func createMatcher(match string) func([]byte) {
	if match == "" {
		return nil
	}
	re, err := regexp.Compile(match)
	if err != nil {
		logger.Warn("Invalid match regex: %s", err)
		return nil
	}

	logger.Info("Matching %s", re.String())
	return func(input []byte) {
		ms := re.FindAll(input, -1)
		for _, m := range ms {
			matchid++
			logger.Info("Match #%d: %s", matchid, string(m))
		}
	}
}

func createReplacer(replace string) func([]byte) []byte {
	if replace == "" {
		return nil
	}
	//split by / (TODO: allow slash escapes)
	parts := strings.Split(replace, "~")
	if len(parts) != 2 {
		logger.Warn("Invalid replace option")
		return nil
	}

	re, err := regexp.Compile(string(parts[0]))
	if err != nil {
		logger.Warn("Invalid replace regex: %s", err)
		return nil
	}

	repl := []byte(parts[1])

	logger.Info("Replacing %s with %s", re.String(), repl)
	return func(input []byte) []byte {
		return re.ReplaceAll(input, repl)
	}
}
