package registry

import (
	"io/fs"
	"os"
	"path/filepath"
)

type InventoryReport struct {
	SchemaVersion int                 `json:"schema_version"`
	Summary       InventorySummary    `json:"summary"`
	Locations     []LocationInventory `json:"locations"`
}

type InventorySummary struct {
	Locations        int   `json:"locations"`
	Files            int   `json:"files"`
	Bytes            int64 `json:"bytes"`
	MissingLocations int   `json:"missing_locations"`
}

type LocationInventory struct {
	ID        string `json:"id"`
	Path      string `json:"path"`
	Class     string `json:"class"`
	Tracked   bool   `json:"tracked"`
	Required  bool   `json:"required"`
	Lifecycle string `json:"lifecycle"`
	Exists    bool   `json:"exists"`
	Files     int    `json:"files"`
	Bytes     int64  `json:"bytes"`
}

func InventoryRepository(root string) (InventoryReport, error) {
	reg, err := Load(root)
	if err != nil {
		return InventoryReport{}, err
	}
	return Inventory(root, reg)
}

func Inventory(root string, reg Registry) (InventoryReport, error) {
	report := InventoryReport{
		SchemaVersion: 1,
		Summary: InventorySummary{
			Locations: len(reg.Locations),
		},
		Locations: make([]LocationInventory, 0, len(reg.Locations)),
	}
	for _, location := range reg.Locations {
		entry, err := inventoryLocation(root, location)
		if err != nil {
			return InventoryReport{}, err
		}
		report.Locations = append(report.Locations, entry)
		if !entry.Exists {
			report.Summary.MissingLocations++
			continue
		}
		report.Summary.Files += entry.Files
		report.Summary.Bytes += entry.Bytes
	}
	return report, nil
}

func inventoryLocation(root string, location Location) (LocationInventory, error) {
	entry := LocationInventory{
		ID:        location.ID,
		Path:      location.Path,
		Class:     location.Class,
		Tracked:   location.Tracked,
		Required:  location.Required,
		Lifecycle: location.Lifecycle,
	}
	path := filepath.Join(root, filepath.FromSlash(location.Path))
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return entry, nil
		}
		return LocationInventory{}, err
	}
	entry.Exists = true
	if !info.IsDir() {
		entry.Files = 1
		entry.Bytes = info.Size()
		return entry, nil
	}
	err = filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		entry.Files++
		entry.Bytes += info.Size()
		return nil
	})
	if err != nil {
		return LocationInventory{}, err
	}
	return entry, nil
}
