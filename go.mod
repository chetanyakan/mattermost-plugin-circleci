module github.com/chetanyakan/mattermost-plugin-circleci

go 1.14

require (
	github.com/TomTucka/go-circleci v1.1.0
	github.com/antihax/optional v1.0.0
	github.com/dustin/go-humanize v1.0.0
	github.com/gorilla/mux v1.8.0
	github.com/mattermost/mattermost-server/v5 v5.36.1
	github.com/pkg/errors v0.9.1
	github.com/thoas/go-funk v0.7.0
	go.uber.org/atomic v1.7.0
)

// To access the TomTucka/go-circleci v2 repo fork
replace github.com/TomTucka/go-circleci => github.com/chetanyakan/go-circleci v1.1.0
