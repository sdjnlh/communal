module github.com/sdjnlh/communal

replace (
	github.com/spf13/viper => github.com/topiot/viper v1.7.1-0.20200721234125-12c3cd96c819
	go.uber.org/zap => github.com/uber-go/zap v1.15.0
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20200709230013-948cd5f35899
	golang.org/x/lint => github.com/golang/lint v0.0.0-20200302205851-738671d3881b
	golang.org/x/mod => github.com/golang/mod v0.3.1-0.20200706160632-89ce4c7ba804
	golang.org/x/net => github.com/golang/net v0.0.0-20200707034311-ab3426394381
	golang.org/x/sync => github.com/golang/sync v0.0.0-20200625203802-6e8e738ad208
	golang.org/x/sys => github.com/golang/sys v0.0.0-20200720211630-cb9d2d5c5666
	golang.org/x/text => github.com/golang/text v0.3.3
	golang.org/x/tools => github.com/golang/tools v0.0.0-20200721223218-6123e77877b2
	golang.org/x/xerrors => github.com/golang/xerrors v0.0.0-20191204190536-9bdfabe68543
)

require (
	github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535
	github.com/bwmarrin/snowflake v0.3.0
	github.com/gin-contrib/sessions v0.0.3
	github.com/gin-gonic/gin v1.7.7
	github.com/go-redis/redis/v8 v8.11.3
	github.com/go-xorm/builder v0.3.4
	github.com/golang-jwt/jwt/v4 v4.1.0
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/gorilla/sessions v1.2.0 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/magiconair/properties v1.8.1
	github.com/mitchellh/mapstructure v1.3.2 // indirect
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.6.1
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/go-playground/validator.v8 v8.18.2
	xorm.io/xorm v1.0.3
)

go 1.14
