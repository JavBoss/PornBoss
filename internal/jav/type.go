package jav

import (
	"errors"

	"pornboss/internal/util"
)

// ResourceNotFonud indicates the requested resource is not available.
var ResourceNotFonud = errors.New("jav: resource not found")

// Provider identifies where JAV metadata or tags came from.
type Provider int

const (
	ProviderUnknown Provider = iota
	ProviderJavBus
	ProviderJavDatabase
	ProviderUser
	ProviderJavDB
)

func (p Provider) String() string {
	switch p {
	case ProviderJavBus:
		return "javbus"
	case ProviderJavDatabase:
		return "javdatabase"
	case ProviderUser:
		return "user"
	case ProviderJavDB:
		return "javdb"
	default:
		return "unknown"
	}
}

// ParseProvider converts a persisted numeric provider to a known enum.
func ParseProvider(value int) Provider {
	p := Provider(value)
	switch p {
	case ProviderJavBus, ProviderJavDatabase, ProviderUser, ProviderJavDB:
		return p
	default:
		return ProviderUnknown
	}
}

// PreferredProvider chooses the metadata source based on the system language.
func PreferredProvider() Provider {
	if util.SystemPrefersChinese() {
		return ProviderJavBus
	}
	return ProviderJavDatabase
}

// PreferredLookupProvider returns the scraper that matches the current system language.
func PreferredLookupProvider() JavLookupProvider {
	switch PreferredProvider() {
	case ProviderJavDatabase:
		return JavDatabaseProvider
	case ProviderJavDB:
		return JavDBProvider
	default:
		return JavBusProvider
	}
}

// Info holds basic metadata extracted from a JAV metadata provider.
type Info struct {
	Title       string
	Code        string
	ReleaseUnix int64
	DurationMin int
	Tags        []string
	Actors      []string
	Provider    Provider
}

// ActressInfo describes basic actress profile fields from JavDatabase.
type ActressInfo struct {
	RomanName    string
	JapaneseName string
	ChineseName  string
	HeightCM     int
	Bust         int
	Waist        int
	Hips         int
	BirthDate    int
	Cup          int
	ProfileURL   string
}

// JavLookupProvider can resolve both JAV metadata and actress profiles by code.
// Return ResourceNotFonud when the code is invalid or not found.
// Return other error for retryable lookup failures.
type JavLookupProvider interface {
	LookupActressByCode(code string) (*ActressInfo, error)
	LookupActressByJapaneseName(name string) (*ActressInfo, error)
	LookupCoverURLByCode(code string) (string, error)
	LookupJavByCode(code string) (*Info, error)
}
