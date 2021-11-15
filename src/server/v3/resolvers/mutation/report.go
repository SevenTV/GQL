package mutation

import (
	"context"
	"fmt"
	"time"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/GQL/src/server/v3/helpers"
	"github.com/SevenTV/GQL/src/server/v3/resolvers/query"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	REPORT_SUBJECT_MIN_LENGTH      = 4
	REPORT_SUBJECT_MAX_LENGTH      = 48
	REPORT_BODY_MIN_LENGTH         = 8
	REPORT_BODY_MAX_LENGTH         = 2000
	REPORT_ALLOWED_ACTIVE_PER_USER = 3
)

func (r *Resolver) CreateReport(ctx context.Context, args struct {
	Data CreateReportInput
}) (*query.ReportResolver, error) {
	// Get actor
	actor, ok := ctx.Value(helpers.UserKey).(*structures.User)
	if !ok {
		return nil, helpers.ErrUnauthorized
	}
	// Actor must have the permission to create reports
	if !actor.HasPermission(structures.RolePermissionReportCreate) {
		return nil, helpers.ErrAccessDenied
	}

	// Verify the target
	targetID, err := primitive.ObjectIDFromHex(args.Data.TargetID)
	if err != nil {
		return nil, helpers.ErrBadObjectID
	}
	var collName mongo.CollectionName
	switch args.Data.TargetKind {
	case structures.ReportTargetKindUser:
		collName = mongo.CollectionNameUsers
	case structures.ReportTargetKindEmote:
		collName = mongo.CollectionNameEmotes
	default:
		return nil, fmt.Errorf("unknown target kind")
	}
	if c, _ := r.Ctx.Inst().Mongo.Collection(collName).CountDocuments(ctx, bson.M{
		"_id": targetID,
	}); c == 0 {
		return nil, fmt.Errorf("target not found")
	}

	// Validate the input
	if len(args.Data.Subject) < REPORT_SUBJECT_MIN_LENGTH {
		return nil, fmt.Errorf("subject must be at least %d characters", REPORT_SUBJECT_MIN_LENGTH)
	}
	if len(args.Data.Subject) > REPORT_SUBJECT_MAX_LENGTH {
		return nil, fmt.Errorf("subject must be under %d characters", REPORT_SUBJECT_MAX_LENGTH)
	}
	if len(args.Data.Body) < REPORT_BODY_MIN_LENGTH {
		return nil, fmt.Errorf("body must be at least %d characters", REPORT_BODY_MIN_LENGTH)
	}
	if len(args.Data.Body) > REPORT_BODY_MAX_LENGTH {
		return nil, fmt.Errorf("body must be under %d characters", REPORT_BODY_MAX_LENGTH)
	}

	// Create the report
	rb := structures.NewReportBuilder(&structures.Report{})
	rb.Report.ID = primitive.NewObjectID()
	rb.SetTargetKind(args.Data.TargetKind).
		SetTargetID(targetID).
		SetReporterID(actor.ID).
		SetSubject(args.Data.Subject).
		SetBody(args.Data.Body).
		SetCreatedAt(time.Now())

	result, err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameReports).InsertOne(ctx, rb.Report)
	if err != nil {
		logrus.WithError(err).Error("mongo")
		return nil, err
	}

	returnID := result.InsertedID.(primitive.ObjectID)
	return query.CreateReportResolver(r.Ctx, ctx, nil, &returnID, query.GenerateSelectedFieldMap(ctx).Children)
}

type CreateReportInput struct {
	TargetKind structures.ReportTargetKind `json:"target_kind"`
	TargetID   string                      `json:"target_id"`
	Subject    string                      `json:"subject"`
	Body       string                      `json:"body"`
}
