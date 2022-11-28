package internaltypes

type Result struct {
	Stage   Stage
	Result  bool
	Skipped bool
	Outputs []Env
	Logs    []string
}

type Input struct {
	Name     string `mapstructure:"name" validate:"nonzero,nowhitespace"`
	Default  string `mapstructure:"default"`
	Optional bool   `mapstructure:"optional,omitempty"`
}

type Stage struct {
	Stage     string   `mapstructure:"stage" validate:"nonzero"`
	ID        string   `mapstructure:"id,omitempty" validate:"nonzero,nowhitespace"`
	Script    []string `mapstructure:"script" validate:"nonzero"`
	If        string   `mapstructure:"if,omitempty"`
	Clean     bool     `mapstructure:"clean,omitempty"`
	Env       []Env    `mapstructure:"env,omitempty"`
	Artifacts []string `mapstructure:"artifacts,omitempty"`
	Image     string   `mapstructure:"image,omitempty"`
	Needs     string   `mapstructure:"needs,omitempty" validate:"nowhitespace"`
	Import    []Import `mapstructure:"import,omitempty"`
}

type Import struct {
	From      string   `mapstructure:"from" validate:"nonzero,nowhitespace"`
	Artifacts []string `mapstructure:"artifacts" validate:"nonzero"`
}

type Env struct {
	Name  string `mapstructure:"name" validate:"nonzero,nowhitespace"`
	Value string `mapstructure:"value" validate:"nonzero"`
}

type Workflow struct {
	ID          string  `mapstructure:"id" validate:"nonzero,nowhitespace"`
	Image       string  `mapstructure:"image" validate:"nonzero,nowhitespace"`
	Description string  `mapstructure:"description"`
	Env         []Env   `mapstructure:"env"`
	Input       []Input `mapstructure:"input"`
	// Mount       bool    `mapstructure:"mount"`
	Stages []Stage `mapstructure:"stages" validate:"nonzero"`
	Repo   string  `mapstructure:"repository,omitempty"` // To be filled automatically. Not part of YAML.
}

type WorkflowArgument struct {
	Repo     string            `json:"repo" xml:"repo" form:"repo" query:"repo" mapstructure:"repo" validate:"nonzero"`
	Workflow string            `json:"workflow" xml:"workflow" form:"workflow" query:"workflow" mapstructure:"workflow" validate:"nonzero"`
	Args     map[string]string `json:"args" xml:"args" form:"args" query:"args" mapstructure:"args" validate:"nonzero"`
}
