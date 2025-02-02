package main

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Program struct {
	Pid   int    `mapstructure:"PID"`
	Log   string `mapstructure:"LOG"`
	Debug bool   `mapstructure:"DEBUG"`
}

type Directories struct {
	Persist string `mapstructure:"PERSIST"`
	Owner   string `mapstructure:"OWNER"`
	Group   string `mapstructure:"GROUP"`
}

type Remote struct {
	User    string   `mapstructure:"USER"`
	Persist int      `mapstructure:"PERSIST"`
	Args    []string `mapstructure:"ARGS"`
}

type Commands struct {
	Ssh             string   `mapstructure:"SSH"`
	Pre             []string `mapstructure:"PRE"`
	Ffmpeg          string   `mapstructure:"FFMPEG"`
	Ffprobe         string   `mapstructure:"FFPROBE"`
	FallbackFfmpeg  string   `mapstructure:"FALLBACK_FFMPEG"`
	FallbackFfprobe string   `mapstructure:"FALLBACK_FFPROBE"`
	SpecialFlags    []string `mapstructure:"SPECIAL_FLAGS"`
}

type Database struct {
	Type        string `mapstructure:"TYPE"`
	Path        string `mapstructure:"PATH"`
	MigratorDir string `mapstructure:"MIGRATOR_DIR"`
	Host        string `mapstructure:"HOST"`
	Port        int    `mapstructure:"PORT"`
	Name        string `mapstructure:"NAME"`
	Username    string `mapstructure:"USERNAME"`
	Password    string `mapstructure:"PASSWORD"`
}

type Config struct {
	Program     Program     `mapstructure:"PROGRAM"`
	Directories Directories `mapstructure:"DIRECTORIES"`
	Remote      Remote      `mapstructure:"REMOTE"`
	Commands    Commands    `mapstructure:"COMMANDS"`
	Database    Database    `mapstructure:"DATABASE"`
}

func LoadConfig() (Config, error) {
	config := Config{
		Program: Program{
			Log:   "/var/log/jellyfin",
			Debug: false,
		},
		Directories: Directories{
			Persist: "/run/shm",
			Owner:   "jellyfin",
			Group:   "jellyfin",
		},
		Remote: Remote{
			User:    "jellyfin",
			Persist: 300,
			Args: []string{
				"-i",
				"/var/lib/rffmpeg/.ssh/id_ed25519",
			},
		},
		Commands: Commands{
			Ssh:             "/usr/bin/ssh",
			Pre:             []string{},
			Ffmpeg:          "/usr/lib/jellyfin-ffmpeg/ffmpeg",
			Ffprobe:         "/usr/lib/jellyfin-ffmpeg/ffprobe",
			FallbackFfmpeg:  "/usr/lib/jellyfin-ffmpeg/ffmpeg",
			FallbackFfprobe: "/usr/lib/jellyfin-ffmpeg/ffprobe",
			SpecialFlags:    []string{},
		},
		Database: Database{
			Type:     "sqlite",
			Path:     "/var/lib/rffmpeg/db",
			Host:     "localhost",
			Port:     5432,
			Name:     "rffmpeg",
			Username: "postgres",
		},
	}

	viper.AddConfigPath("/etc/rffmpeg")
	viper.SetConfigName("rffmpeg")
	viper.SetConfigType("yaml")

	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("rffmpeg")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return config, fmt.Errorf("failed parsing config: %w", err)
		}
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		return config, fmt.Errorf("failed unmarshaling config: %w", err)
	}

	config.Program.Pid = os.Getpid()

	year, month, day := time.Now().Date()
	datetime := fmt.Sprintf("%d%d%d", year, month, day)
	if month < 10 {
		if day < 10 {
			datetime = fmt.Sprintf("%d0%d0%d", year, month, day)
		} else {
			datetime = fmt.Sprintf("%d0%d%d", year, month, day)
		}
	} else {
		if day < 10 {
			datetime = fmt.Sprintf("%d%d0%d", year, month, day)
		}
	}
	if config.Program.Log[len(config.Program.Log)-1] == '/' {
		config.Program.Log = config.Program.Log + "log_rffmpeg_" + datetime + ".log"
	} else {
		config.Program.Log = config.Program.Log + "/log_rffmpeg_" + datetime + ".log"
	}

	if config.Database.Path[len(config.Database.Path)-1] == '/' {
		config.Database.Path = config.Database.Path + "rffmpeg.db"
	} else {
		config.Database.Path = config.Database.Path + "/rffmpeg.db"
	}

	switch config.Database.Type {
	case "sqlite":
		{
			err := os.MkdirAll(filepath.Dir(config.Database.Path), os.ModePerm)
			if err != nil {
				return config, fmt.Errorf("failed creating database directory: %w", err)
			}
			dbpath, err := filepath.Abs(config.Database.Path)
			if err != nil {
				return config, fmt.Errorf("failed loading sqlite file: %w", err)
			}
			config.Database.Path = dbpath + "?_foreign_keys=on"
			config.Database.MigratorDir = "migrations/sqlite"
		}
	case "postgres":
		{
			config.Database.Path = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", config.Database.Host, config.Database.Port, config.Database.Username, config.Database.Password, config.Database.Name)
			config.Database.MigratorDir = "migrations/postgres"
		}
	default:
		return config, fmt.Errorf("database type isn't supported")
	}

	defaultSpecialFlags := []string{
		"-version",
		"-encoders",
		"-decoders",
		"-hwaccels",
		"-filters",
		"-h",
		"-muxers",
		"-fp_format",
	}
	config.Commands.SpecialFlags = append(config.Commands.SpecialFlags, defaultSpecialFlags...)

	return config, err
}
