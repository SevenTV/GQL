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
	"go.mongodb.org/mongo-driver/mongo/options"
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
		SetStatus(structures.ReportStatusOpen).
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

func (r *Resolver) EditReport(ctx context.Context, args struct {
	ReportID string
	Data     EditReportInput
}) (*query.ReportResolver, error) {
	actor, ok := ctx.Value(helpers.UserKey).(*structures.User)
	if !ok {
		return nil, helpers.ErrUnauthorized
	}
	if !actor.HasPermission(structures.RolePermissionManageReports) {
		return nil, helpers.ErrAccessDenied
	}

	// Find the report
	report := &structures.Report{}
	reportID, err := primitive.ObjectIDFromHex(args.ReportID)
	if err != nil {
		return nil, helpers.ErrBadObjectID
	}
	if err = r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameReports).FindOne(ctx, bson.M{"_id": reportID}).Decode(report); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, helpers.ErrUnknownEmote
		}

		logrus.WithError(err).Error("mongo")
		return nil, err
	}

	// Apply mutations
	rb := structures.NewReportBuilder(report)
	if args.Data.Priority != nil { // Changing the priority of the report
		rb.SetPriority(*args.Data.Priority)
	}
	if args.Data.Status != nil { // Changing report status
		rb.SetStatus(*args.Data.Status)
	}
	if args.Data.Assignee != nil { // Adding or removing assignees
		a := *args.Data.Assignee
		assigneeID, err := primitive.ObjectIDFromHex(a[1:])
		if err != nil {
			return nil, helpers.ErrBadObjectID
		}

		state := a[0]
		switch state {
		case '+':
			rb.AddAssignee(assigneeID)
		case '-':
			rb.RemoveAssignee(assigneeID)
		default:
			return nil, fmt.Errorf("invalid operator for assignee mutation")
		}
	}

	// Do update
	if err = r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameReports).FindOneAndUpdate(ctx, bson.M{
		"_id": reportID,
	}, rb.Update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(report); err != nil {
		return nil, err
	}

	return query.CreateReportResolver(r.Ctx, ctx, report, &report.ID, query.GenerateSelectedFieldMap(ctx).Children)
}

type CreateReportInput struct {
	TargetKind structures.ReportTargetKind `json:"target_kind"`
	TargetID   string                      `json:"target_id"`
	Subject    string                      `json:"subject"`
	Body       string                      `json:"body"`
}

type EditReportInput struct {
	Priority *int32                   `json:"priority"`
	Status   *structures.ReportStatus `json:"status"`
	Assignee *string                  `json:"assignee"`
	Note     *EditReportNoteInput     `json:"note"`
}

type EditReportNoteInput struct {
	Timestamp time.Time `json:"timestamp"`
	Content   *string   `json:"content"`
	Internal  *bool     `json:"internal"`
	Reply     *string   `json:"reply"`
}
