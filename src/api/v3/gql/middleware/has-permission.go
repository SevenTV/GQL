package middleware

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/api/v3/gql/auth"
	"github.com/SevenTV/GQL/src/global"
)

func hasPermission(gCtx global.Context) func(ctx context.Context, obj interface{}, next graphql.Resolver, role []model.Permission) (res interface{}, err error) {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver, role []model.Permission) (res interface{}, err error) {
		user := auth.For(ctx)
		if user == nil {
			return nil, errors.ErrUnauthorized()
		}

		var perms structures.RolePermission
		for _, v := range role {
			switch v {
			case model.PermissionBypassPrivacy:
				perms |= structures.RolePermissionBypassPrivacy
			case model.PermissionEmotesetCreate:
				perms |= structures.RolePermissionCreateEmoteSet
			case model.PermissionEmotesetEdit:
				perms |= structures.RolePermissionEditEmoteSet
			case model.PermissionEmoteCreate:
				perms |= structures.RolePermissionCreateEmote
			case model.PermissionEmoteEditAny:
				perms |= structures.RolePermissionEditAnyEmote
			case model.PermissionEmotesetEditAny:
				perms |= structures.RolePermissionEditAnyEmoteSet
			case model.PermissionEmoteEdit:
				perms |= structures.RolePermissionEditEmote
			case model.PermissionFeatureProfilePictureAnimation:
				perms |= structures.RolePermissionFeatureProfilePictureAnimation
			case model.PermissionFeatureZerowidthEmoteType:
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
			return nil, errors.ErrUnauthorized()
		}

		return next(ctx)
	}
}
