package agents

const (
	IDClaude   = "claude"
	IDCodex    = "codex"
	IDOpenCode = "opencode"
	IDAgents   = "agents"
)

type Adapter interface {
	ID() string
	DisplayName() string
	Detect() bool
	SkillsDir() string
	CommandsDir() string
}
