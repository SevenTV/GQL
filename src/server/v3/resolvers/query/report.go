package query

import (
	"context"
	"fmt"
	"time"

	"github.com/SevenTV/Common/aggregations"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/GQL/src/global"
	"github.com/SevenTV/GQL/src/server/v3/helpers"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReportResolver struct {
	ctx context.Context
	*structures.ReportBuilder
	fields map[string]*SelectedField
	gCtx   global.Context
}

// CreateReportResolver: generate a GQL resolver for a report
func CreateReportResolver(gCtx global.Context, ctx context.Context, report *structures.Report, reportID *primitive.ObjectID, fields map[string]*SelectedField) (*ReportResolver, error) {
	rb := structures.NewReportBuilder(report)

	var pipeline mongo.Pipeline
	if rb.Report == nil && reportID == nil {
		return nil, fmt.Errorf("unresolvable")
	}
	if rb.Report == nil {
		rb.Report = &structures.Report{}
		pipeline = mongo.Pipeline{
			{{
				Key:   "$match",
				Value: bson.M{"_id": reportID},
			}},
		}
	} else {
		pipeline = mongo.Pipeline{
			{{
				Key:   "$replaceRoot",
				Value: bson.M{"newRoot": report},
			}},
		}
	}

	// Relation: target
	if _, ok := fields["target"]; ok {
		pipeline = append(pipeline, aggregations.ReportRelationTarget...)
	}

	// Relation: reporter
	if _, ok := fields["reporter"]; ok {
		pipeline = append(pipeline, aggregations.ReportRelationReporter...)
	}

	// Relation: assignees
	if _, ok := fields["assignees"]; ok {
		pipeline = append(pipeline, aggregations.ReportRelationAssignees...)
	}

	if rb.Report.ID.IsZero() || len(pipeline) > 1 {
		cur, err := gCtx.Inst().Mongo.Collection(mongo.CollectionNameReports).Aggregate(ctx, pipeline)
		if err != nil {
			return nil, err
		}
		cur.Next(ctx)
		cur.Close(ctx)
		if err = cur.Decode(rb.Report); err != nil {
			logrus.WithError(err).Error("mongo")
		}
	}

	return &ReportResolver{
		ctx:           ctx,
		ReportBuilder: rb,
		fields:        fields,
		gCtx:          gCtx,
	}, nil
}

func (r *Resolver) Report(ctx context.Context, args struct {
	ID string
}) (*ReportResolver, error) {
	actor, ok := ctx.Value(helpers.UserKey).(*structures.User)
	if !ok {
		return nil, helpers.ErrUnauthorized
	}
	if !actor.HasPermission(structures.RolePermissionManageReports) {
		return nil, helpers.ErrAccessDenied
	}

	reportID, err := primitive.ObjectIDFromHex(args.ID)
	if err != nil {
		return nil, helpers.ErrBadObjectID
	}

	return CreateReportResolver(r.Ctx, ctx, nil, &reportID, GenerateSelectedFieldMap(ctx).Children)
}

func (r *ReportResolver) ID() string {
	return r.Report.ID.Hex()
}

func (r *ReportResolver) TargetKind() string {
	return string(r.Report.TargetKind)
}

func (r *ReportResolver) Target(ctx context.Context) (*UserResolver, error) {
	if r.Report.TargetID.IsZero() {
		return nil, nil
	}

	return CreateUserResolver(r.gCtx, ctx, r.Report.Target, &r.Report.TargetID, GenerateSelectedFieldMap(ctx).Children)
}

func (r *ReportResolver) Reporter(ctx context.Context) (*UserResolver, error) {
	if r.Report.ReporterID.IsZero() {
		return nil, nil
	}

	return CreateUserResolver(r.gCtx, ctx, r.Report.Reporter, &r.Report.ReporterID, GenerateSelectedFieldMap(ctx).Children)
}

func (r *ReportResolver) Assignees(ctx context.Context) ([]*UserResolver, error) {
	resolvers := make([]*UserResolver, len(r.Report.Assignees))

	for i, user := range r.Report.Assignees {
		resolver, err := CreateUserResolver(r.gCtx, ctx, user, &user.ID, r.fields)
		if err != nil {
			return nil, err
		}

		resolvers[i] = resolver
	}

	return resolvers, nil
}

func (r *ReportResolver) Subject() string {
	return r.Report.Subject
}

func (r *ReportResolver) Body() string {
	return r.Report.Body
}

func (r *ReportResolver) Priority() int32 {
	return r.Report.Priority
}

func (r *ReportResolver) Status() string {
	return string(r.Report.Status)
}

func (r *ReportResolver) CreatedAt() string {
	return r.Report.CreatedAt.Format(time.RFC3339)
}
