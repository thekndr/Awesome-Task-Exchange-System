module github.com/thekndr/ates/auth

go 1.22

replace github.com/thekndr/ates/auth_client => ../auth_client

replace github.com/thekndr/ates/common => ../common

require (
	github.com/confluentinc/confluent-kafka-go v1.9.2
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/lib/pq v1.10.9
	github.com/thekndr/ates/auth_client v0.0.0-00010101000000-000000000000
	github.com/thekndr/ates/common v0.0.0-00010101000000-000000000000
)

require (
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
)
