//go:build !windows

package notify

type Notifier struct{}

func New(appID string) *Notifier {
	_ = appID
	return &Notifier{}
}

func (n *Notifier) Notify(title string, message string) error {
	_ = title
	_ = message
	return nil
}
