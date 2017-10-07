// +build windows

package distribution

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"

	"github.com/docker/distribution"
	"github.com/docker/distribution/context"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/registry/client/transport"
	"github.com/docker/docker/pkg/system"
	"github.com/sirupsen/logrus"
)

var _ distribution.Describable = &v2LayerDescriptor{}

func (ld *v2LayerDescriptor) Descriptor() distribution.Descriptor {
	if ld.src.MediaType == schema2.MediaTypeForeignLayer && len(ld.src.URLs) > 0 {
		return ld.src
	}
	return distribution.Descriptor{}
}

func (ld *v2LayerDescriptor) open(ctx context.Context) (distribution.ReadSeekCloser, error) {
	blobs := ld.repo.Blobs(ctx)
	rsc, err := blobs.Open(ctx, ld.digest)

	if len(ld.src.URLs) == 0 {
		return rsc, err
	}

	// We're done if the registry has this blob.
	if err == nil {
		// Seek does an HTTP GET.  If it succeeds, the blob really is accessible.
		if _, err = rsc.Seek(0, os.SEEK_SET); err == nil {
			return rsc, nil
		}
		rsc.Close()
	}

	// Find the first URL that results in a 200 result code.
	for _, url := range ld.src.URLs {
		logrus.Debugf("Pulling %v from foreign URL %v", ld.digest, url)
		rsc = transport.NewHTTPReadSeeker(http.DefaultClient, url, nil)

		// Seek does an HTTP GET.  If it succeeds, the blob really is accessible.
		_, err = rsc.Seek(0, os.SEEK_SET)
		if err == nil {
			break
		}
		logrus.Debugf("Download for %v failed: %v", ld.digest, err)
		rsc.Close()
		rsc = nil
	}
	return rsc, err
}

func filterManifests(manifests []manifestlist.ManifestDescriptor, os string) []manifestlist.ManifestDescriptor {
	osVersion := ""
	if os == "windows" {
		version := system.GetOSVersion()
		osVersion = fmt.Sprintf("%d.%d.%d", version.MajorVersion, version.MinorVersion, version.Build)
		logrus.Debugf("will only match entries with version %s", osVersion)
	}

	var matches []manifestlist.ManifestDescriptor
	for _, manifestDescriptor := range manifests {
		if manifestDescriptor.Platform.Architecture == runtime.GOARCH && manifestDescriptor.Platform.OS == os {
			if os == "windows" && !versionMatch(manifestDescriptor.Platform.OSVersion, osVersion) {
				logrus.Debugf("skipping %s", manifestDescriptor.Platform.OSVersion)
				continue
			}
			matches = append(matches, manifestDescriptor)
			logrus.Debugf("found match for %s/%s with media type %s, digest %s", os, runtime.GOARCH, manifestDescriptor.MediaType, manifestDescriptor.Digest.String())
		}
	}
	if os == "windows" {
		sort.Stable(manifestsByVersion(matches))
	}
	return matches
}

func versionMatch(actual, expected string) bool {
	// Check whether actual and expected are equivalent, or whether
	// expected is a version prefix of actual.
	return actual == "" || expected == "" || actual == expected || strings.HasPrefix(actual, expected+".")
}

type manifestsByVersion []manifestlist.ManifestDescriptor

func (mbv manifestsByVersion) Less(i, j int) bool {
	if mbv[i].Platform.OSVersion == "" {
		return false
	}
	if mbv[j].Platform.OSVersion == "" {
		return true
	}
	// TODO: Split version by parts and compare
	// TODO: Prefer versions which have a greater version number
	return false
}

func (mbv manifestsByVersion) Len() int {
	return len(mbv)
}

func (mbv manifestsByVersion) Swap(i, j int) {
	mbv[i], mbv[j] = mbv[j], mbv[i]
}
