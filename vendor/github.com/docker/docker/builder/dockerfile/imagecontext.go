package dockerfile

import (
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/builder"
	"github.com/docker/docker/builder/remotecontext"
	"github.com/pkg/errors"
)

type pathCache interface {
	Load(key interface{}) (value interface{}, ok bool)
	Store(key, value interface{})
}

// imageContexts is a helper for stacking up built image rootfs and reusing
// them as contexts
type imageContexts struct {
	b      *Builder
	list   []*imageMount
	byName map[string]*imageMount
	cache  pathCache
}

func (ic *imageContexts) newImageMount(id string) *imageMount {
	return &imageMount{ic: ic, id: id}
}

func (ic *imageContexts) add(name string) (*imageMount, error) {
	im := &imageMount{ic: ic}
	if len(name) > 0 {
		if ic.byName == nil {
			ic.byName = make(map[string]*imageMount)
		}
		if _, ok := ic.byName[name]; ok {
			return nil, errors.Errorf("duplicate name %s", name)
		}
		ic.byName[name] = im
	}
	ic.list = append(ic.list, im)
	return im, nil
}

func (ic *imageContexts) update(imageID string, runConfig *container.Config) {
	ic.list[len(ic.list)-1].id = imageID
	ic.list[len(ic.list)-1].runConfig = runConfig
}

func (ic *imageContexts) validate(i int) error {
	if i < 0 || i >= len(ic.list)-1 {
		var extraMsg string
		if i == len(ic.list)-1 {
			extraMsg = " refers current build block"
		}
		return errors.Errorf("invalid from flag value %d%s", i, extraMsg)
	}
	return nil
}

func (ic *imageContexts) get(indexOrName string) (*imageMount, error) {
	index, err := strconv.Atoi(indexOrName)
	if err == nil {
		if err := ic.validate(index); err != nil {
			return nil, err
		}
		return ic.list[index], nil
	}
	if im, ok := ic.byName[strings.ToLower(indexOrName)]; ok {
		return im, nil
	}
	im, err := mountByRef(ic.b, indexOrName)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid from flag value %s", indexOrName)
	}
	return im, nil
}

func (ic *imageContexts) unmount() (retErr error) {
	for _, im := range ic.list {
		if err := im.unmount(); err != nil {
			logrus.Error(err)
			retErr = err
		}
	}
	for _, im := range ic.byName {
		if err := im.unmount(); err != nil {
			logrus.Error(err)
			retErr = err
		}
	}
	return
}

func (ic *imageContexts) getCache(id, path string) (interface{}, bool) {
	if ic.cache != nil {
		if id == "" {
			return nil, false
		}
		return ic.cache.Load(id + path)
	}
	return nil, false
}

func (ic *imageContexts) setCache(id, path string, v interface{}) {
	if ic.cache != nil {
		ic.cache.Store(id+path, v)
	}
}

// imageMount is a reference for getting access to a buildcontext that is backed
// by an existing image
type imageMount struct {
	id        string
	source    builder.Source
	release   func() error
	ic        *imageContexts
	runConfig *container.Config
}

func (im *imageMount) context() (builder.Source, error) {
	if im.source == nil {
		if im.id == "" {
			return nil, errors.Errorf("could not copy from empty context")
		}
		p, release, err := im.ic.b.docker.MountImage(im.id)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to mount %s", im.id)
		}
		source, err := remotecontext.NewLazyContext(p)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create lazycontext for %s", p)
		}
		im.release = release
		im.source = source
	}
	return im.source, nil
}

func (im *imageMount) unmount() error {
	if im.release != nil {
		if err := im.release(); err != nil {
			return errors.Wrapf(err, "failed to unmount previous build image %s", im.id)
		}
		im.release = nil
	}
	return nil
}

func (im *imageMount) ImageID() string {
	return im.id
}
func (im *imageMount) RunConfig() *container.Config {
	return im.runConfig
}
