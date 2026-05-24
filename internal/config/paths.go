package config

import (
	"errors"
	"os"
	"path/filepath"
)

const appName = "todushka"

var ErrNoHome = errors.New("config: HOME is not set and no XDG override provided")

type Env func(string) string

func DataDir(env Env) (string, error) {
	return resolveDir(env, "XDG_DATA_HOME", filepath.Join(".local", "share"))
}

func StateDir(env Env) (string, error) {
	return resolveDir(env, "XDG_STATE_HOME", filepath.Join(".local", "state"))
}

func LogPath(env Env) (string, error) {
	dir, err := StateDir(env)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "log"), nil
}

func OSEnv(key string) string { return os.Getenv(key) }

func resolveDir(env Env, xdgKey, fallbackSubdir string) (string, error) {
	if env == nil {
		env = OSEnv
	}
	if v := env(xdgKey); v != "" {
		return filepath.Join(v, appName), nil
	}
	home := env("HOME")
	if home == "" {
		return "", ErrNoHome
	}
	return filepath.Join(home, fallbackSubdir, appName), nil
}
