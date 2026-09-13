package userspeed

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/config"
	"github.com/mhsanaei/3x-ui/v3/internal/logger"
)

const GostVersion = "3.3.0"

type DesiredRoute struct {
	ID           string
	InboundID    int
	ClientID     int
	ClientLabel  string
	ExternalPort int
	InternalPort int
	UploadMbps   int
	DownloadMbps int
	Networks     []string
}

type Status struct {
	Installed bool   `json:"installed"`
	Running   bool   `json:"running"`
	Version   string `json:"version,omitempty"`
	Services  int    `json:"services"`
	LastError string `json:"lastError,omitempty"`
}

type gostConfig struct {
	Limiters []gostLimiter `json:"limiters"`
	Services []gostService `json:"services"`
}

type gostLimiter struct {
	Name   string   `json:"name"`
	Reload string   `json:"reload"`
	File   gostFile `json:"file"`
}

type gostFile struct {
	Path string `json:"path"`
}

type gostService struct {
	Name      string        `json:"name"`
	Addr      string        `json:"addr"`
	Limiter   string        `json:"limiter"`
	Handler   gostType      `json:"handler"`
	Listener  gostListener  `json:"listener"`
	Forwarder gostForwarder `json:"forwarder"`
}

type gostType struct {
	Type string `json:"type"`
}

