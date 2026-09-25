package config

import (
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gotify/configor"
	"github.com/pkg/errors"
)

type Config struct {
	Root         string        `env:"JOB_SEARCH_ROOT" default:"../.."`
	Port         int           `env:"JOB_SEARCH_PORT" default:"8080"`
	Timeout      time.Duration `env:"JOB_SEARCH_TIMEOUT" default:"30m"`
	DraftTimeout time.Duration `env:"JOB_DRAFT_TIMEOUT" default:"2m"`
	DraftModel   string        `env:"JOB_DRAFT_MODEL"`
}

func Load() (Config, error) {
	var c Config
	if err := configor.Load(&c); err != nil {
		return c, errors.Wrap(err, "load config")
	}
	if c.Port < 1 || c.Port > 65535 || c.Timeout <= 0 || c.DraftTimeout <= 0 {
		return c, errors.New("invalid port or timeout")
	}
	root, err := filepath.Abs(c.Root)
	if err != nil {
		return c, errors.Wrap(err, "resolve root")
	}
	c.Root = root
	for _, name := range []string{"run-search.sh", "frontend/index.html"} {
		info, err := os.Stat(filepath.Join(root, name))
		if err != nil {
			return c, errors.Wrap(err, "validate workspace")
		}
		if !info.Mode().IsRegular() {
			return c, errors.New("invalid workspace file")
		}
	}
	return c, nil
}
func (c Config) Address() string { return net.JoinHostPort("127.0.0.1", strconv.Itoa(c.Port)) }
