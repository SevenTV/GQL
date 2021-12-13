package middleware

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/global"
	"github.com/SevenTV/GQL/src/server/v3/gql/auth"
	"github.com/SevenTV/GQL/src/server/v3/gql/errors"
)

func hasPermission(gCtx global.Context) func(ctx context.Context, obj interface{}, next graphql.Resolver, role []model.Permission) (res interface{}, err error) {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver, role []model.Permission) (res interface{}, err error) {
		user := auth.For(ctx)
		if user == nil {
			return nil, errors.ErrLoginRequired
		}

		var perms structures.RolePermission
		for _, v := range role {
			switch v {
			case model.PermissionBypassPrivacy:
				perms |= structures.RolePermissionBypassPrivacy
			case model.PermissionSetChannelEmote:
				perms |= structures.RolePermissionSetChannelEmote
			case model.PermissionCreateEmote:
				perms |= structures.RolePermissionCreateEmote
			case model.PermissionEditAnyEmote:
				perms |= structures.RolePermissionEditAnyEmote
			case model.PermissionEditAnyEmoteSet:
				perms |= structures.RolePermissionEditAnyEmoteSet
			case model.PermissionEditEmote:
				perms |= structures.RolePermissionEditEmote
			case model.PermissionFeatureProfilePictureAnimation:
				perms |= structures.RolePermissionFeatureProfilePictureAnimation
			case model.PermissionFeatureZeroWidthEmoteType:
				perms |= structures.RolePermissionFeatureZeroWidthEmoteType
			case model.PermissionManageBans:
				perms |= structures.RolePermissionManageBans
			case model.PermissionManageCosmetics:
				perms |= structures.RolePermissionManageCosmetics
			case model.PermissionManageNews:
				perms |= structures.RolePermissionManageNews
			case model.PermissionManageReports:
				perms |= structures.RolePermissionManageReports
			case model.PermissionManageRoles:
				perms |= structures.RolePermissionManageRoles
			case model.PermissionManageStack:
				perms |= structures.RolePermissionManageStack
			case model.PermissionManageUsers:
				perms |= structures.RolePermissionManageUsers
			case model.PermissionReportCreate:
				perms |= structures.RolePermissionReportCreate
			case model.PermissionSendMessages:
				perms |= structures.RolePermissionSendMessages
			case model.PermissionSuperAdministrator:
				perms |= structures.RolePermissionSuperAdministrator
			}
		}

		if !user.HasPermission(perms) {
			return nil, errors.ErrAccessDenied
		}

		return next(ctx)
	}
}
