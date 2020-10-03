package util

import (
	"errors"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
)

func KVCompareAndSet(key string, oldData, newData []byte, updater func(oldData []byte) ([]byte, error)) error {
	for i := 0; i < config.KVCompareAndSetMaxRetries; i++ {
		inserted, appErr := config.Mattermost.KVCompareAndSet(key, oldData, newData)
		if appErr != nil {
			return errors.New(appErr.Error())
		}

		if inserted {
			return nil
		}

		oldData, appErr = config.Mattermost.KVGet(key)
		if appErr != nil {
			return errors.New(appErr.Error())
		}

		var err error
		newData, err = updater(oldData)
		if err != nil {
			return err
		}
	}

	return nil
}
