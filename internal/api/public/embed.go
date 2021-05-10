package public

import "embed"

// FS embed this folder and childs into binary
//go:embed *
var FS embed.FS
