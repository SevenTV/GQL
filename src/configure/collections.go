package configure

import (
	"github.com/SevenTV/Common/mongo"
)

var (
	CollectionNameEmotes                                 = mongo.CollectionNameEmotes
	CollectionNameUsers                                  = mongo.CollectionNameUsers
	CollectionNameBans              mongo.CollectionName = "bans"
	CollectionNameReports           mongo.CollectionName = "reports"
	CollectionNameBadges            mongo.CollectionName = "badges"
	CollectionNameRoles             mongo.CollectionName = "roles"
	CollectionNameAudit             mongo.CollectionName = "audit"
	CollectionNameEntitlements      mongo.CollectionName = "entitlements"
	CollectionNameNotifications     mongo.CollectionName = "notifications"
	CollectionNameNotificationsRead mongo.CollectionName = "notifications_read"
)

var Indexes = []mongo.IndexRef{}
