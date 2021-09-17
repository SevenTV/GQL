package auth

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/SevenTV/ThreeLetterAPI/src/configure"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

type OAuthClient struct {
	ctx context.Context
}

var googleConfig *oauth2.Config

var YouTubeClient *youtube.Service

// getYTGConfig: retrieves google credentials and store them
func getYTGConfig() (*oauth2.Config, error) {
	if googleConfig != nil {
		return googleConfig, nil
	}

	// Read google secrets file
	b, err := ioutil.ReadFile("credentials/google.json")
	if err != nil {
		log.WithError(err).Error("auth, oauth2, unable to parse google credentials: %v", err)
		return nil, err
	}

	// Parse google config
	config, err := google.ConfigFromJSON(b, youtube.YoutubeScope)
	if err != nil {
		log.WithError(err).Error("auth, oauth2, unable to parse google config: %v", err)
		return nil, err
	}
	config.RedirectURL = fmt.Sprintf("%v/youtube", configure.Config.GetString("http.oauth_redirect_uri"))

	googleConfig = config
	return config, err
}

// GetYTGAuthURL: get an authorization url to redirect an end user to
func (c OAuthClient) GetYTGAuthURL() string {
	authURL := googleConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	return authURL
}

// GetYTGToken: exchange an authorization code for a token
func (c OAuthClient) GetYTGToken(code string) (*oauth2.Token, error) {
	tok, err := googleConfig.Exchange(oauth2.NoContext, code)

	return tok, err
}
