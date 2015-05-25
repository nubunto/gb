package vendor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
)

// gb-vendor manifest support

// Manfest describes the layout of $PROJECT/vendor/vendorfile.
type Manifest struct {
	// Manifest version. Current manifest version is 0.
	Version int `json:"version"`

	// Depenencies is a list of vendored dependencies.
	Dependencies []Dependency `json:"dependencies"`
}

// AddDependency adds a Dependency to the current Manifest.
// If the dependency exists already then it returns and error.
func (m *Manifest) AddDependency(dep Dependency) error {
	if m.HasImportpath(dep.Importpath) {
		return fmt.Errorf("already registered")
	}
	m.Dependencies = append(m.Dependencies, dep)
	return nil
}

// RemoveDependency removes a Dependency from the current Manifest
// If the dependency does not exist then it returns an error
func (m *Manifest) RemoveDependency(dep Dependency) error {
	for i, d := range m.Dependencies {
		if reflect.DeepEqual(d, dep) {
			m.Dependencies = append(m.Dependencies[:i], m.Dependencies[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("dependency does not exist")
}

// HasImportpath reports whether the Manifest contains the import path.
func (m *Manifest) HasImportpath(path string) bool {
	for _, d := range m.Dependencies {
		if d.Importpath == path {
			return true
		}
	}
	return false
}

// GetDependencyForRepository return a dependency for specified URL
// If the dependency does not exist it returns an error
func (m *Manifest) GetDependencyForImportpath(path string) (Dependency, error) {
	for _, d := range m.Dependencies {
		if d.Importpath == path {
			return d, nil
		}
	}
	return Dependency{}, fmt.Errorf("dependency for %s does not exist", path)
}

// Dependency describes one vendored import path of code
// A Dependency is an Importpath sources from a Respository
// at Revision from Path.
type Dependency struct {
	// Importpath is name by which this dependency is known.
	Importpath string `json:"importpath"`

	// Repository is the remote DVCS location that this
	// dependency was fetched from.
	Repository string `json:"repository"`

	// Revision is the revision that descibves the dependency's
	// remote revision.
	Revision string `json:"revision"`

	// Branch is the branch the Revision was located on.
	// Can be blank if not needed.
	Branch string `json:"branch"`

	// Path is the path inside the Repository where the
	// dependency was fetched from.
	Path string `json:"path"`
}

// WriteManifest writes a Manifest to the path. If the manifest does
// not exist, it is created. If it does exist, it will be overwritten.
// TODO(dfc) write to temporary file and move atomically to avoid
// destroying a working vendorfile.
func WriteManifest(path string, m *Manifest) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := writeManifest(f, m); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func writeManifest(w io.Writer, m *Manifest) error {
	e := json.NewEncoder(w)
	return e.Encode(m)
}

// ReadManifest reads a Manifest from path. If the Manifest is not
// found, a blank Manifest will be returned.
func ReadManifest(path string) (*Manifest, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return new(Manifest), nil
		}
		return nil, err
	}
	defer f.Close()
	return readManifest(f)
}

func readManifest(r io.Reader) (*Manifest, error) {
	var m Manifest
	d := json.NewDecoder(r)
	err := d.Decode(&m)
	return &m, err
}
