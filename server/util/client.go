package util

import circleci2 "github.com/TomTucka/go-circleci/circleci"

func GetCircleciClient(authToken string) *circleci2.APIClient {
	conf := circleci2.NewConfiguration()
	conf.AddDefaultHeader("Circle-Token", string(authToken))
	return circleci2.NewAPIClient(conf)
}
