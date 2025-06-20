package dto

import (
	"github.com/hughbliss/my_database/pkg/gen/dbauth"
	admrolserv1 "github.com/hughbliss/my_protobuf/go/pkg/gen/admin/roles/v1"
	"github.com/hughbliss/my_protobuf/go/pkg/gen/common/v1"
	rolserv1 "github.com/hughbliss/my_protobuf/go/pkg/gen/roles/v1"
	"github.com/rs/xid"
)

type DomainRoles struct {
	DomainID   string
	DomainName string
	Roles      []*Role
}

func DomainRolesFromEnt(e *dbauth.Domain) *DomainRoles {
	roles := make([]*Role, len(e.Edges.Roles))
	for i, role := range e.Edges.Roles {
		roles[i] = RoleFromEnt(role)
	}

	return &DomainRoles{
		DomainID:   e.ID.String(),
		DomainName: e.Name,
		Roles:      roles,
	}
}

func (r DomainRoles) ToProto() *admrolserv1.DomainRoles {
	roles := make([]*rolserv1.Role, len(r.Roles))
	for i, role := range r.Roles {
		roles[i] = role.ToProto()
	}
	return &admrolserv1.DomainRoles{
		Domain: &common.Domain{
			Id:   r.DomainID,
			Name: r.DomainName,
		},
		Roles: roles,
	}
}

type Role struct {
	ID          xid.ID
	Name        string
	Description string
	Permissions []string
	DomainId    xid.ID
}

func RoleFromProto(p *rolserv1.Role) *Role {
	id, _ := xid.FromString(p.Id)
	domainId, _ := xid.FromString(p.DomainId)
	return &Role{
		ID:          id,
		Name:        p.Name,
		Description: p.Description,
		Permissions: p.Permissions,
		DomainId:    domainId,
	}
}

func RoleFromEnt(e *dbauth.Role) *Role {
	return &Role{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		Permissions: e.Permissions,
		DomainId:    e.DomainID,
	}
}

func (r *Role) ToProto() *rolserv1.Role {
	return &rolserv1.Role{
		Id:          r.ID.String(),
		Name:        r.Name,
		Description: r.Description,
		Permissions: r.Permissions,
		DomainId:    r.DomainId.String(),
	}
}

type DomainRolesList []*DomainRoles

func DomainRolesListFromEnt(e []*dbauth.Domain) DomainRolesList {
	if len(e) == 0 {
		return nil
	}
	res := make(DomainRolesList, len(e))
	for i, domain := range e {
		res[i] = DomainRolesFromEnt(domain)
	}
	return res
}

func (l DomainRolesList) ToProto() []*admrolserv1.DomainRoles {
	roles := make([]*admrolserv1.DomainRoles, len(l))
	for i, role := range l {
		roles[i] = role.ToProto()
	}
	return roles

}
