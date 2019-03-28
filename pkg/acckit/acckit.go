package acckit

type Error struct {
	Message   string        `json:"message"`
	Type      string        `json:"type"`
	Code      int           `json:"code"`
	ErrorData []interface{} `json:"error_data"`
	FBTraceId string        `json:"fbtrace_id"`
}

type Phone struct {
	Number         string `json:"number"`
	CountryPrefix  string `json:"country_prefix"`
	NationalNumber string `json:"national_number"`
}

type Application struct {
	ID string `json:"id"`
}

type GraphAPIResponse struct {
	ID          string      `json:"id"`
	Phone       Phone       `json:"phone"`
	Application Application `json:"application"`
}

type AccessTokenResponse struct {
	ID                      string `json:"id,omitempty"`
	AccessToken             string `json:"access_token,omitempty"`
	TokenRefreshIntervalSec int    `json:"token_refresh_interval_sec,omitempty"`
}

var (
	fbAppID     string
	fbAppSecret string
)

const (
	getMeUrl       = "https://graph.accountkit.com/v1.3/me?access_token=%s&appsecret_proof=%s"
	accessTokenUrl = "https://graph.accountkit.com/v1.3/access_token?grant_type=authorization_code&code=%s&access_token=AA|%s|%s"
)

// Init Account Kit
func Init(appID string, appSecret string) {
	fbAppID = appID
	fbAppSecret = appSecret
}

// GetMe get detail user's phone info from Account Kit server
func GetMe(accessToken string) (*GraphAPIResponse, error) {
	return getMe(accessToken, fbAppSecret)
}

func GetAccessToken(authCode string) (*AccessTokenResponse, error) {
	return getAccessToken(fbAppID, fbAppSecret, authCode)
}
