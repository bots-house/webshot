package internal

type BuildInfo struct {
	Version string `json:"version"`
	Ref     string `json:"ref"`
	Time    string `json:"time"`
}

func (info *BuildInfo) Release() string {
	return info.Version + "-" + info.Ref
}
