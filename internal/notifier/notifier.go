package notifier

type (
	Notifier struct {
		dataDir string
	}
)

func NewNotifier(dataDir string) *Notifier {
	return &Notifier{dataDir: dataDir}
}
