package instance

import "github.com/SevenTV/Common/mongo"

type Mongo interface {
	mongo.Instance
}

type mongoInst struct {
	mongo.Instance
}

func WrapMongo(mongo mongo.Instance) Mongo {
	return &mongoInst{Instance: mongo}
}
