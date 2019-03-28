package vpn

import "errors"

var (
	ErrCountryNotFound       = errors.New("country was not found")
	ErrVPNServerNotFound = errors.New("vpn server was not found")
)


type Repository interface {
	// Create country
	CreateCountry(Country) (int64, error)
	// FindCountryByCode finds a country by code
	FindCountryByCode(string) (*Country, error)
	// FindAppCountry
	FindAllCountryHaveVPNServer() ([]*Country, error)

	// Create VPNServer
	Create(VPNServer) (int64, error)
	// FindVPNServerByCountryCode
	FindVPNServerByCountryCode(string) ([]*VPNServer, error)
	// FindAllVPNServer
	FindAllVPNServer() ([]*VPNServer, error)
	// Truncate
	Truncate() error
}