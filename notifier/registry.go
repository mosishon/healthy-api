package notifier

type Registry struct {
	notifiers map[string]Notifier
}

func NewRegistry() *Registry {
	return &Registry{
		notifiers: make(map[string]Notifier),
	}
}

func (r *Registry) Register(name string, notifier Notifier) {
	r.notifiers[name] = notifier
}

func (r *Registry) Get(name string) Notifier {
	n, _ := r.notifiers[name]
	return n
}
