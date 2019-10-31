package main

import (
	"bytes"
	"context"
	"flag"
	"log"
	"strconv"
	"strings"

	"github.com/docker/distribution/configuration"
	"github.com/docker/distribution/registry/storage"
	"github.com/docker/distribution/registry/storage/driver/factory"
	_ "github.com/docker/distribution/registry/storage/driver/filesystem"
	"github.com/docker/libtrust"
)

var defaultConfig = `version: 0.1
storage:
  filesystem:
    rootdirectory: ${rootdirectory}
`

var (
	storageRoot   = "/var/lib/registry"
	enableSchema1 bool
)

func init() {
	flag.StringVar(&storageRoot, "r", storageRoot, "storage root directory")
	flag.BoolVar(&enableSchema1, "s", enableSchema1, "enable schema1")
}

func main() {
	flag.Parse()

	config, err := resolveConfiguration(storageRoot, enableSchema1)
	if err != nil {
		log.Fatal(err)
	}

	driver, err := factory.Create(config.Storage.Type(), config.Storage.Parameters())
	if err != nil {
		log.Fatal(err)
	}

	key, err := libtrust.GenerateECP256PrivateKey()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	registry, err := storage.NewRegistry(ctx, driver, storage.Schema1SigningKey(key))
	if err != nil {
		log.Fatal(err)
	}

	err = storage.MarkAndSweep(ctx, driver, registry, storage.GCOpts{
		DryRun:         false,
		RemoveUntagged: true,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func resolveConfiguration(storageRoot string, enableSchema1 bool) (*configuration.Configuration, error) {
	if storageRoot == "" {
		storageRoot = ""
	}

	c := defaultConfig
	c = strings.Replace(c, "${rootdirectory}", storageRoot, -1)
	c = strings.Replace(c, "${schema1}", strconv.FormatBool(enableSchema1), -1)

	buff := bytes.NewBufferString(c)
	return configuration.Parse(buff)
}
