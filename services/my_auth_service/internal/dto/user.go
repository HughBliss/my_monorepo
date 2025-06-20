package dto

import (
	"github.com/hughbliss/my_database/pkg/gen/dbauth"
	admusrserv1 "github.com/hughbliss/my_protobuf/go/pkg/gen/admin/users/v1"
	usrserv1 "github.com/hughbliss/my_protobuf/go/pkg/gen/users/v1"
	"github.com/rs/xid"
)

type User struct {
	dbauth.User
}

func (u *User) FromEnt(e *dbauth.User) *User {
	*u = User{
		User: *e,
	}
	return u
}

type UserList []*User

func (l *UserList) FromEnt(e []*dbauth.User) UserList {
	*l = make(UserList, 0, len(e))
	for _, u := range e {
		*l = append(*l, new(User).FromEnt(u))
	}
	return *l
}

func (l *UserList) ToProto() []*admusrserv1.UserDomains {
	res := make([]*admusrserv1.UserDomains, len(*l))
	for i, u := range *l {
		res[i] = u.ToProto()
	}
	return res
}

func (u *User) ToProto() *admusrserv1.UserDomains {
	user := &usrserv1.User{
		Id:              u.ID.String(),
		Name:            u.Name,
		Email:           u.Email,
		PasswordHash:    "***",
		CurrentDomainId: u.CurrentDomainID.String(),
	}

	domainRoles := make([]*admusrserv1.DomainRole, len(u.Edges.UserDomain))
	for i, e := range u.Edges.UserDomain {
		domainRoles[i] = &admusrserv1.DomainRole{
			Domain: new(Domain).FromEnt(e.Edges.Domain).ToProto(),
			Role:   RoleFromEnt(e.Edges.Role).ToProto(),
		}
	}

	return &admusrserv1.UserDomains{
		User:        user,
		DomainRoles: domainRoles,
	}
}

func (u *User) FromProto(p *usrserv1.User) *User {
	*u = User{}

	u.ID, _ = xid.FromString(p.Id)
	u.Name = p.Name
	u.Email = p.Email
	u.CurrentDomainID, _ = xid.FromString(p.CurrentDomainId)

	return u
}
