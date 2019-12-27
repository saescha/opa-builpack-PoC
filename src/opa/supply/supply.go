package supply

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cloudfoundry/libbuildpack"
)

type Stager interface {
	//TODO: See more options at https://github.com/cloudfoundry/libbuildpack/blob/master/stager.go
	AddBinDependencyLink(string, string) error
	BuildDir() string
	DepDir() string
	DepsIdx() string
	DepsDir() string
	WriteProfileD(string, string) error
}

type Manifest interface {
	//TODO: See more options at https://github.com/cloudfoundry/libbuildpack/blob/master/manifest.go
	AllDependencyVersions(string) []string
	DefaultVersion(string) (libbuildpack.Dependency, error)
	RootDir() string
}

type Installer interface {
	//TODO: See more options at https://github.com/cloudfoundry/libbuildpack/blob/master/installer.go
	InstallDependency(libbuildpack.Dependency, string) error
	InstallOnlyVersion(string, string) error
}

type Command interface {
	//TODO: See more options at https://github.com/cloudfoundry/libbuildpack/blob/master/command.go
	Execute(string, io.Writer, io.Writer, string, ...string) error
	Output(dir string, program string, args ...string) (string, error)
}

type Supplier struct {
	Manifest  Manifest
	Installer Installer
	Stager    Stager
	Command   Command
	Log       *libbuildpack.Logger
	Config    Config
}

type Config struct {
	OpaVersion                 string       `yaml:"opa_version"`
	AuthorzationContentVersion string       `yaml:"authorization_content_version"`
	AdcPort                    int          `yaml:"adc_port"`
	Bundle                     BundleConfig `yaml:"bundle"`
}

type BundleConfig struct {
	Polling PollingConfig `yaml:"polling"`
}

type PollingConfig struct {
	Min int `yaml:"min_delay_seconds"`
	Max int `yaml:"max_delay_seconds"`
}

type OpaCfg struct {
	Services OpaServiceCfg `yaml:"services"`
	Bundles  OpaBundleCfg  `yaml:"bundles"`
}

type OpaServiceCfg struct {
	BundleProvider OpaBundleProviderCfg `yaml:"bundle-provider"`
}
type OpaBundleProviderCfg struct {
	Url         string               `yaml:"url"`
	Credentials OpaCredentialsConfig `yaml:"credentials"`
}
type OpaCredentialsConfig struct {
	Url       string       `yaml:"url"`
	clientTls OpaClientTls `yaml:"client_tls"`
}
type OpaClientTls struct {
	Cert       string `yaml:"cert"`
	privateKey string `yaml:"private_key"`
}
type OpaBundleCfg struct {
	Authz OpaAuthzCfg `yaml:"authz"`
}
type OpaAuthzCfg struct {
	Service  string        `yaml:"service"`
	Resource string        `yaml:"resource"`
	Polling  PollingConfig `yaml:"polling"`
}

func New(stager Stager, manifest Manifest, installer Installer, logger *libbuildpack.Logger, command Command) *Supplier {
	return &Supplier{
		Stager:    stager,
		Manifest:  manifest,
		Installer: installer,
		Log:       logger,
		Command:   command,
	}
}

func (s *Supplier) Run() error {
	s.Log.BeginStep("Supplying opa")

	if err := s.Setup(); err != nil {
		s.Log.Error("Could not setup: %s", err.Error())
		return err
	}

	if err := s.InstallOPA(); err != nil {
		s.Log.Error("Could not install opa: %s", err.Error())
		return err
	}

	return nil
}

func (s *Supplier) Setup() error {
	configPath := filepath.Join(s.Stager.BuildDir(), "ADCConfig.yml")
	if exists, err := libbuildpack.FileExists(configPath); err != nil {
		return err
	} else if exists {
		if err := libbuildpack.NewYAML().Load(configPath, &s.Config); err != nil {
			return err
		}
	}

	logsDirPath := filepath.Join(s.Stager.BuildDir(), "logs")
	if err := os.Mkdir(logsDirPath, os.ModePerm); err != nil {
		return fmt.Errorf("Could not create 'logs' directory: %v", err)
	}

	return nil
}

func (s *Supplier) InstallOPA() error {
	dep, err := s.findMatchingVersion("opa", s.Config.OpaVersion)
	if err != nil {
		s.Log.Info(`Available versions: ` + strings.Join(s.availableVersions(), ", "))
		return fmt.Errorf("Could not determine version: %s", err)
	}
	if s.Config.OpaVersion == "" {
		s.Log.BeginStep("No OPA version specified - using default => %s", dep.Version)
	} else {
		s.Log.BeginStep("Requested OPA version: %s => %s", s.Config.OpaVersion, dep.Version)
	}

	dir := filepath.Join(s.Stager.DepDir(), "opa")

	if err := s.Installer.InstallDependency(dep, dir); err != nil {
		return err
	}

	return s.Stager.AddBinDependencyLink(filepath.Join(dir, "opa", "sbin", "opa"), "opa")
}
func (s *Supplier) availableVersions() []string {
	allVersions := s.Manifest.AllDependencyVersions("opa")
	allNames := []string{}
	allSemver := []string{}
	sort.Strings(allNames)
	sort.Strings(allSemver)

	return append(append(allNames, allSemver...), allVersions...)
}

func (s *Supplier) findMatchingVersion(depName string, version string) (libbuildpack.Dependency, error) {
	if version == "" {
		ver, err := s.Manifest.DefaultVersion(depName)
		if err != nil {
			return libbuildpack.Dependency{}, err
		}
		version = ver.Version
	}

	versions := s.Manifest.AllDependencyVersions(depName)
	if ver, err := libbuildpack.FindMatchingVersion(version, versions); err != nil {
		return libbuildpack.Dependency{}, err
	} else {
		version = ver
	}

	return libbuildpack.Dependency{Name: depName, Version: version}, nil
}
