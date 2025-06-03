package main

import (
	"encoding/json"
	"errors"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

type Options struct {
	Package string `json:"package"`
}

func ParseOptions(req *plugin.GenerateRequest) (*Options, error) {
	var options Options
	if err := json.Unmarshal(req.GetPluginOptions(), &options); err != nil {
		return nil, err
	}
	return &options, nil
}

func ValidateOptions(opts *Options) error {
	if opts.Package == "" {
		return errors.New(`options: "package" is required`)
	}
	return nil
}
