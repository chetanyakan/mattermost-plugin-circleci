package store

import (
	"bytes"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
)

// from https://github.com/mattermost/mattermost-plugin-jira/blob/0c04ea41daf62fcfb6682644ea5927370fc7ebe5/server/subscribe.go#L655
func AtomicModify(key string, modify func(initialValue []byte) ([]byte, error)) error {
	readModify := func() ([]byte, []byte, error) {
		initialBytes, appErr := config.Mattermost.KVGet(key)
		if appErr != nil {
			return nil, nil, errors.Wrap(appErr, "unable to read initial value")
		}

		modifiedBytes, err := modify(initialBytes)
		if err != nil {
			return nil, nil, errors.Wrap(err, "modification error")
		}

		return initialBytes, modifiedBytes, nil
	}

	var (
		retryLimit     = 5
		retryWait      = 30 * time.Millisecond
		success        = false
		currentAttempt = 0
	)
	for !success {
		initialBytes, newValue, err := readModify()

		if err != nil {
			return err
		}

		var setError *model.AppError
		success, setError = config.Mattermost.KVCompareAndSet(key, initialBytes, newValue)
		if setError != nil {
			return errors.Wrap(setError, "problem writing value")
		}

		if currentAttempt == 0 && bytes.Equal(initialBytes, newValue) {
			return nil
		}

		currentAttempt++
		if currentAttempt >= retryLimit {
			return errors.New("reached write attempt limit")
		}

		time.Sleep(retryWait)
	}

	return nil
}
