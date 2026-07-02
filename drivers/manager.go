package drivers

import (
	"embed"
	"encoding/json"
	"sort"
	"strings"
)

//go:embed data
var embeddedData embed.FS

type Registry struct {
	Version   int      `json:"version"`
	Timestamp string   `json:"timestamp"`
	Drivers   []Driver `json:"drivers"`
}

type Driver struct {
	ID          string   `json:"id"`
	Vendor      string   `json:"vendor"`
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	DeviceTypes []string `json:"deviceTypes"`
	Aliases     []string `json:"aliases"`
	DownloadURL string   `json:"downloadUrl"`
	SupportURL  string   `json:"supportUrl"`
	Notes       string   `json:"notes"`
}

type Manager struct {
	reg Registry
}

func NewManager() *Manager {
	m := &Manager{}
	_ = m.LoadEmbedded()
	return m
}

func (m *Manager) LoadEmbedded() error {
	data, err := embeddedData.ReadFile("data/drivers.json")
	if err != nil {
		return err
	}
	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return err
	}
	m.reg = reg
	return nil
}

func (m *Manager) GetAll() []Driver {
	out := make([]Driver, len(m.reg.Drivers))
	copy(out, m.reg.Drivers)
	sortDrivers(out)
	return out
}

func (m *Manager) Search(query, category string) []Driver {
	q := strings.TrimSpace(strings.ToLower(query))
	cat := strings.TrimSpace(strings.ToLower(category))
	var out []Driver
	for _, d := range m.reg.Drivers {
		if cat != "" && cat != "all" && strings.ToLower(d.Category) != cat {
			continue
		}
		if q == "" || matches(d, q) {
			out = append(out, d)
		}
	}
	sortDrivers(out)
	return out
}

func (m *Manager) Categories() []string {
	seen := map[string]bool{}
	for _, d := range m.reg.Drivers {
		seen[d.Category] = true
	}
	cats := make([]string, 0, len(seen))
	for c := range seen {
		cats = append(cats, c)
	}
	sort.Strings(cats)
	return cats
}

func (m *Manager) Timestamp() string {
	return m.reg.Timestamp
}

func matches(d Driver, q string) bool {
	fields := []string{d.ID, d.Vendor, d.Name, d.Category, d.Notes}
	fields = append(fields, d.DeviceTypes...)
	fields = append(fields, d.Aliases...)
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), q) {
			return true
		}
	}
	return false
}

func sortDrivers(drivers []Driver) {
	sort.Slice(drivers, func(i, j int) bool {
		if drivers[i].Category == drivers[j].Category {
			if drivers[i].Vendor == drivers[j].Vendor {
				return drivers[i].Name < drivers[j].Name
			}
			return drivers[i].Vendor < drivers[j].Vendor
		}
		return drivers[i].Category < drivers[j].Category
	})
}
