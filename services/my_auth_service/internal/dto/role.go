package dto

import (
	"github.com/hughbliss/my_database/pkg/gen/dbauth"
	"github.com/hughbliss/my_protobuf/go/pkg/gen/common/v1"
	roleserv1 "github.com/hughbliss/my_protobuf/go/pkg/gen/roles/v1"
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

func (r DomainRoles) ToProto() *roleserv1.DomainRoles {
	roles := make([]*roleserv1.Role, len(r.Roles))
	for i, role := range r.Roles {
		roles[i] = role.ToProto()
	}
	return &roleserv1.DomainRoles{
		Domain: &common.Domain{
			Id:   r.DomainID,
			Name: r.DomainName,
		},
		Roles: roles,
	}
}

type Role struct {
	Name        string
	Description string
	Permissions []string
	DomainId    string
}

func RoleFromEnt(e *dbauth.Role) *Role {
	return &Role{
		Name:        e.Name,
		Description: e.Description,
		Permissions: e.Permissions,
		DomainId:    e.DomainID.String(),
	}
}

func (r *Role) ToProto() *roleserv1.Role {
	return &roleserv1.Role{
		Name:        r.Name,
		Description: r.Description,
		Permissions: r.Permissions,
		DomainId:    r.DomainId,
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

func (l DomainRolesList) ToProto() []*roleserv1.DomainRoles {
	roles := make([]*roleserv1.DomainRoles, len(l))
	for i, role := range l {
		roles[i] = role.ToProto()
	}
	return roles

}
