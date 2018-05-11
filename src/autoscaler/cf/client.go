package cf

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"code.cloudfoundry.org/cfhttp"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-incubator/uaago"

	"autoscaler/models"
)

const (
	PathCfInfo                                   = "/v2/info"
	PathCfAuth                                   = "/oauth/token"
	GrantTypePassword                            = "password"
	GrantTypeClientCredentials                   = "client_credentials"
	GrantTypeRefreshToken                        = "refresh_token"
	TimeToRefreshBeforeTokenExpire time.Duration = 10 * time.Minute
)

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type Endpoints struct {
	AuthEndpoint    string `json:"authorization_endpoint"`
	TokenEndpoint   string `json:"token_endpoint"`
	DopplerEndpoint string `json:"doppler_logging_endpoint"`
}

type CfClient interface {
	Login() error
	RefreshAuthToken() (string, error)
	GetTokens() Tokens
	GetTokensWithRefresh() Tokens
	GetEndpoints() Endpoints
	GetApp(string) (*models.AppEntity, error)
	SetAppInstances(string, int) error
	GetAuthToken() (string, error)
}

type cfClient struct {
	logger     lager.Logger
	conf       *CfConfig
	clk        clock.Clock
	tokens     Tokens
	endpoints  Endpoints
	infoUrl    string
	authUrl    string
	loginForm  url.Values
	authHeader string
	httpClient *http.Client
	lock       *sync.Mutex
	grantTime  time.Time
	uaaClient  *uaago.Client
}

func NewCfClient(conf *CfConfig, logger lager.Logger, clk clock.Clock) CfClient {
	c := &cfClient{}
	c.logger = logger
	c.conf = conf
	c.clk = clk
	c.infoUrl = conf.Api + PathCfInfo

	if conf.GrantType == GrantTypePassword {
		c.loginForm = url.Values{
			"grant_type": {GrantTypePassword},
			"username":   {conf.Username},
			"password":   {conf.Password},
		}
		c.authHeader = "Basic Y2Y6"
	} else {
		c.loginForm = url.Values{
			"grant_type":    {GrantTypeClientCredentials},
			"client_id":     {conf.ClientId},
			"client_secret": {conf.Secret},
		}
		c.authHeader = "Basic " + base64.StdEncoding.EncodeToString([]byte(conf.ClientId+":"+conf.Secret))
	}

	c.httpClient = cfhttp.NewClient()
	c.httpClient.Transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation}

	c.lock = &sync.Mutex{}

	uaaClient, err := uaago.NewClient(conf.UAAEndpoint)
	if err != nil {
		c.logger.Error("Failed to create uaago client:", err)
	}
	c.uaaClient = uaaClient

	return c
}

func (c *cfClient) GetAuthToken() (string, error) {
	authToken, err := c.uaaClient.GetAuthToken(c.conf.ClientId, c.conf.Secret, true)
	if err != nil {
		c.logger.Error("Error getting oauth token: %s. Please check your username and password", err)
		return "", err
	}
	return authToken, nil
}

func (c *cfClient) retrieveEndpoints() error {
	c.logger.Info("retrieve-endpoints", lager.Data{"infoUrl": c.infoUrl})

	resp, err := c.httpClient.Get(c.infoUrl)
	if err != nil {
		c.logger.Error("retrieve-endpoints-get", err)
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Error requesting endpoints: %s [%d] %s", c.infoUrl, resp.StatusCode, resp.Status)
		c.logger.Error("retrieve-endpoints-response", err)
		return err
	}

	err = json.NewDecoder(resp.Body).Decode(&c.endpoints)
	if err != nil {
		c.logger.Error("retrieve-endpoints-decode", err)
		return err
	}

	c.authUrl = c.endpoints.AuthEndpoint + PathCfAuth
	return nil
}

func (c *cfClient) requestTokenGrant(formData *url.Values) error {
	c.logger.Info("request-token-grant", lager.Data{"authUrl": c.authUrl, "form": *formData})

	req, err := http.NewRequest("POST", c.authUrl, strings.NewReader(formData.Encode()))
	if err != nil {
		c.logger.Error("request-token-grant-new-request", err)
		return err
	}
	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")

	var resp *http.Response
	resp, err = c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("request-token-grant-do", err)
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("request token grant failed: %s [%d] %s", c.authUrl, resp.StatusCode, resp.Status)
		c.logger.Error("request-token-grant-response", err)
		return err
	}

	err = json.NewDecoder(resp.Body).Decode(&c.tokens)
	if err != nil {
		c.logger.Error("request-token-grant-decode", err)
		return err
	}
	c.grantTime = time.Now()
	return nil
}

func (c *cfClient) Login() error {
	c.logger.Info("login", lager.Data{"infoUrl": c.infoUrl})

	err := c.retrieveEndpoints()
	if err != nil {
		return err
	}

	return c.requestTokenGrant(&c.loginForm)
}

func (c *cfClient) refresh() error {
	c.logger.Info("refresh", lager.Data{"authUrl": c.authUrl})

	if c.tokens.RefreshToken == "" {
		err := fmt.Errorf("empty refresh_token")
		c.logger.Error("refresh", err)
		return err
	}

	refreshForm := url.Values{
		"grant_type":    {GrantTypeRefreshToken},
		"refresh_token": {c.tokens.RefreshToken},
		"scope":         {""},
	}

	return c.requestTokenGrant(&refreshForm)
}

func (c *cfClient) RefreshAuthToken() (string, error) {
	c.logger.Info("refresh-auth-token", lager.Data{"authUrl": c.authUrl})

	err := c.refresh()
	if err != nil {
		c.logger.Info("refresh-auth-token-login-again")
		if err = c.Login(); err != nil {
			return "", err
		}
	}
	return TokenTypeBearer + " " + c.tokens.AccessToken, nil
}

func (c *cfClient) GetTokens() Tokens {
	return c.tokens
}

func (c *cfClient) GetTokensWithRefresh() Tokens {
	c.lock.Lock()
	if c.isTokenToBeExpired() {
		c.RefreshAuthToken()
	}
	c.lock.Unlock()
	return c.tokens
}

func (c *cfClient) isTokenToBeExpired() bool {
	return c.clk.Now().Sub(c.grantTime) > (time.Duration(c.tokens.ExpiresIn)*time.Second - TimeToRefreshBeforeTokenExpire)
}

func (c *cfClient) GetEndpoints() Endpoints {
	return c.endpoints
}
