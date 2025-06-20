package dto

import (
	"github.com/hughbliss/my_database/pkg/gen/dbauth"
	"github.com/hughbliss/my_protobuf/go/pkg/gen/common/v1"
)

type Domain struct {
	dbauth.Domain
}

func (d *Domain) FromEnt(e *dbauth.Domain) *Domain {
	*d = Domain{Domain: *e}
	return d
}

func (d *Domain) ToProto() *common.Domain {
	return &common.Domain{
		Id:   d.ID.String(),
		Name: d.Name,
	}
}
