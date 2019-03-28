package vpn

import (
	"encoding/csv"
	"github.com/golang/protobuf/ptypes"
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"hub.ahiho.com/ahiho/squirrel-srv/pkg/api/v1"
	"hub.ahiho.com/ahiho/squirrel-srv/pkg/logger"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

var (
	// apiVersion is version that supports by server
	apiVersion = "v1"
)

type serviceServer struct {
	repo Repository
}

func (s *serviceServer) ListCountries(context.Context, *v1.ListCountriesRequest) (*v1.ListCountriesResponse, error) {
	countries, err := s.repo.FindAllCountryHaveVPNServer()
	if err != nil {
		return nil, status.Error(codes.Unknown, "unknown error -> "+err.Error())
	}
	var resCountries []*v1.Country
	for _, c := range countries {
		resCountry := v1.Country{
			Id: c.ID,
			Name: c.Name,
			Code: c.Code,
		}
		resCountries = append(resCountries, &resCountry)
	}
	return &v1.ListCountriesResponse{
		Api:  apiVersion,
		Data: resCountries,
	}, nil
}

func (s *serviceServer) ListVPNServers(_ context.Context, req *v1.ListVPNServerRequest) (*v1.ListVPNServerResponse, error) {
	var vpns []*VPNServer
	var err error
	if len(req.CountryCode) == 0 {
		vpns, err = s.repo.FindAllVPNServer()
		if err != nil {
			return nil, status.Error(codes.Unknown, "unknown error -> "+err.Error())
		}
	} else {
		vpns, err = s.repo.FindVPNServerByCountryCode(req.CountryCode)
		if err != nil {
			return nil, status.Error(codes.Unknown, "unknown error -> "+err.Error())
		}
	}
	var resVPNs []*v1.VPNServer
	for _, v := range vpns {
		createdAt, _ := ptypes.TimestampProto(v.CreatedAt)
		updatedAt, _ := ptypes.TimestampProto(v.UpdatedAt)
		resVPN := v1.VPNServer{
			Id: v.ID,
			HostName: v.HostName,
			Ip: v.IP,
			Score: v.Score,
			Ping: v.Ping,
			Speed: v.Speed,
			Country: &v1.Country{
				Id: v.CountryID,
				Name: v.Country.Name,
				Code: v.Country.Code,
			},
			NumVPNSessions: v.NumVPNSessions,
			Uptime: v.Uptime,
			TotalUsers: v.TotalUsers,
			TotalTraffic: v.TotalTraffic,
			LogType: v.LogType,
			Operator: v.Operator,
			Message: v.Message,
			OpenVPNConfig: v.OpenVPNConfig,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
		resVPNs = append(resVPNs, &resVPN)
	}

	return &v1.ListVPNServerResponse{
		Api: apiVersion,
		Data: resVPNs,
	}, nil
}

func (s *serviceServer) VPNGateCrawler(context.Context, *v1.VPNGateCrawlerRequest) (*v1.VPNGateCrawlerResponse, error) {
	var servers []*VPNServer
	response, err := http.Get("http://www.vpngate.net/api/iphone/")
	if err != nil {
		return nil, err
	} else {
		defer response.Body.Close()
		content, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		csvString := strings.TrimLeft(string(content), "*vpn_servers\n")

		r := csv.NewReader(strings.NewReader(csvString))
		i := 0
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				logger.Log.Warn("Wrong CSV record: "+err.Error())
			} else {
				if len(record) == 15 && i > 0 {
					server := VPNServer{}
					server.HostName = record[0]
					server.IP = record[1]
					if score, err := strconv.ParseInt(record[2], 10, 32); err == nil {
						server.Score = int32(score)
					}
					if ping, err := strconv.ParseInt(record[3], 10, 16); err == nil {
						server.Ping = int32(ping)
					}
					if speed, err := strconv.ParseInt(record[4], 10, 64); err == nil {
						server.Speed = int64(speed)
					}
					country := Country{}
					country.Name = record[5]
					country.Code = record[6]
					server.Country = country
					if num, err := strconv.ParseInt(record[7], 10, 32); err == nil {
						server.NumVPNSessions = int32(num)
					}
					if upTime, err := strconv.ParseInt(record[8], 10, 64); err == nil {
						server.Uptime = int64(upTime)
					}
					if users, err := strconv.ParseInt(record[9], 10, 64); err == nil {
						server.TotalUsers = int32(users)
					}
					if traffic, err := strconv.ParseInt(record[10], 10, 64); err == nil {
						server.TotalTraffic = int64(traffic)
					}
					server.LogType = record[11]
					server.Operator = record[12]
					server.Message = record[13]
					server.OpenVPNConfig = record[14]
					servers = append(servers, &server)
				}
			}
			i++
		}
	}
	if err := s.repo.Truncate(); err == nil {
		for _, srv := range servers {
			countryID, err := s.repo.CreateCountry(srv.Country)
			if err == nil {
				srv.CountryID = int32(countryID)
				_, err := s.repo.Create(*srv)
				if err != nil {
					logger.Log.Error(err.Error())
				}
			} else {
				logger.Log.Error(err.Error())
			}
		}
		if err != nil {
			logger.Log.Error(err.Error())
		}
	}
	var resVPNs []*v1.VPNServer
	for _, v := range servers {
		createdAt, _ := ptypes.TimestampProto(v.CreatedAt)
		updatedAt, _ := ptypes.TimestampProto(v.UpdatedAt)
		resVPN := v1.VPNServer{
			Id: v.ID,
			HostName: v.HostName,
			Ip: v.IP,
			Score: v.Score,
			Ping: v.Ping,
			Speed: v.Speed,
			Country: &v1.Country{
				Id: v.CountryID,
				Name: v.Country.Name,
				Code: v.Country.Code,
			},
			NumVPNSessions: v.NumVPNSessions,
			Uptime: v.Uptime,
			TotalUsers: v.TotalUsers,
			TotalTraffic: v.TotalTraffic,
			LogType: v.LogType,
			Operator: v.Operator,
			Message: v.Message,
			OpenVPNConfig: v.OpenVPNConfig,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
		resVPNs = append(resVPNs, &resVPN)
	}

	return &v1.VPNGateCrawlerResponse{
		Api: apiVersion,
		Data: resVPNs,
	}, nil
}


func NewServiceServer(db *sqlx.DB) v1.ServiceServer {
	repo := NewRepository(db)
	return &serviceServer{repo}
}
