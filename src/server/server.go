package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/casapps/caspbx/src/config"
	"github.com/casapps/caspbx/src/server/handler"
	"github.com/casapps/caspbx/src/server/model"
	"github.com/casapps/caspbx/src/server/service"
	"github.com/casapps/caspbx/src/server/store"
)

type Bootstrap struct {
	Routes RouteCatalog
}

type App struct {
	Bootstrap Bootstrap
	mux       *http.ServeMux
	auth      service.AuthService
}

func NewBootstrap(apiVersion string, adminPath string) (Bootstrap, error) {
	routes, routeError := NewRouteCatalog(apiVersion, adminPath)
	if routeError != nil {
		return Bootstrap{}, routeError
	}
	return Bootstrap{Routes: routes}, nil
}

func NewApp(apiVersion string, adminPath string, projectName string, version string, commitID string, officialSite string) (App, error) {
	defaultConfig := config.DefaultConfig()
	return NewAppWithStore(apiVersion, adminPath, projectName, version, commitID, officialSite, defaultConfig.Server, store.NewMemoryStore())
}

func NewAppWithStore(apiVersion string, adminPath string, projectName string, version string, commitID string, officialSite string, serverConfig config.ServerConfig, runtimeStore store.RuntimeStore) (App, error) {
	bootstrap, bootstrapError := NewBootstrap(apiVersion, adminPath)
	if bootstrapError != nil {
		return App{}, bootstrapError
	}

	authService := service.NewAuthService(runtimeStore, service.SessionConfig{
		AdminTTL:         timeDurationHours(serverConfig.Session.Admin.MaxAgeHours),
		UserTTL:          timeDurationHours(serverConfig.Session.User.MaxAgeHours),
		ExtendOnActivity: serverConfig.Session.ExtendOnActivity,
	})

	healthResponse := handler.HealthResponse{
		Status:                "ok",
		Project:               projectName,
		Version:               version,
		CommitID:              commitID,
		APIBasePath:           bootstrap.Routes.APIBasePath,
		AdminPath:             bootstrap.Routes.AdminBasePath,
		AsteriskAdminPath:     bootstrap.Routes.AsteriskAdminBasePath,
		OfficialSite:          officialSite,
		RuntimeImplementation: "scaffold",
	}

	userCookie := handler.SessionCookieConfig{
		Name:     serverConfig.Session.User.CookieName,
		Path:     "/",
		HTTPOnly: serverConfig.Session.HTTPOnly,
		Secure:   serverConfig.Session.Secure,
		SameSite: sameSiteMode(serverConfig.Session.SameSite),
	}
	adminCookie := handler.SessionCookieConfig{
		Name:     serverConfig.Session.Admin.CookieName,
		Path:     bootstrap.Routes.AdminBasePath,
		HTTPOnly: serverConfig.Session.HTTPOnly,
		Secure:   serverConfig.Session.Secure,
		SameSite: sameSiteMode(serverConfig.Session.SameSite),
	}

	mux := http.NewServeMux()
	mux.Handle("/", handler.NewRootHandler(projectName, officialSite, bootstrap.Routes.AdminBasePath, bootstrap.Routes.APIBasePath))
	mux.Handle("/health", handler.NewHealthHandler(healthResponse))
	mux.Handle("/healthz", handler.NewHealthHandler(healthResponse))
	mux.Handle("/version", handler.NewVersionHandler(projectName, version, commitID))
	registerSurface(mux, bootstrap.Routes.AuthBasePath, handler.NewAuthHandler(bootstrap.Routes.AuthBasePath, authService, userCookie, model.DefaultRegistrationMode()))
	registerSurface(mux, bootstrap.Routes.AuthAPIBasePath, handler.NewAPIAuthHandler(bootstrap.Routes.AuthAPIBasePath, authService, model.DefaultRegistrationMode()))
	registerSurface(mux, bootstrap.Routes.UsersBasePath, handler.NewUserHandler(bootstrap.Routes.UsersBasePath, authService, userCookie))
	registerSurface(mux, bootstrap.Routes.UsersAPIBasePath, handler.NewAPIUserHandler(bootstrap.Routes.UsersAPIBasePath, authService))
	registerSurface(mux, bootstrap.Routes.OrgsBasePath, handler.NewOrgHandler(bootstrap.Routes.OrgsBasePath, authService, runtimeStore, userCookie))
	registerSurface(mux, bootstrap.Routes.OrgsAPIBasePath, handler.NewAPIOrgHandler(bootstrap.Routes.OrgsAPIBasePath, authService, runtimeStore))
	registerSurface(mux, bootstrap.Routes.AdminBasePath, handler.NewAdminHandler(bootstrap.Routes.AdminBasePath, authService, adminCookie))
	registerSurface(mux, bootstrap.Routes.AdminAPIBasePath, handler.NewAPIAdminHandler(bootstrap.Routes.AdminAPIBasePath, authService))
	registerSurface(mux, bootstrap.Routes.AsteriskAdminBasePath, handler.NewAdminHandler(bootstrap.Routes.AsteriskAdminBasePath, authService, adminCookie))
	registerSurface(mux, bootstrap.Routes.AsteriskAdminAPIPath, handler.NewAPIAdminHandler(bootstrap.Routes.AsteriskAdminAPIPath, authService))

	return App{Bootstrap: bootstrap, mux: mux, auth: authService}, nil
}

func (bootstrap Bootstrap) Summary() string {
	return fmt.Sprintf(
		"API base path: %s\nAdmin path: %s\nAsterisk admin path: %s",
		bootstrap.Routes.APIBasePath,
		bootstrap.Routes.AdminBasePath,
		bootstrap.Routes.AsteriskAdminBasePath,
	)
}

func (app App) Handler() http.Handler {
	return app.mux
}

func (app App) Summary() string {
	return fmt.Sprintf(
		"%s\nHealth path: /health\nVersion path: /version",
		app.Bootstrap.Summary(),
	)
}

func registerSurface(mux *http.ServeMux, routePrefix string, surfaceHandler http.Handler) {
	mux.Handle(routePrefix, surfaceHandler)
	mux.Handle(routePrefix+"/", surfaceHandler)
}

func sameSiteMode(value string) http.SameSite {
	switch value {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

func timeDurationHours(hours int) time.Duration {
	return time.Duration(hours) * time.Hour
}
