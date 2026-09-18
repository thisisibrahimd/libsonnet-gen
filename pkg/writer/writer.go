package writer

import (
	"github.com/thisisibrahimd/libsonnet-gen/pkg/config"
	"github.com/thisisibrahimd/libsonnet-gen/pkg/model"
)

type Writer interface {
	Render(dir string, group model.Groups, spec config.Target) error
}
