package supply

import (
	"io"

	"github.com/cloudfoundry/libbuildpack"
)

type Stager interface {
	//TODO: See more options at https://github.com/cloudfoundry/libbuildpack/blob/master/stager.go
	BuildDir() string
	DepDir() string
	DepsIdx() string
	DepsDir() string
}

type Manifest interface {
	//TODO: See more options at https://github.com/cloudfoundry/libbuildpack/blob/master/manifest.go
	AllDependencyVersions(string) []string
	DefaultVersion(string) (libbuildpack.Dependency, error)
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

func (s *Supplier) Run() error {
	s.Log.BeginStep("Supplying opa")

	// TODO: Install any dependencies here...

	return nil
}
