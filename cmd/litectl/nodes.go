package main

func nodes(params []string) {
	client := getClient()
	if len(params) == 0 {
		// Get all
		n, err := client.Nodes().List()
		if err != nil {
			exitErr(err)
		}
		output(n.All())
	} else {
		// Get specific
		name := params[0]
		n, err := client.Nodes().Get(name)
		if err != nil {
			exitErr(err)
		}
		output(n)
	}
}
