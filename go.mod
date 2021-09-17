module github.com/SevenTV/GQL

go 1.16

require (
	github.com/bugsnag/panicwrap v1.3.4
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gobuffalo/packr/v2 v2.2.0
	github.com/gofiber/fiber/v2 v2.17.0
	github.com/golang-jwt/jwt/v4 v4.0.0
	github.com/graph-gophers/graphql-go v1.1.0
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/kr/pretty v0.3.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/viper v1.8.1
	go.mongodb.org/mongo-driver v1.7.1
)

replace github.com/graph-gophers/graphql-go => github.com/troydota/graphql-go v0.0.0-20210702180404-92fc941a47cf
