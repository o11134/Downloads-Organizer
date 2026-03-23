//go:build windows

package notify

import "github.com/go-toast/toast"

type Notifier struct {
	appID string
}

func New(appID string) *Notifier {
	return &Notifier{appID: appID}
}

func (n *Notifier) Notify(title string, message string) error {
	notification := toast.Notification{
		AppID:   n.appID,
		Title:   title,
		Message: message,
		Audio:   toast.Silent,
	}

	return notification.Push()
}
