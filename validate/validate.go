package validate

// ImageDescription defines base image type and source
type ImageDescription struct {
	Image    string `yaml:"image"`
	Location string `yaml:"location"`
}

// Packages defines system packages to install in container
type Packages struct {
	Manager string   `yaml:"manager,omitempty"`
	System  []string `yaml:"system,omitempty"`
}

// RunCommand defines custom commands to run inside continer
type RunCommand struct {
	Command string   `yaml:"command,omitempty"`
	Content string   `yaml:"content,omitempty"`
	Args    []string `yaml:"args,omitempty"`
}

//CopyCommand defines source and target for files to be copied into container
type CopyCommand struct {
	Source string `yaml:"source,omitempty"`
	Target string `yaml:"target,omitempty"`
	Action string `yaml:"action,omitempty"`
}

// Service defines command to install app
type Service struct {
	Name      string    `yaml:"name,omitempty"`
	Version   string    `yaml:"version,omitempty"`
	IP        string    `yaml:"ip"`
	Ports     []string  `yaml:"ports"`
	Resources Resources `yaml:"resources"`
}

// Resources defines resources allocated to service
type Resources struct {
	RAM string `yaml:"ram"`
	CPU int64  `yaml:"cpu"`
	GPU bool   `yaml:"gpu"`
}

// Bravefile describes unit configuration
type Bravefile struct {
	Base            ImageDescription `yaml:"base"`
	SystemPackages  Packages         `yaml:"packages,omitempty"`
	Run             []RunCommand     `yaml:"run,omitempty"`
	Copy            []CopyCommand    `yaml:"copy,omitempty"`
	PlatformService Service          `yaml:"service,omitempty"`
}

// NewBravefile ..
func NewBravefile() *Bravefile {
	return &Bravefile{}
}
