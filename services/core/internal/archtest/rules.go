// Package archtest enforces the module boundaries described in docs/architecture/ARCHITECTURE.md §7.1.
package archtest

import (
	"fmt"
	"strings"
)

const internalPrefix = "github.com/PabloViniegra/parking-app/services/core/internal/"

var infrastructurePackages = map[string]bool{
	"database/sql": true,
	"net/http":     true,
}

var allowedOwnLayers = map[string][]string{
	"domain":      {"domain"},
	"application": {"domain", "ports", "application"},
	"ports":       {"domain", "ports"},
	"adapters":    {"domain", "ports", "application", "adapters"},
}

type location struct {
	module string
	layer  string
}

func locate(pkg string) (location, bool) {
	rest, ok := strings.CutPrefix(pkg, internalPrefix)
	if !ok {
		return location{}, false
	}
	parts := strings.SplitN(rest, "/", 3)
	loc := location{module: parts[0]}
	if len(parts) > 1 {
		loc.layer = parts[1]
	}
	return loc, true
}

func isBusinessModule(module string) bool {
	switch module {
	case "shared", "bootstrap", "archtest":
		return false
	}
	return true
}

func isStdlib(pkg string) bool {
	first, _, _ := strings.Cut(pkg, "/")
	return !strings.Contains(first, ".")
}

func isInner(layer string) bool {
	return layer == "domain" || layer == "application" || layer == "ports"
}

// CheckImport returns an error when package from is not allowed to import package imp.
func CheckImport(from, imp string) error {
	src, ok := locate(from)
	if !ok || src.module == "bootstrap" || src.module == "archtest" {
		return nil
	}

	dst, internal := locate(imp)
	if !internal {
		return checkExternal(src, from, imp)
	}

	if src.module == "shared" {
		if dst.module != "shared" {
			return violation(from, imp, "shared packages must not depend on business modules")
		}
		return nil
	}

	if !isBusinessModule(dst.module) && dst.module != "shared" {
		return violation(from, imp, "business modules must not import "+dst.module)
	}

	if dst.module == "shared" {
		if src.layer == "domain" && dst.layer != "events" {
			return violation(from, imp, "domain may only import shared/events")
		}
		return nil
	}

	if dst.module != src.module {
		if dst.layer != "ports" || src.layer == "domain" {
			return violation(from, imp, "modules communicate only through events or another module's ports")
		}
		return nil
	}

	for _, allowed := range allowedOwnLayers[src.layer] {
		if dst.layer == allowed {
			return nil
		}
	}
	return violation(from, imp, src.layer+" must not depend on "+dst.layer)
}

func checkExternal(src location, from, imp string) error {
	if src.module == "shared" || !isInner(src.layer) {
		return nil
	}
	if !isStdlib(imp) {
		return violation(from, imp, src.layer+" must not import third-party packages")
	}
	if infrastructurePackages[imp] || strings.HasPrefix(imp, "net/http/") {
		return violation(from, imp, src.layer+" must not import infrastructure packages")
	}
	return nil
}

func violation(from, imp, reason string) error {
	return fmt.Errorf("%s imports %s: %s", from, imp, reason)
}
