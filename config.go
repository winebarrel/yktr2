package yktr2

import (
	"fmt"
	"net/url"
)

type Oauth2Config struct {
	ClientID     string `toml:"client_id"`
	ClientSecret string `toml:"client_secret"`
	RedirectHost string `toml:"redirect_host"`
}

type Config struct {
	Addr          string `default:"127.0.0.1"`
	Port          int    `default:"8080"`
	PerPage       int    `toml:"per_page" default:"5"`
	Team          string `toml:"team"`
	SessionSecret string `toml:"session_secret"`
	Oauth2        Oauth2Config
}

func (cfg *Config) Validate() error {
	if cfg.Team == "" {
		return fmt.Errorf("config error: 'team' is required")
	}

	if cfg.SessionSecret == "" {
		return fmt.Errorf("config error: 'session_secret' is required")
	}

	if cfg.Oauth2.ClientID == "" {
		return fmt.Errorf("config error: 'oauth2.client_id' is required")
	}

	if cfg.Oauth2.ClientSecret == "" {
		return fmt.Errorf("config error: 'oauth2.client_secret' is required")
	}

	if cfg.Oauth2.RedirectHost == "" {
		return fmt.Errorf("config error: 'oauth2.redirect_host' is required")
	}

	_, err := url.Parse(cfg.Oauth2.RedirectHost)

	if err != nil {
		return fmt.Errorf("config error: 'oauth2.redirect_host' is invalid url: %s", err)
	}

	return nil
}
