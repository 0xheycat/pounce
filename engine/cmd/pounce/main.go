// Command pounce runs the Pounce download engine: a small HTTP daemon that
// performs segmented, resumable downloads and serves the dashboard.
package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/0xheycat/pounce/internal/api"
	"github.com/0xheycat/pounce/internal/download"
	"github.com/0xheycat/pounce/internal/model"
	"github.com/0xheycat/pounce/internal/settings"
	"github.com/0xheycat/pounce/internal/store"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:7766", "address the engine listens on")
	dataDir := flag.String("data", defaultDataDir(), "directory for engine state")
	static := flag.String("static", "", "path to the built dashboard (dist) to serve")
	authToken := flag.String("auth-token", "", "require this bearer token on /api routes (empty = no auth)")
	remote := flag.Bool("remote", false, "enable remote access: bind all interfaces and require a token (auto-generated if not set)")
	flag.Parse()

	// Ensure the web app manifest is served with a sensible content type.
	_ = mime.AddExtensionType(".webmanifest", "application/manifest+json")

	listenAddr := *addr
	token := *authToken

	if *remote {
		// Bind every interface so other devices on the network can reach Pounce.
		if host, port, err := net.SplitHostPort(listenAddr); err == nil {
			if host == "127.0.0.1" || host == "localhost" || host == "" {
				listenAddr = net.JoinHostPort("0.0.0.0", port)
			}
		}
		// Never expose the engine to the network without a token.
		if token == "" {
			token = generateToken()
			log.Printf("pounce: --remote enabled with an auto-generated token")
		}
	}

	st, err := store.New(filepath.Join(*dataDir, "meta"))
	if err != nil {
		log.Fatalf("pounce: cannot open state dir: %v", err)
	}

	set, err := settings.New(filepath.Join(*dataDir, "settings.json"))
	if err != nil {
		log.Fatalf("pounce: cannot open settings: %v", err)
	}

	hub := api.NewHub()
	mgr := download.NewManager(st, func(d *model.Download) {
		b, _ := json.Marshal(map[string]any{"type": "download", "data": d})
		hub.Broadcast(b)
	})
	if err := mgr.LoadExisting(); err != nil {
		log.Printf("pounce: could not restore previous downloads: %v", err)
	}

	srv := api.New(mgr, hub, set, *static, token)

	log.Printf("\xF0\x9F\x90\xBE Pounce engine listening on http://%s", listenAddr)
	if *static != "" {
		log.Printf("   dashboard served from %s", *static)
	}
	if token != "" {
		log.Printf("   authentication enabled (bearer token required)")
	}
	if !*remote && !isLoopback(listenAddr) && token == "" {
		log.Printf("   \u26A0 WARNING: listening on a non-loopback address without a token. Anyone on the network can control your downloads. Use --auth-token or --remote.")
	}
	if *remote {
		printPairing(listenAddr, token, *static != "")
	}

	log.Fatal(http.ListenAndServe(listenAddr, srv.Handler()))
}

func defaultDataDir() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".pounce")
	}
	return ".pounce"
}

// generateToken returns a random 128-bit hex token for remote access.
func generateToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "pounce-insecure-fallback-token"
	}
	return hex.EncodeToString(b)
}

func isLoopback(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// lanIPs returns this host's non-loopback IPv4 addresses, for device pairing.
func lanIPs() []string {
	var out []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return out
	}
	for _, ifc := range ifaces {
		if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := ifc.Addrs()
		for _, a := range addrs {
			var ip net.IP
			switch v := a.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ip4 := ip.To4(); ip4 != nil {
				out = append(out, ip4.String())
			}
		}
	}
	return out
}

// printPairing logs ready-to-open URLs (with the token) for each LAN address.
func printPairing(listenAddr, token string, hasUI bool) {
	_, port, err := net.SplitHostPort(listenAddr)
	if err != nil {
		port = "7766"
	}
	log.Printf("")
	log.Printf("\xF0\x9F\x93\xB1 Pounce Anywhere \u2014 pair a device:")
	if !hasUI {
		log.Printf("   (build the dashboard and pass --static to get the scannable QR screen)")
	}
	ips := lanIPs()
	if len(ips) == 0 {
		log.Printf("   no LAN address detected; reachable on this host at port %s", port)
	}
	for _, ip := range ips {
		log.Printf("   http://%s", net.JoinHostPort(ip, port)+"/?token="+token)
	}
	log.Printf("   token: %s", token)
	log.Printf("   Open a link above on your phone, or scan the QR in the dashboard's \xF0\x9F\x93\xB1 Pair device panel.")
	log.Printf("")
	_ = fmt.Sprintf
}
