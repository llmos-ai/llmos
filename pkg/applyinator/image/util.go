package image

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/rancher/wharfie/pkg/credentialprovider/plugin"
	"github.com/rancher/wharfie/pkg/registries"
	"github.com/rancher/wharfie/pkg/tarfile"
	"github.com/sirupsen/logrus"
)

const (
	baseLLMOSDir                         string = "/var/lib/llmos/"
	defaultRegistriesFile                string = "/etc/llmos/registries.yaml"
	defaultImagesDir                            = baseLLMOSDir + "images"
	defaultImageCredentialProviderConfig        = baseLLMOSDir + "credentialprovider/config.yaml"
	defaultImageCredentialProviderBinDir        = baseLLMOSDir + "credentialprovider/bin"

	rke2RegistriesFile string = "/etc/rancher/rke2/registries.yaml"
	k3sRegistriesFile  string = "/etc/rancher/k3s/registries.yaml"
)

type Utility struct {
	ImagesDir                     string
	ImageCredentialProviderConfig string
	ImageCredentialProviderBinDir string
	AgentRegistriesFile           string
}

func NewUtility(u *Utility) *Utility {
	if u == nil {
		u = &Utility{}
	}

	if u.ImagesDir == "" {
		u.ImagesDir = defaultImagesDir
	}

	if u.ImageCredentialProviderConfig == "" {
		u.ImageCredentialProviderConfig = defaultImageCredentialProviderConfig
	}

	if u.ImageCredentialProviderBinDir == "" {
		u.ImageCredentialProviderBinDir = defaultImageCredentialProviderBinDir
	}

	if u.AgentRegistriesFile == "" {
		u.AgentRegistriesFile = defaultRegistriesFile
	}

	logrus.Debugf("Instantiated new image utility with imagesDir: %s, imageCredentialProviderConfig: %s, "+
		"imageCredentialProviderBinDir: %s, agentRegistriesFile: %s",
		u.ImagesDir, u.ImageCredentialProviderConfig, u.ImageCredentialProviderBinDir, u.AgentRegistriesFile)

	return u
}

func (u *Utility) Stage(destDir string, imgString string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	var img v1.Image
	image, err := name.ParseReference(imgString)
	if err != nil {
		return err
	}
	logrus.Debugf("Pulling image info %+v", image)

	imagesDir, err := filepath.Abs(u.ImagesDir)
	if err != nil {
		return err
	}

	i, err := tarfile.FindImage(imagesDir, image)
	if err != nil && !errors.Is(err, tarfile.ErrNotFound) {
		return err
	}
	img = i

	if img == nil {
		registry, err := registries.GetPrivateRegistries(u.findRegistriesYaml())
		if err != nil {
			return err
		}

		if _, err = os.Stat(u.ImageCredentialProviderConfig); os.IsExist(err) {
			logrus.Debugf("Image Credential Provider Configuration file %s existed, using plugins from directory %s",
				u.ImageCredentialProviderConfig, u.ImageCredentialProviderBinDir)
			plugins, err := plugin.RegisterCredentialProviderPlugins(u.ImageCredentialProviderConfig, u.ImageCredentialProviderBinDir)
			if err != nil {
				return err
			}
			registry.DefaultKeychain = plugins
		} else {
			// The kubelet image credential provider plugin also falls back to checking legacy Docker credentials, so only
			// explicitly set up the go-containerregistry DefaultKeychain if plugins are not configured.
			// DefaultKeychain tries to read config from the home dir, and will error if HOME isn't set, so also gate on that.
			if os.Getenv("HOME") != "" {
				registry.DefaultKeychain = authn.DefaultKeychain
			}
		}

		logrus.Infof("Pulling image %s", image.Name())
		img, err = registry.Image(image,
			remote.WithPlatform(v1.Platform{
				Architecture: runtime.GOARCH,
				OS:           runtime.GOOS,
			}),
		)
		if err != nil {
			return fmt.Errorf("%v: failed to get image %s", err, image.Name())
		}
		logrus.Debugf("[debug] Pulled image manifests %+v", img)
		digest, err := img.Digest()
		if err != nil {
			return err
		}
		logrus.Debugf("[debug] Pulled image digest %v", digest)
	}

	logrus.Debugf("[debug] Extracting image %s to %s. image: %+v", image.Name(), destDir, image)
	return extractFiles(img, destDir)
}

func (u *Utility) findRegistriesYaml() string {
	if _, err := os.Stat(u.AgentRegistriesFile); err == nil {
		return u.AgentRegistriesFile
	}
	if _, err := os.Stat(rke2RegistriesFile); err == nil {
		return rke2RegistriesFile
	}
	if _, err := os.Stat(k3sRegistriesFile); err == nil {
		return k3sRegistriesFile
	}
	return ""
}
