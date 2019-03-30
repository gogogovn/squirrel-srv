package vpn

import (
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type mysqlRepository struct {
	db *sqlx.DB
}

func (m *mysqlRepository) CreateCountry(country Country) (int64, error) {
	existCountry, err := m.FindCountryByCode(country.Code)
	if existCountry != nil {
		return int64(existCountry.ID), nil
	}
	if err == ErrCountryNotFound {
		insert, args, err := sq.Insert("countries").
			Columns("name", "code").
			Values(country.Name, country.Code).
			ToSql()
		if err != nil {
			return 0, err
		}
		result, err := m.db.Exec(insert, args...)
		if err != nil {
			return 0, err
		}
		return result.LastInsertId()
	}
	return 0, err
}

func (m *mysqlRepository) FindCountryByCode(code string) (*Country, error) {
	query, args, err := sq.Select("*").From("countries").Where(sq.Eq{"code": code}).ToSql()
	if err != nil {
		return nil, err
	}
	c := Country{}
	if err := m.db.Get(&c, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCountryNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (m *mysqlRepository) FindAllCountryHaveVPNServer() ([]*Country, error) {
	query, args, err := sq.Select("countries.*").
		Distinct().
		From("countries").
		LeftJoin("vpn_servers on vpn_servers.country_id = countries.id").
		Where(sq.NotEq{"code": ""}).
		OrderBy("countries.name").
		ToSql()
	if err != nil {
		return nil, err
	}
	var countries []*Country
	rows, err := m.db.Queryx(query, args...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		c := Country{}
		err := rows.StructScan(&c)
		if err == nil {
			countries = append(countries, &c)
		}
	}
	return countries, nil
}

func (m *mysqlRepository) Create(server VPNServer) (int64, error) {
	insert, args, err := sq.Insert("vpn_servers").
		Columns("host_name",
			"ip",
			"score",
			"ping",
			"speed",
			"country_id",
			"num_vpn_sessions",
			"uptime",
			"total_users",
			"total_traffic",
			"log_type",
			"operator",
			"message",
			"open_vpn_config").
		Values(server.HostName,
			server.IP,
			server.Score,
			server.Ping,
			server.Speed,
			server.CountryID,
			server.NumVPNSessions,
			server.Uptime,
			server.TotalUsers,
			server.TotalTraffic,
			server.LogType,
			server.Operator,
			server.Message,
			server.OpenVPNConfig).
		ToSql()
	if err != nil {
		return 0, err
	}
	result, err := m.db.Exec(insert, args...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (m *mysqlRepository) FindVPNServerByCountryCode(code string) ([]*VPNServer, error) {
	country, err := m.FindCountryByCode(code)
	if err != nil {
		return nil, err
	}
	query, args, err := sq.Select(`vpn_servers.*,
		countries.name "country.name",
		countries.code "country.code",
		countries.id "country.id"`).
		Distinct().
		From("vpn_servers").
		LeftJoin("countries on countries.id = vpn_servers.country_id").
		Where(sq.Eq{"country_id": country.ID}).
		OrderBy("vpn_servers.speed desc").
		ToSql()
	if err != nil {
		return nil, err
	}
	var vpnServers []*VPNServer
	err = m.db.Select(&vpnServers, query, args...)
	if err != nil {
		return nil, err
	}
	return vpnServers, nil
}

func (m *mysqlRepository) FindAllVPNServer() ([]*VPNServer, error) {
	query, args, err := sq.Select(`vpn_servers.*,
		countries.name "country.name",
		countries.code "country.code",
		countries.id "country.id"`).
		Distinct().
		From("vpn_servers").
		LeftJoin("countries on countries.id = vpn_servers.country_id").
		OrderBy("vpn_servers.speed desc").
		ToSql()
	if err != nil {
		return nil, err
	}
	var vpnServers []*VPNServer
	err = m.db.Select(&vpnServers, query, args...)
	if err != nil {
		return nil, err
	}
	return vpnServers, nil
}

func (m *mysqlRepository) Truncate() error {
	smt := `TRUNCATE TABLE vpn_servers`
	_, err := m.db.Exec(smt)
	if err != nil {
		return err
	}
	return nil

}

func NewRepository(db *sqlx.DB) Repository {
	return &mysqlRepository{db}
}
