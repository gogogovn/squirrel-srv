package vpn

import "time"

// Country entity
type Country struct {
	ID        int32      `db:"id"`
	Name      string     `db:"name"`
	Code      string     `db:"code"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

// VPNServer entity
type VPNServer struct {
	ID             int32      `db:"id"`
	HostName       string     `db:"host_name"`
	IP             string     `db:"ip"`
	Score          int32      `db:"score"`
	Ping           int32      `db:"ping"`
	Speed          int64      `db:"speed"`
	CountryID      int32      `db:"country_id"`
	NumVPNSessions int32      `db:"num_vpn_sessions"`
	Uptime         int64      `db:"uptime"`
	TotalUsers     int32      `db:"total_users"`
	TotalTraffic   int64      `db:"total_traffic"`
	LogType        string     `db:"log_type"`
	Operator       string     `db:"operator"`
	Message        string     `db:"message"`
	OpenVPNConfig  string     `db:"open_vpn_config"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
	DeletedAt      *time.Time `db:"deleted_at"`
	Country        `db:",prefix=country"`
}
