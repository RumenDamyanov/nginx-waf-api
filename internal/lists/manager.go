package lists

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var validNameRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)

// Manager handles IP list file operations.
type Manager struct {
	baseDir string
}

// NewManager creates a list manager for the given directory.
func NewManager(baseDir string) *Manager {
	return &Manager{baseDir: baseDir}
}

// ListInfo describes an IP list file.
type ListInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Entries int    `json:"entries"`
	ModTime string `json:"mod_time"`
}

// ListDetail describes an IP list with its entries.
type ListDetail struct {
	ListInfo
	IPs []string `json:"ips"`
}

// List returns all IP list files in the base directory.
func (m *Manager) List() ([]ListInfo, error) {
	entries, err := os.ReadDir(m.baseDir)
	if err != nil {
		return nil, fmt.Errorf("read lists dir: %w", err)
	}

	var lists []ListInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".txt") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		ips, _ := m.readIPs(e.Name())
		name := strings.TrimSuffix(e.Name(), ".txt")
		lists = append(lists, ListInfo{
			Name:    name,
			Path:    e.Name(),
			Entries: len(ips),
			ModTime: info.ModTime().UTC().Format(time.RFC3339),
		})
	}
	sort.Slice(lists, func(i, j int) bool { return lists[i].Name < lists[j].Name })
	return lists, nil
}

// Get returns details of a specific list.
func (m *Manager) Get(name string) (*ListDetail, error) {
	if !validNameRe.MatchString(name) {
		return nil, fmt.Errorf("invalid list name")
	}

	filename := name + ".txt"
	ips, err := m.readIPs(filename)
	if err != nil {
		return nil, err
	}

	fpath := filepath.Join(m.baseDir, filename)
	info, err := os.Stat(fpath)
	if err != nil {
		return nil, fmt.Errorf("stat: %w", err)
	}

	return &ListDetail{
		ListInfo: ListInfo{
			Name:    name,
			Path:    filename,
			Entries: len(ips),
			ModTime: info.ModTime().UTC().Format(time.RFC3339),
		},
		IPs: ips,
	}, nil
}

// AddEntry adds an IP/CIDR to a list. Creates the file if needed.
func (m *Manager) AddEntry(name, ip string) error {
	if !validNameRe.MatchString(name) {
		return fmt.Errorf("invalid list name")
	}
	ip = strings.TrimSpace(ip)
	if !isValidIP(ip) {
		return fmt.Errorf("invalid IP or CIDR: %s", ip)
	}

	filename := name + ".txt"
	existing, _ := m.readIPs(filename)

	for _, e := range existing {
		if e == ip {
			return fmt.Errorf("entry already exists: %s", ip)
		}
	}

	existing = append(existing, ip)
	return m.writeIPs(filename, name, existing)
}

// RemoveEntry removes an IP/CIDR from a list.
func (m *Manager) RemoveEntry(name, ip string) error {
	if !validNameRe.MatchString(name) {
		return fmt.Errorf("invalid list name")
	}
	ip = strings.TrimSpace(ip)

	filename := name + ".txt"
	existing, err := m.readIPs(filename)
	if err != nil {
		return err
	}

	found := false
	var updated []string
	for _, e := range existing {
		if e == ip {
			found = true
			continue
		}
		updated = append(updated, e)
	}

	if !found {
		return fmt.Errorf("entry not found: %s", ip)
	}

	return m.writeIPs(filename, name, updated)
}

func (m *Manager) readIPs(filename string) ([]string, error) {
	fpath := filepath.Join(m.baseDir, filename)
	f, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var ips []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ips = append(ips, line)
	}
	return ips, scanner.Err()
}

func (m *Manager) writeIPs(filename, listName string, ips []string) error {
	if err := os.MkdirAll(m.baseDir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	fpath := filepath.Join(m.baseDir, filename)
	tmp, err := os.CreateTemp(m.baseDir, ".waf-api-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()

	success := false
	defer func() {
		if !success {
			tmp.Close()
			os.Remove(tmpName)
		}
	}()

	header := fmt.Sprintf("# Managed by nginx-waf-api\n# List: %s\n# Updated: %s\n# Entries: %d\n",
		listName, time.Now().UTC().Format(time.RFC3339), len(ips))

	if _, err := tmp.WriteString(header); err != nil {
		return err
	}
	for _, ip := range ips {
		if _, err := fmt.Fprintln(tmp, ip); err != nil {
			return err
		}
	}

	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmpName, fpath); err != nil {
		return err
	}

	success = true
	return nil
}

func isValidIP(s string) bool {
	if strings.Contains(s, "/") {
		_, _, err := net.ParseCIDR(s)
		return err == nil
	}
	return net.ParseIP(s) != nil
}
