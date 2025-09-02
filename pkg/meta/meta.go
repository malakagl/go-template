package meta

import "fmt"

type Meta struct {
	BuildDate string
	Version   string
	Commit    string
}

func (m *Meta) String() string {
	return fmt.Sprintf("BuildDate: %s, Version: %s, Commit: %s", m.BuildDate, m.Version, m.Commit)
}
