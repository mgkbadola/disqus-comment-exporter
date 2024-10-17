package cmd

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/mgkbadola/disqus-comment-exporter/api"
	"io"
	"os"
	"strings"
)

type CommonOptionsCommander interface {
	Execute(args []string) error
}

type ExportCommand struct {
	ConfigFile string `short:"c" long:"config" description:"Config file name" required:"true"`
}

func (ec *ExportCommand) Execute(_ []string) error {
	reader, err := ec.reader(ec.ConfigFile)
	if err != nil {
		return fmt.Errorf("can't open config file %s: %w", ec.ConfigFile, err)
	}
	config, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("can't read config file %s: %w", ec.ConfigFile, err)
	}
	r := strings.NewReader(string(config))
	decoder := json.NewDecoder(r)
	var configObject api.Config
	err = decoder.Decode(&configObject)
	if err != nil {
		return fmt.Errorf("can't de-serialize config file %s: %w", ec.ConfigFile, err)
	}
	d := api.NewDisqusAPIWrapper(configObject)
	d.BeginCommentExport()
	return nil
}

func (ec *ExportCommand) reader(inp string) (reader io.Reader, err error) {
	inpFile, err := os.Open(inp) // nolint
	if err != nil {
		return nil, fmt.Errorf("import failed, can't open %s: %w", inp, err)
	}

	reader = inpFile
	if strings.HasSuffix(ec.ConfigFile, ".gz") {
		if reader, err = gzip.NewReader(inpFile); err != nil {
			return nil, fmt.Errorf("can't make gz reader: %w", err)
		}
	}
	return reader, nil
}
