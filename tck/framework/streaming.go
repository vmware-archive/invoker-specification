package framework

var Streaming = Suite{
	Name:        "s",
	Description: "Streaming Interaction",
	Port:        8081,
	Cases: []*Testcase{
		{
			Name:        "s-0001",
			Description: "TODO",
			Optional:    true,
			Image:       "upper",
			T: func(port int) {
				panic("BOOM")
			},
		},
		{
			Name:        "s-0002",
			Description: "TODO",
			Image:       "upper",
			T: func(port int) {
			},
		},
	},
}
