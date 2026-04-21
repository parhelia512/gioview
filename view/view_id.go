package view

import (
	"fmt"
	"net/url"
	"runtime"
	"strings"
)

func (id ViewID) Name() string {
	return id.name
}

func (id ViewID) Path() url.URL {
	u, err := url.Parse(fmt.Sprintf("gioview://%s/%s", id.path, id.name))
	if err != nil {
		panic("Invalid view id")
	}

	return url.URL{
		Scheme: "gioview",
		Host:   u.Host,
		Path:   u.Path,
	}
}

func (id ViewID) String() string {
	return fmt.Sprintf("%s/%s", id.path, id.name)
}

// NewViewID constructs a ViewID from a name, and the caller's
// function call path. The function call path is converted to a
// URL like path. Please be sure to place the call of NewViewID
// along with the view definition to get an accurate path.
func NewViewID(name string) ViewID {
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		funcName := runtime.FuncForPC(pc).Name()
		lastSlash := strings.LastIndexByte(funcName, '/')
		if lastSlash < 0 {
			lastSlash = 0
		}
		lastDot := strings.LastIndexByte(funcName[lastSlash:], '.') + lastSlash
		return ViewID{name: name, path: funcName[:lastDot]}
	}

	return ViewID{name: name}
}
