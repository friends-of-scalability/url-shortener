package urlshortener

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	valid "github.com/asaskevich/govalidator"
	"github.com/friends-of-scalability/url-shortener/cmd/config"
)

// User is a representation of a User. Dah.
type shortURL struct {
	ID            uint64
	URL           string `json:"url,omitempty"`
	VisitsCounter uint64
}

type shortURLService struct {
	urlDatabase  shortURLStorage
	makeFakeLoad bool
}

func (s *shortURLService) IsHealthy() (bool, error) {
	return true, nil
}

func (s *shortURLService) generateFakeLoad(span string) error {
	duration, err := time.ParseDuration(span)
	if err != nil {
		return err
	}
	if duration.Seconds() <= 0 {
		return nil
	}
	argCPUs := fmt.Sprintf("%d", runtime.NumCPU())
	argSeconds := fmt.Sprintf("%d", int(duration.Seconds()))

	cmd := exec.Command("stress", "--cpu", strings.Trim(argCPUs, " "), "--timeout", strings.Trim(argSeconds, " "))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	go func() error {
		err = cmd.Run()
		if err != nil {
			return err
		}
		return nil
	}()
	return nil
}

// NewService gets you a shiny shortURLService!
func NewService(cfg *config.Config) (Service, error) {
	var store shortURLStorage
	var err error
	switch cfg.StorageType {
	case "inmemory":
		store, err = newInMemory()
	case "postgres":
		store, err = newPostgresStorage(
			cfg.Postgresql.Host,
			strconv.Itoa(cfg.Postgresql.Port),
			cfg.Postgresql.User,
			cfg.Postgresql.Password,
			"urlshortener")
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid storage expected inmemory or postgres got %s", cfg.StorageType)
	}

	return &shortURLService{
		urlDatabase:  store,
		makeFakeLoad: cfg.EnableFakeLoad,
	}, nil
}

// Login to the system.
func (s *shortURLService) Shortify(URL string) (mapping *shortURL, err error) {

	if !valid.IsURL(URL) {
		return nil, errMalformedURL
	}
	_, err = s.urlDatabase.ByURL(URL)
	// URL not found is an expected error, otherwise return err
	if err != errURLNotFound && err != nil {
		return nil, err
	}
	item := &shortURL{URL: URL}
	item, err = s.urlDatabase.Save(item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *shortURLService) GetInfo(shortURL string) (mapping *shortURL, err error) {
	URL, err := s.urlDatabase.ByID(shortURL)
	if err != nil {
		return nil, err
	}
	return URL, nil
}

func (s *shortURLService) Resolve(shortURL string) (mapping *shortURL, err error) {
	URL, err := s.GetInfo(shortURL)
	if err != nil {
		return nil, err
	}
	if s.makeFakeLoad {
		err = s.generateFakeLoad("5s")
		if err != nil {
			return nil, fmt.Errorf("something went wrong generating load %v", err)
		}
	}
	URL.VisitsCounter++
	return URL, nil
}
