package interfaces

import (
	"go.aoe.com/flamingo/core/auth/application"
	"go.aoe.com/flamingo/framework/web"
	"go.aoe.com/flamingo/framework/web/responder"

	"net/url"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"go.aoe.com/flamingo/core/auth/domain"
	"go.aoe.com/flamingo/framework/flamingo"
)

type (
	// LoginController handles the login redirect
	LoginController struct {
		responder.RedirectAware `inject:""`
		AuthManager             *application.AuthManager `inject:""`
	}

	// LogoutController handles the logout
	LogoutController struct {
		responder.RedirectAware `inject:""`
		AuthManager             *application.AuthManager    `inject:""`
		EventPublisher          *application.EventPublisher `inject:""`
	}

	// CallbackController handles the oauth2.0 callback
	CallbackController struct {
		responder.RedirectAware `inject:""`
		responder.ErrorAware    `inject:""`
		AuthManager             *application.AuthManager    `inject:""`
		Logger                  flamingo.Logger             `inject:""`
		EventPublisher          *application.EventPublisher `inject:""`
	}
)

// Get handler for logins (redirect)
func (l *LoginController) Get(c web.Context) web.Response {
	redirecturl, err := c.Param1("redirecturl")
	if err != nil || redirecturl == "" {
		redirecturl = c.Request().Referer()
	}

	state := uuid.NewV4().String()
	c.Session().Values["auth.state"] = state
	c.Session().Values["auth.redirect"] = redirecturl

	return l.RedirectURL(l.AuthManager.OAuth2Config().AuthCodeURL(state))
}

// Get handler for logout
func (l *LogoutController) Get(c web.Context) web.Response {
	delete(c.Session().Values, application.KeyAuthstate)
	delete(c.Session().Values, application.KeyToken)
	delete(c.Session().Values, application.KeyRawIDToken)

	var claims struct {
		EndSessionEndpoint string `json:"end_session_endpoint"`
	}

	l.AuthManager.OpenIDProvider().Claims(&claims)
	endurl, _ := url.Parse(claims.EndSessionEndpoint)
	query := url.Values{}
	query.Set("redirect_uri", l.AuthManager.MyHost)
	endurl.RawQuery = query.Encode()

	c.Session().AddFlash("successful logged out", "warning")
	l.EventPublisher.PublishLogoutEvent(c, &domain.LogoutEvent{})

	return l.RedirectURL(endurl.String())
}

// Get handler for callbacks
func (cc *CallbackController) Get(c web.Context) web.Response {
	// Verify state and errors.
	defer delete(c.Session().Values, application.KeyAuthstate)

	if c.Session().Values[application.KeyAuthstate] != c.MustQuery1("state") {
		cc.Logger.Errorf("Invalid State %v vs %v", c.Session().Values[application.KeyAuthstate], c.MustQuery1("state"))
		return cc.Error(c, errors.New("Invalid State"))
	}

	finish := c.Profile("auth.callback", "code: "+c.MustQuery1("code"))
	oauth2Token, err := cc.AuthManager.OAuth2Config().Exchange(c, c.MustQuery1("code"))
	finish()
	if err != nil {
		cc.Logger.Errorf("core.auth.callback Error OAuth2Config Exchange %v", err)
		return cc.Error(c, errors.WithStack(err))
	}

	c.Session().Values[application.KeyToken] = oauth2Token
	c.Session().Values[application.KeyRawIDToken], err = cc.AuthManager.ExtractRawIDToken(oauth2Token)
	if err != nil {
		cc.Logger.Errorf("core.auth.callback Error ExtractRawIDToken %v", err)
		return cc.Error(c, errors.WithStack(err))
	}
	cc.EventPublisher.PublishLoginEvent(c, &domain.LoginEvent{Context: c})
	cc.Logger.Debugf("successful logged in and saved tokens: %v", oauth2Token)
	c.Session().AddFlash("successful logged in", "info")

	if redirect, ok := c.Session().Values["auth.redirect"]; ok {
		delete(c.Session().Values, "auth.redirect")
		return cc.RedirectURL(redirect.(string))
	}
	return cc.Redirect("home", nil)
}
