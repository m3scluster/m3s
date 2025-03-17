package types

type MesosFrameworkStatus struct {
	Frameworks []struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		UsedResources struct {
			Disk float64 `json:"disk"`
			Mem  float64 `json:"mem"`
			Gpus float64 `json:"gpus"`
			Cpus float64 `json:"cpus"`
		} `json:"used_resources"`
		OfferedResources struct {
			Disk float64 `json:"disk"`
			Mem  float64 `json:"mem"`
			Gpus float64 `json:"gpus"`
			Cpus float64 `json:"cpus"`
		} `json:"offered_resources"`
		Capabilities     []interface{} `json:"capabilities"`
		Hostname         string        `json:"hostname"`
		WebuiURL         string        `json:"webui_url"`
		Active           bool          `json:"active"`
		Connected        bool          `json:"connected"`
		Recovered        bool          `json:"recovered"`
		User             string        `json:"user"`
		FailoverTimeout  float64       `json:"failover_timeout"`
		Checkpoint       bool          `json:"checkpoint"`
		RegisteredTime   float64       `json:"registered_time"`
		UnregisteredTime float64       `json:"unregistered_time"`
		Principal        string        `json:"principal"`
		Resources        struct {
			Disk float64 `json:"disk"`
			Mem  float64 `json:"mem"`
			Gpus float64 `json:"gpus"`
			Cpus float64 `json:"cpus"`
		} `json:"resources"`
		Role             string        `json:"role"`
		Tasks            []interface{} `json:"tasks"`
		UnreachableTasks []interface{} `json:"unreachable_tasks"`
		CompletedTasks   []interface{} `json:"completed_tasks"`
		Offers           []interface{} `json:"offers"`
		Executors        []interface{} `json:"executors"`
		OfferConstraints struct {
		} `json:"offer_constraints"`
	} `json:"frameworks"`
	CompletedFrameworks    []interface{} `json:"completed_frameworks"`
	UnregisteredFrameworks []interface{} `json:"unregistered_frameworks"`
}
