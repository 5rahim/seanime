//go:build android || ios

package notifier

func defaultPush(title, message, icon string) error {
	return nil
}
