package ami

import "context"

type UseCase interface {
	NewClient(ctx context.Context, host, unique_id, channel, number string) error
	Hangup(host, unique_id, channel string) error
}
