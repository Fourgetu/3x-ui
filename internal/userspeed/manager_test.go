package userspeed

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestCandidatePathKeepsJSONExtension(t *testing.T) {
	path := CandidatePath("/var/lib/x-ui/gost/config.json", "tmp-12-34")
	if got, want := path, "/var/lib/x-ui/gost/config.tmp-12-34.json"; got != want {
		t.Fatalf("CandidatePath() = %q, want %q", got, want)
	}
	if filepath.Ext(CandidatePath("config.json", "known-good")) != ".json" {
		t.Fatal("known-good GOST config must retain the .json extension")
	}
}

func TestBuildConfigAndLimiterDirections(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XUI_GOST_CONFIG_PATH", filepath.Join(dir, "config.json"))
	routes := []DesiredRoute{
		{ID: "1-10", ExternalPort: 32001, InternalPort: 42001, UploadMbps: 10, DownloadMbps: 20, Networks: []string{"tcp"}},
		{ID: "1-11", ExternalPort: 32002, InternalPort: 42002, UploadMbps: 50, DownloadMbps: 100, Networks: []string{"tcp", "udp"}},
	}
	cfg := buildConfig(routes)
	if len(cfg.Limiters) != 2 || len(cfg.Services) != 3 {
		t.Fatalf("generated %d limiters and %d services, want 2 and 3", len(cfg.Limiters), len(cfg.Services))
	}
	limiterDir := filepath.Join(dir, "limiters")
	if err := os.MkdirAll(limiterDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := writeLimiters(limiterDir, routes); err != nil {
		t.Fatal(err)
	}
	value, err := os.ReadFile(filepath.Join(limiterDir, "user-speed-1-10"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(value), "$ 1250000B 2500000B\n"; got != want {
		t.Fatalf("limiter input/output = %q, want %q", got, want)
	}
}

func TestGost330ValidatesAllDesiredStates(t *testing.T) {
	binary := requireRealGost(t)
	dir := t.TempDir()
	t.Setenv("GOST_BINARY_PATH", binary)
	t.Setenv("XUI_GOST_CONFIG_PATH", filepath.Join(dir, "config.json"))
	cases := []struct {
		name   string
		routes []DesiredRoute
	}{
		{name: "empty desired state"},
		{name: "one TCP forward", routes: []DesiredRoute{{ID: "1-1", ExternalPort: 32001, InternalPort: 42001, Networks: []string{"tcp"}}}},
		{name: "one UDP forward", routes: []DesiredRoute{{ID: "1-2", ExternalPort: 32002, InternalPort: 42002, Networks: []string{"udp"}}}},
		{name: "multiple forwards and limiter datasource", routes: []DesiredRoute{
			{ID: "1-3", ExternalPort: 32003, InternalPort: 42003, UploadMbps: 10, DownloadMbps: 20, Networks: []string{"tcp"}},
			{ID: "1-4", ExternalPort: 32004, InternalPort: 42004, UploadMbps: 50, DownloadMbps: 100, Networks: []string{"tcp", "udp"}},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			limiterDir := filepath.Join(dir, "limiters")
			if err := os.MkdirAll(limiterDir, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := writeLimiters(limiterDir, tc.routes); err != nil {
				t.Fatal(err)
			}
			value, err := json.Marshal(buildConfig(tc.routes))
			if err != nil {
				t.Fatal(err)
			}
			candidate := CandidatePath(ConfigPath(), "test-"+strings.ReplaceAll(tc.name, " ", "-"))
			if err := os.WriteFile(candidate, value, 0o600); err != nil {
				t.Fatal(err)
			}
			if err := validateConfig(candidate); err != nil {
				t.Fatal(err)
			}
		})
	}

	good := []byte("{\"services\":[]}")
	if err := os.WriteFile(ConfigPath(), good, 0o600); err != nil {
		t.Fatal(err)
	}
	invalid := CandidatePath(ConfigPath(), "invalid")
	if err := os.WriteFile(invalid, []byte("{not-json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := validateConfig(invalid); err == nil {
		t.Fatal("invalid candidate unexpectedly passed GOST validation")
	}
	unchanged, err := os.ReadFile(ConfigPath())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(unchanged, good) {
		t.Fatal("invalid candidate changed the active known-good config")
	}
}

func TestGost330ReconcileTCPAndUDP(t *testing.T) {
	binary := requireRealGost(t)
	dir := t.TempDir()
	t.Setenv("GOST_BINARY_PATH", binary)
	t.Setenv("XUI_GOST_CONFIG_PATH", filepath.Join(dir, "config.json"))

	tcpTarget, stopTCP := startTCPEcho(t)
	defer stopTCP()
	udpTarget, stopUDP := startUDPEcho(t)
	defer stopUDP()
	tcpPublic := freeTCPPort(t)
	udpPublic := freeUDPPort(t)

	m := &Manager{}
	defer func() { _ = m.StopAll() }()
	routes := []DesiredRoute{
		{ID: "tcp", ExternalPort: tcpPublic, InternalPort: tcpTarget, Networks: []string{"tcp"}},
		{ID: "udp", ExternalPort: udpPublic, InternalPort: udpTarget, Networks: []string{"udp"}},
	}
	if err := m.Reconcile(routes); err != nil {
		t.Fatal(err)
	}
	assertTCPEcho(t, tcpPublic)
	assertUDPEcho(t, udpPublic)
	if status := m.Status(); !status.Running || status.Services != 2 {
		t.Fatalf("unexpected runtime status: %+v", status)
	}
	if err := m.Reconcile(nil); err != nil {
		t.Fatal(err)
	}
	if status := m.Status(); status.Running || status.Services != 0 {
		t.Fatalf("GOST did not stop for empty desired state: %+v", status)
	}
}

func requireRealGost(t *testing.T) string {
	t.Helper()
	if os.Getenv("GOST_REAL_E2E") != "1" {
		t.Skip("set GOST_REAL_E2E=1 with GOST_BINARY_PATH pointing to GOST v3.3.0")
	}
	binary := BinaryPath()
	out, err := exec.Command(binary, "-V").CombinedOutput()
	if err != nil {
		t.Fatalf("run GOST: %v (%s)", err, out)
	}
	if !strings.Contains(string(out), GostVersion) {
		t.Fatalf("GOST version = %q, want %s", strings.TrimSpace(string(out)), GostVersion)
	}
	return binary
}

func startTCPEcho(t *testing.T) (int, func()) {
	t.Helper()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			conn, acceptErr := listener.Accept()
			if acceptErr != nil {
				return
			}
			go func() {
				defer conn.Close()
				buffer := make([]byte, 64)
				n, _ := conn.Read(buffer)
				_, _ = conn.Write(buffer[:n])
			}()
		}
	}()
	return listener.Addr().(*net.TCPAddr).Port, func() { _ = listener.Close() }
}

func startUDPEcho(t *testing.T) (int, func()) {
	t.Helper()
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		buffer := make([]byte, 64)
		for {
			n, addr, readErr := conn.ReadFromUDP(buffer)
			if readErr != nil {
				return
			}
			_, _ = conn.WriteToUDP(buffer[:n], addr)
		}
	}()
	return conn.LocalAddr().(*net.UDPAddr).Port, func() { _ = conn.Close() }
}

func freeTCPPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()
	return port
}

func freeUDPPort(t *testing.T) int {
	t.Helper()
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	port := conn.LocalAddr().(*net.UDPAddr).Port
	_ = conn.Close()
	return port
}

func assertTCPEcho(t *testing.T, port int) {
	t.Helper()
	conn, err := net.DialTimeout("tcp4", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), 3*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	if _, err := conn.Write([]byte("tcp")); err != nil {
		t.Fatal(err)
	}
	value := make([]byte, 3)
	if _, err := io.ReadFull(conn, value); err != nil {
		t.Fatal(err)
	}
	if string(value) != "tcp" {
		t.Fatalf("TCP echo = %q", value)
	}
}

func assertUDPEcho(t *testing.T, port int) {
	t.Helper()
	conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: port})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	if _, err := conn.Write([]byte("udp")); err != nil {
		t.Fatal(err)
	}
	value := make([]byte, 3)
	if _, err := io.ReadFull(conn, value); err != nil {
		t.Fatal(err)
	}
	if string(value) != "udp" {
		t.Fatalf("UDP echo = %q", value)
	}
}
