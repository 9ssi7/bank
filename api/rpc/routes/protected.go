package routes

import "fmt"

func protectedActions(srv string, a ...string) []string {
	var actions []string
	for _, method := range a {
		actions = append(actions, fmt.Sprintf("%s/%s", srv, method))
	}
	return actions
}
