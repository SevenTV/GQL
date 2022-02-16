package middleware

import (
	"strings"
	"time"

	"github.com/SevenTV/Common/auth"
	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/structures/v3/aggregations"
	"github.com/SevenTV/Common/structures/v3/query"
	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/src/global"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Auth(gCtx global.Context) Middleware {
	return func(ctx *fasthttp.RequestCtx) errors.APIError {
		// Parse token from header
		h := utils.B2S(ctx.Request.Header.Peek("Authorization"))
		if len(h) == 0 {
			return nil
		}

		s := strings.Split(h, "Bearer ")
		if len(s) != 2 {
			return errors.ErrUnauthorized().SetDetail("Bad Authorization Header")
		}
		t := s[1]

		// Verify the token
		_, claims, err := auth.VerifyJWT(gCtx.Config().Credentials.JWTSecret, strings.Split(t, "."))
		if err != nil {
			return nil
		}

		// User ID from parsed token
		u := claims["u"]
		if u == nil {
			return errors.ErrUnauthorized().SetDetail("Bad Token")
		}
		userID, err := primitive.ObjectIDFromHex(u.(string))
		if err != nil {
			return errors.ErrUnauthorized().SetDetail(err.Error())
		}

		// Version of parsed token
		user := &structures.User{}
		v := claims["v"].(float64)

		pipeline := mongo.Pipeline{{{Key: "$match", Value: bson.M{"_id": userID}}}}
		pipeline = append(pipeline, aggregations.UserRelationRoles...)
		pipeline = append(pipeline, aggregations.UserRelationBans...)
		cur, err := gCtx.Inst().Mongo.Collection(mongo.CollectionNameUsers).Aggregate(ctx, pipeline)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return errors.ErrUnauthorized().SetDetail("Token has Unknown Bound User")
			}

			logrus.WithError(err).Error("mongo")
			return errors.ErrInternalServerError()
		}
		cur.Next(ctx)
		err = cur.Decode(user)
		if err != nil {
			return errors.ErrInternalServerError().SetDetail(err.Error())
		}
		err = cur.Close(ctx)
		if err != nil {
			return errors.ErrInternalServerError().SetDetail(err.Error())
		}

		// Check bans
		for _, ban := range user.Bans {
			// Check for No Auth effect
			if ban.HasEffect(structures.BanEffectNoAuth) {
				return errors.ErrInsufficientPrivilege().SetDetail("You are banned!").SetFields(errors.Fields{
					"ban": map[string]string{
						"reason":    ban.Reason,
						"expire_at": ban.ExpireAt.Format(time.RFC3339),
					},
				})
			}
			// Check for No Permissions effect
			if ban.HasEffect(structures.BanEffectNoPermissions) {
				user.Roles = []*structures.Role{structures.RevocationRole}

			}
		}
		defaultRoles := query.DefaultRoles.Fetch(ctx, gCtx.Inst().Mongo, gCtx.Inst().Redis)
		user.AddRoles(defaultRoles...)

		if user.TokenVersion != v {
			return errors.ErrUnauthorized().SetDetail("Token Version Mismatch")
		}

		ctx.SetUserValue("user", user)
		return nil
	}
}
