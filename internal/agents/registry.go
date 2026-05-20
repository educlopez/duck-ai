package agents

import "os"

var all = []Adapter{
	claudeAdapter{},
	codexAdapter{},
	openCodeAdapter{},
	genericAdapter{},
}

func All() []Adapter {
	out := make([]Adapter, len(all))
	copy(out, all)
	return out
}

func ByID(id string) (Adapter, bool) {
	for _, a := range all {
		if a.ID() == id {
			return a, true
		}
	}
	return nil, false
}

func Detected() []Adapter {
	var out []Adapter
	for _, a := range all {
		if a.Detect() {
			out = append(out, a)
		}
	}
	return out
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
