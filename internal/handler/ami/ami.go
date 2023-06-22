package ami

import "context"

type UseCase interface {
	NewClient(ctx context.Context, host, uniqueID, channel, number string) error
	Hangup(host, uniqueID, channel string) error
}
