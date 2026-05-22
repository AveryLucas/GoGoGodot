package ScriptLanguageExtension

import gd "github.com/AveryLucas/gogogd/internal"

type ProfilingInfo = gd.ScriptLanguageExtensionProfilingInfo

type TemplateLocation int

const (
	TemplateBuiltIn TemplateLocation = iota
	TemplateEditor
	TemplateProject
)
