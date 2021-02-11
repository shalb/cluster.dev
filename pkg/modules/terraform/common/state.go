package common

type StateSpecCommon struct {
	InfraName    string            `json:"infra_name"`
	BackendState interface{}       `json:"backend"`
	Name         string            `json:"name"`
	PreHook      *hookSpec         `json:"pre_hook,omitempty"`
	PostHook     *hookSpec         `json:"post_hook,omitempty"`
	Providers    interface{}       `json:"providers,omitempty"`
	Markers      map[string]string `json:"markers,omitempty"`
}

func (m *Module) GetStateCommon() (StateSpecCommon, error) {
	b := m.backendPtr.State()
	if m.postHook != nil {

	}
	st := StateSpecCommon{
		InfraName:    m.InfraName(),
		BackendState: b,
		Name:         m.name,
		PreHook:      m.preHook,
		PostHook:     m.postHook,
		Providers:    m.providers,
		Markers:      m.markers,
	}
	// log.Log.Debugf("%+v", st)
	return st, nil
}
