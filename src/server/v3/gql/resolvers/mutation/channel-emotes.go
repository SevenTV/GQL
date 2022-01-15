package mutation

import (
	"context"

	"github.com/SevenTV/GQL/graph/model"
)

func (r *Resolver) SetChannelEmote(ctx context.Context, userID string, target model.UserEmoteInput, action model.ChannelEmoteListItemAction) (*model.User, error) {
	/*
		user := auth.For(ctx)

		var targetUser *structures.User
		targetUserID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			return nil, helpers.ErrBadObjectID
		}

		var targetEmote *structures.Emote
		targetEmoteID, err := primitive.ObjectIDFromHex(target.ID)
		if err != nil {
			return nil, helpers.ErrBadObjectID
		}

		connMap := map[primitive.ObjectID]bool{}
		for _, v := range target.Channels {
			id, err := primitive.ObjectIDFromHex(v)
			if err != nil {
				return nil, helpers.ErrBadObjectID
			}
			connMap[id] = true
		}

		connectionIDs := make([]primitive.ObjectID, len(connMap))
		i := 0
		for k := range connMap {
			connectionIDs[i] = k
			i++
		}

		if targetUserID != user.ID {
			cur, err := r.Ctx.Inst().Mongo.Collection(structures.CollectionNameUsers).Aggregate(ctx, append(mongo.Pipeline{
				{{Key: "$match", Value: bson.M{"_id": targetUserID}}},
			}, aggregations.UserRelationRoles...))
			if err != nil {
				if err == mongo.ErrNoDocuments {
					return nil, helpers.ErrUnknownUser
				}

				logrus.WithError(err).Error("failed to run mongo pipeline")
				return nil, helpers.ErrInternalServerError
			}

			cur.Next(ctx)
			err = multierror.Append(cur.Decode(targetUser), cur.Close(ctx)).ErrorOrNil()
			if err != nil {
				logrus.WithError(err).Error("failed to decode and close mongo cursor")
				return nil, err
			}

			if !user.HasPermission(structures.RolePermissionManageUsers) {
				for _, v := range targetUser.Editors {
					if v.ID == user.ID {
						if v.HasPermission(structures.UserEditorPermissionModifyChannelEmotes) {
							break
						} else {
							return nil, helpers.ErrAccessDenied
						}
					}
				}
			}
		} else {
			targetUser = user
		}

		switch action {
		case model.ChannelEmoteListItemActionAdd:
			// if they are adding an emote we need to make sure the emote is not already added.
			for _, v := range targetUser.ChannelEmotes {
				if v.ID == targetEmoteID {
					return nil, helpers.ErrDontBeSilly
				}
			}
		case model.ChannelEmoteListItemActionRemove, model.ChannelEmoteListItemActionUpdate:
			// if they are removing/updating an emote we need to make sure the emote is added.
			found := false
			for _, v := range targetUser.ChannelEmotes {
				if v.ID == targetEmoteID {
					found = true
					break
				}
			}
			if !found {
				return nil, helpers.ErrUnknownEmote
			}
		}
			// we cannot add the emote to the channel if its already added to the channel
			for _, v := range connectionIDs {
				found := false
				for _, c := range targetUser.Connections {
					if c.ID == v {
						found = true
						break
					}
				}
				if !found {
					return nil, helpers.ErrUnknownUser
				}
			}
		opts := aggregations.UserRelationshipOptions{}

		// if they are adding an emote we need to make sure they are allowed to add this emote
		if action == model.ChannelEmoteListItemActionAdd {
			// at this point we have permission to add emotes to the user, but do we have permission to add that particular emote
			if !user.HasPermission(structures.RolePermissionManageUsers) && !targetUser.HasPermission(structures.RolePermissionManageUsers) {
				opts.Editors = true
			}
		}

		cur, err := r.Ctx.Inst().Mongo.Collection(structures.CollectionNameEmotes).Aggregate(ctx, append(mongo.Pipeline{
			{{Key: "$match", Value: bson.M{"_id": targetEmoteID}}},
		}, aggregations.GetEmoteRelationshipOwner(opts)...))
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, helpers.ErrUnknownEmote
			}

			logrus.WithError(err).Error("failed to run mongo pipeline")
			return nil, helpers.ErrInternalServerError
		}

		cur.Next(ctx)
		err = multierror.Append(cur.Decode(targetEmote), cur.Close(ctx)).ErrorOrNil()
		if err != nil {
			logrus.WithError(err).Error("failed to decode and close mongo cursor")
			return nil, err
		}

		// if they are adding an emote we need to make sure they are allowed to add this emote
		if action == model.ChannelEmoteListItemActionAdd {
			// private emote check if we have permission to add this emote
			if targetEmote.Flags&structures.EmoteFlagsPrivate != 0 && !user.HasPermission(structures.RolePermissionManageUsers) && !targetUser.HasPermission(structures.RolePermissionManageUsers) {
				for _, v := range targetEmote.Owner.Editors {
					if v.ID == targetUserID {
						if v.HasPermission(structures.UserEditorPermissionUsePrivateEmotes) {
							break
						} else {
							return nil, helpers.ErrAccessDenied
						}
					}
				}
			}

			// zero width emote check if target user has that permission
			if targetEmote.Flags&structures.EmoteFlagsZeroWidth != 0 && !user.HasPermission(structures.RolePermissionManageUsers) && !targetUser.HasPermission(structures.RolePermissionManageUsers) && !targetUser.HasPermission(structures.RolePermissionFeatureZeroWidthEmoteType) {
				return nil, helpers.ErrAccessDenied
			}
		}

		tagColor := 0
		if len(targetUser.Roles) != 0 {
			tagColor = int(targetUser.Roles[0].Color)
		}

		// at this point we can assume that the emote can be modified on the channel since all other checks passed.
		var res *mongoRaw.SingleResult
		switch action {
		case model.ChannelEmoteListItemActionAdd:
			res = r.Ctx.Inst().Mongo.Collection(structures.CollectionNameUsers).FindOneAndUpdate(ctx, bson.M{"_id": targetUserID}, bson.M{
				"$addToSet": bson.M{
					"channel_emotes": structures.UserEmote{
						ID:          targetEmoteID,
						Alias:       target.Alias,
						Connections: connectionIDs,
					},
				},
			})
		case model.ChannelEmoteListItemActionRemove:
			res = r.Ctx.Inst().Mongo.Collection(structures.CollectionNameUsers).FindOneAndUpdate(ctx, bson.M{"_id": targetUserID}, bson.M{
				"$pull": bson.M{
					"channel_emotes": bson.M{
						"id": targetEmoteID,
					},
				},
			})
		case model.ChannelEmoteListItemActionUpdate:
			res = r.Ctx.Inst().Mongo.Collection(structures.CollectionNameUsers).FindOneAndUpdate(ctx, bson.M{"_id": targetUserID, "channel_emotes": bson.M{"id": targetEmoteID}}, bson.M{
				"channel_emotes.$": structures.UserEmote{
					ID:          targetEmoteID,
					Alias:       target.Alias,
					Connections: connectionIDs,
				},
			})
		}
		err = res.Err()
		if err == nil {
			err = res.Decode(targetUser)
		}
		if err != nil {
			logrus.WithError(err).Error("failed to update user")
			return nil, err
		}

		channelEmotes := make([]*model.UserEmote, len(targetUser.ChannelEmotes))
		for i, v := range targetUser.ChannelEmotes {
			connections := make([]string, len(v.Connections))
			for i, v := range v.Connections {
				connections[i] = v.Hex()
			}

			channelEmotes[i] = &model.UserEmote{
				Alias:       v.Alias,
				AddedAt:     v.AddedAt,
				Connections: connections,
			}
		}

		editors := make([]*model.UserEditor, len(targetUser.Editors))
		for i, v := range targetUser.Editors {
			connections := make([]string, len(v.Connections))
			for i, v := range v.Connections {
				connections[i] = v.Hex()
			}

			editors[i] = &model.UserEditor{
				Permissions: int(v.Permissions),
				Visible:     v.Visible,
				AddedAt:     v.AddedAt,
				Connections: connections,
			}
		}

		roles := make([]*model.Role, len(targetUser.RoleIDs))
		for i, v := range targetUser.RoleIDs {
			roles[i] = &model.Role{
				ID: v.Hex(),
			}
		}

		connections := make([]*model.UserConnection, len(targetUser.Connections))
		for i, v := range targetUser.Connections {
			connections[i] = &model.UserConnection{
				ID: v.ID,
			}
		}

		return &model.User{
			ID:            targetUser.ID.Hex(),
			UserType:      string(targetUser.UserType),
			Username:      targetUser.Username,
			DisplayName:   targetUser.DisplayName,
			CreatedAt:     targetUser.ID.Timestamp(),
			AvatarURL:     targetUser.AvatarURL,
			Biography:     targetUser.Biography,
			TagColor:      tagColor,
			Editors:       editors,
			Roles:         roles,
			Connections:   connections,
			ChannelEmotes: channelEmotes,
		}, nil
	*/
	return nil, nil
}
