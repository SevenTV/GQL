package youtube

import (
	"context"
	"fmt"

	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/src/global"
	"github.com/SevenTV/GQL/src/instance"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

type youtubeInst struct {
	googleConfig *oauth2.Config

	// YouTubeClient *youtube.Service
}

func New(gCtx global.Context) (instance.Youtube, error) {
	// Parse google config
	config, err := google.ConfigFromJSON(utils.S2B(gCtx.Config().GoogleCredentials), youtube.YoutubeScope)
	if err != nil {
		logrus.WithError(err).Error("auth, oauth2, unable to parse google config")
		return nil, err
	}
	config.RedirectURL = fmt.Sprintf("%s/youtube", gCtx.Config().Http.OauthRedirectURI)

	return &youtubeInst{
		googleConfig: config,
	}, nil
}

// GetYTGAuthURL: get an authorization url to redirect an end user to
func (i *youtubeInst) GetYTGAuthURL() string {
	return i.googleConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
}

// GetYTGToken: exchange an authorization code for a token
func (i *youtubeInst) GetYTGToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return i.googleConfig.Exchange(ctx, code)
}
