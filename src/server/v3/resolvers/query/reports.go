package query

import (
	"context"

	"github.com/SevenTV/Common/aggregations"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/GQL/src/server/v3/helpers"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const REPORTS_QUERY_LIMIT int32 = 50

func (r *Resolver) Reports(ctx context.Context, args struct {
	Limit    *int32
	AfterID  *string
	BeforeID *string
}) ([]*ReportResolver, error) {
	// Get the actor
	actor, ok := ctx.Value(helpers.UserKey).(*structures.User)
	if !ok {
		return nil, helpers.ErrUnauthorized
	}
	if !actor.HasPermission(structures.RolePermissionManageReports) {
		return nil, helpers.ErrAccessDenied
	}

	// Define limit
	limit := int32(10)
	if args.Limit != nil {
		limit = *args.Limit
	}
	if limit > REPORTS_QUERY_LIMIT {
		limit = REPORTS_QUERY_LIMIT
	}

	// Define pipeline
	pipeline := mongo.Pipeline{}

	// Paginate
	pagination := bson.M{}
	if args.AfterID != nil && *args.AfterID != "" {
		if afterID, err := primitive.ObjectIDFromHex(*args.AfterID); err != nil {
			return nil, helpers.ErrBadObjectID
		} else {
			pagination["$gt"] = afterID
		}
	}
	if args.BeforeID != nil && *args.BeforeID != "" {
		if beforeID, err := primitive.ObjectIDFromHex(*args.BeforeID); err != nil {
			return nil, helpers.ErrBadObjectID
		} else {
			pagination["$lt"] = beforeID
		}
	}
	if len(pagination) > 0 {
		pipeline = append(pipeline, bson.D{{
			Key: "$match",
			Value: bson.M{
				"_id": pagination,
			},
		}})
	}
	pipeline = append(pipeline, bson.D{{
		Key:   "$limit",
		Value: limit,
	}})

	// Add relations
	fields := GenerateSelectedFieldMap(ctx).Children
	if _, ok := fields["reporter"]; ok {
		pipeline = append(pipeline, aggregations.ReportRelationReporter...)
	}
	if _, ok := fields["target"]; ok {
		pipeline = append(pipeline, aggregations.ReportRelationTarget...)
	}

	// Execute aggregation
	reports := make([]*structures.Report, limit)
	cur, err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameReports).Aggregate(ctx, pipeline)
	if err != nil {
		logrus.WithError(err).Error("mongo")
		return nil, err
	}
	if err = cur.All(ctx, &reports); err != nil {
		logrus.WithError(err).Error("mongo")
		return nil, err
	}

	// Generate resolvers
	resolvers := make([]*ReportResolver, len(reports))
	for i, report := range reports {
		resolver, err := CreateReportResolver(r.Ctx, ctx, report, &report.ID, fields)
		if err != nil {
			logrus.WithError(err).Error("mongo")
			return nil, err
		}
		resolvers[i] = resolver
	}

	return resolvers, nil
}