type gostListener struct {
	Type     string         `json:"type"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type gostForwarder struct {
	Nodes    []gostNode   `json:"nodes"`
	Selector gostSelector `json:"selector"`
}

type gostNode struct {
	Name string `json:"name"`
	Addr string `json:"addr"`
}

type gostSelector struct {
	Strategy    string `json:"strategy"`
	MaxFails    int    `json:"maxFails"`
	FailTimeout string `json:"failTimeout"`
}

type Manager struct {
	mu         sync.Mutex
	cmd        *exec.Cmd
	running    bool
	services   int
	lastConfig []byte
	lastError  string
}

var manager = &Manager{}

func GetManager() *Manager { return manager }

func BinaryName() string {
	name := fmt.Sprintf("gost-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

func BinaryPath() string {
	if value := strings.TrimSpace(os.Getenv("GOST_BINARY_PATH")); value != "" {
		return value
	}
	return filepath.Join(config.GetBinFolderPath(), BinaryName())
}

func ConfigPath() string {
	if value := strings.TrimSpace(os.Getenv("XUI_GOST_CONFIG_PATH")); value != "" {
		return value
	}
	return filepath.Join(config.GetDBFolderPath(), "gost", "config.json")
}

func CandidatePath(path, suffix string) string {
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(filepath.Base(path), ext)
	return filepath.Join(filepath.Dir(path), base+"."+suffix+ext)
}

func XrayTag(inboundID, clientID int) string {
	return fmt.Sprintf("user-speed-%d-%d", inboundID, clientID)
}

func MegabitsToBytesPerSecond(mbps int) int64 {
	if mbps <= 0 {
		return 0
	}
	return int64(mbps) * 1_000_000 / 8
}

func (m *Manager) Reconcile(routes []DesiredRoute) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sort.Slice(routes, func(i, j int) bool { return routes[i].ID < routes[j].ID })
	cfg := buildConfig(routes)
	configText, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	configText = append(configText, '\n')
	if len(routes) == 0 {
		if err := m.stopLocked(); err != nil {
			return err
		}
		if err := writeAtomicJSON(ConfigPath(), configText); err != nil {
			return err
		}
		m.lastConfig = append(m.lastConfig[:0], configText...)
		m.services = 0
		m.lastError = ""
		return nil
	}

	if runtime.GOOS == "windows" {
		return errors.New("user speed-limit runtime requires Linux")
	}
	if stat, statErr := os.Stat(BinaryPath()); statErr != nil || stat.IsDir() {
		return fmt.Errorf("GOST binary not found at %s", BinaryPath())
	}
	if bytes.Equal(configText, m.lastConfig) && m.running {
		return nil
	}

	if err := m.applyLocked(routes, configText); err != nil {
		m.lastError = err.Error()
		return err
	}
	m.lastConfig = append(m.lastConfig[:0], configText...)
	m.services = len(cfg.Services)
	m.lastError = ""
	return nil
}

func (m *Manager) applyLocked(routes []DesiredRoute, configText []byte) error {
	configPath := ConfigPath()
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		return err
	}
	limiterDir := filepath.Join(filepath.Dir(configPath), "limiters")
	if err := os.MkdirAll(limiterDir, 0o700); err != nil {
		return err
	}

	candidate := CandidatePath(configPath, fmt.Sprintf("tmp-%d-%d", os.Getpid(), time.Now().UnixMilli()))
	knownGood := CandidatePath(configPath, "known-good")
	oldConfig, oldConfigErr := os.ReadFile(configPath)
	oldLimiters, err := snapshotLimiters(limiterDir)
	if err != nil {
		return err
	}

	if err := os.WriteFile(candidate, configText, 0o600); err != nil {
		return err
	}
	defer os.Remove(candidate)
	if err := validateConfig(candidate); err != nil {
		return err
	}
	if oldConfigErr == nil {
		if err := os.WriteFile(knownGood, oldConfig, 0o600); err != nil {
			return err
		}
	}
	if err := replaceFile(candidate, configPath); err != nil {
		return err
	}
	if err := writeLimiters(limiterDir, routes); err != nil {
		_ = restoreFiles(configPath, oldConfig, oldConfigErr, limiterDir, oldLimiters)
		return err
	}
	if err := m.reloadOrStartLocked(); err != nil {
		_ = restoreFiles(configPath, oldConfig, oldConfigErr, limiterDir, oldLimiters)
		_ = m.reloadOrStartLocked()
		return err
	}
	if err := verifyListeners(routes); err != nil {
		_ = restoreFiles(configPath, oldConfig, oldConfigErr, limiterDir, oldLimiters)
		_ = m.reloadOrStartLocked()
		return err
	}
	if err := os.WriteFile(knownGood, configText, 0o600); err != nil {
		return err
	}
	for _, route := range routes {
		logger.Infof("User speed route applied: client=%s inbound=%d external=%d internal=%d upload=%dMbps download=%dMbps", route.ClientLabel, route.InboundID, route.ExternalPort, route.InternalPort, route.UploadMbps, route.DownloadMbps)
	}
	return nil
}

func buildConfig(routes []DesiredRoute) gostConfig {
	cfg := gostConfig{Limiters: make([]gostLimiter, 0, len(routes))}
	for _, route := range routes {
		limiter := "user-speed-" + route.ID
		cfg.Limiters = append(cfg.Limiters, gostLimiter{
			Name: limiter, Reload: "5s", File: gostFile{Path: filepath.Join(filepath.Dir(ConfigPath()), "limiters", limiter)},
		})
		for _, network := range route.Networks {
			listener := gostListener{Type: network}
			if network == "udp" {
				listener.Metadata = map[string]any{"keepAlive": true}
			}
			cfg.Services = append(cfg.Services, gostService{
				Name: route.ID + "-" + network, Addr: "0.0.0.0:" + strconv.Itoa(route.ExternalPort), Limiter: limiter,
				Handler: gostType{Type: network}, Listener: listener,
				Forwarder: gostForwarder{
					Nodes:    []gostNode{{Name: "xray", Addr: "127.0.0.1:" + strconv.Itoa(route.InternalPort)}},
					Selector: gostSelector{Strategy: "fifo", MaxFails: 1, FailTimeout: "10s"},
				},
			})
		}
	}
	return cfg
}

func validateConfig(path string) error {
	cmd := exec.Command(BinaryPath(), "-C", path, "-O", "json")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GOST config validation failed: %s", strings.TrimSpace(stderr.String()))
	}
	return nil
}

func writeLimiters(dir string, routes []DesiredRoute) error {
	desired := make(map[string]struct{}, len(routes))
	for _, route := range routes {
		path := filepath.Join(dir, "user-speed-"+route.ID)
		desired[path] = struct{}{}
		line := fmt.Sprintf("$ %dB %dB\n", MegabitsToBytesPerSecond(route.UploadMbps), MegabitsToBytesPerSecond(route.DownloadMbps))
		tmp := path + fmt.Sprintf(".tmp-%d-%d", os.Getpid(), time.Now().UnixNano())
		if err := os.WriteFile(tmp, []byte(line), 0o600); err != nil {
			return err
		}
		if err := replaceFile(tmp, path); err != nil {
			return err
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if _, keep := desired[path]; !keep && !strings.Contains(entry.Name(), ".tmp-") {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}
	return nil
}

func snapshotLimiters(dir string) (map[string][]byte, error) {
	out := map[string][]byte{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() || strings.Contains(entry.Name(), ".tmp-") {
			continue
		}
		value, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		out[entry.Name()] = value
	}
	return out, nil
}

func restoreFiles(configPath string, oldConfig []byte, oldConfigErr error, limiterDir string, oldLimiters map[string][]byte) error {
	var errs []error
	if oldConfigErr == nil {
		if err := writeAtomicJSON(configPath, oldConfig); err != nil {
			errs = append(errs, err)
		}
	} else if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		errs = append(errs, err)
	}
	entries, _ := os.ReadDir(limiterDir)
	for _, entry := range entries {
		if err := os.Remove(filepath.Join(limiterDir, entry.Name())); err != nil && !os.IsNotExist(err) {
			errs = append(errs, err)
		}
	}
	for name, value := range oldLimiters {
		if err := os.WriteFile(filepath.Join(limiterDir, name), value, 0o600); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func writeAtomicJSON(path string, value []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp := CandidatePath(path, fmt.Sprintf("tmp-%d-%d", os.Getpid(), time.Now().UnixNano()))
	if err := os.WriteFile(tmp, value, 0o600); err != nil {
		return err
	}
	return replaceFile(tmp, path)
}

func replaceFile(from, to string) error {
	if runtime.GOOS == "windows" {
		_ = os.Remove(to)
	}
	return os.Rename(from, to)
}

func (m *Manager) reloadOrStartLocked() error {
	if m.running && m.cmd != nil && m.cmd.Process != nil {
		if err := m.cmd.Process.Signal(syscall.SIGHUP); err == nil {
			time.Sleep(750 * time.Millisecond)
			if m.running {
				return nil
			}
		}
	}
	cmd := exec.Command(BinaryPath(), "-C", ConfigPath(), "-R", "5s")
	cmd.Stdout = &gostLogWriter{}
	cmd.Stderr = &gostLogWriter{}
	if err := cmd.Start(); err != nil {
		return err
	}
	m.cmd = cmd
	m.running = true
	go func(active *exec.Cmd) {
		err := active.Wait()
		m.mu.Lock()
		if m.cmd == active {
			m.running = false
			if err != nil {
				m.lastError = err.Error()
			}
		}
		m.mu.Unlock()
	}(cmd)
	time.Sleep(750 * time.Millisecond)
	if !m.running {
		return errors.New("GOST exited during startup")
	}
	return nil
}

func (m *Manager) StopAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stopLocked()
}

func (m *Manager) stopLocked() error {
	if !m.running || m.cmd == nil || m.cmd.Process == nil {
		m.running = false
		return nil
	}
	err := m.cmd.Process.Signal(syscall.SIGTERM)
	m.running = false
	m.cmd = nil
	return err
}

func (m *Manager) Status() Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	version := ""
	installed := false
	if stat, err := os.Stat(BinaryPath()); err == nil && !stat.IsDir() {
		installed = true
		if out, runErr := exec.Command(BinaryPath(), "-V").CombinedOutput(); runErr == nil {
			version = strings.TrimSpace(strings.Split(string(out), "\n")[0])
		}
	}
	return Status{Installed: installed, Running: m.running, Version: version, Services: m.services, LastError: m.lastError}
}

func verifyListeners(routes []DesiredRoute) error {
	deadline := time.Now().Add(10 * time.Second)
	for _, route := range routes {
		for _, network := range route.Networks {
			for {
				var ok bool
				if network == "udp" {
					conn, err := net.ListenPacket("udp4", "127.0.0.1:"+strconv.Itoa(route.ExternalPort))
					ok = err != nil
					if conn != nil {
						_ = conn.Close()
					}
				} else {
					conn, err := net.DialTimeout("tcp", "127.0.0.1:"+strconv.Itoa(route.ExternalPort), 500*time.Millisecond)
					ok = err == nil
					if conn != nil {
						_ = conn.Close()
					}
				}
				if ok {
					break
				}
				if time.Now().After(deadline) {
					return fmt.Errorf("GOST %s listener did not bind port %d", network, route.ExternalPort)
				}
				time.Sleep(250 * time.Millisecond)
			}
		}
	}
	return nil
}

type gostLogWriter struct{}

func (*gostLogWriter) Write(p []byte) (int, error) {
	line := strings.TrimSpace(string(p))
	if line != "" {
		logger.Infof("GOST: %s", line)
	}
	return len(p), nil
}
