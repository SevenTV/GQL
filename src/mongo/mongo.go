package mongo

import (
	"context"
	"time"

	"github.com/SevenTV/ThreeLetterAPI/src/configure"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var Database *mongo.Database

type Pipeline = mongo.Pipeline

func init() {
	uri := configure.Config.GetString("mongo_uri")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	// Send a Ping
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}

	Database = client.Database(configure.Config.GetString("mongo_db"))

	log.Info("mongo, ok")
}

func Collection(name CollectionName) *mongo.Collection {
	return Database.Collection(string(name))
}

type CollectionName string

var (
	CollectionNameEmotes            CollectionName = "emotes"
	CollectionNameUsers             CollectionName = "users"
	CollectionNameBans              CollectionName = "bans"
	CollectionNameReports           CollectionName = "reports"
	CollectionNameBadges            CollectionName = "badges"
	CollectionNameRoles             CollectionName = "roles"
	CollectionNameAudit             CollectionName = "audit"
	CollectionNameEntitlements      CollectionName = "entitlements"
	CollectionNameNotifications     CollectionName = "notifications"
	CollectionNameNotificationsRead CollectionName = "notifications_read"
)
