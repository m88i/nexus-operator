package controller

import (
	"github.com/m88i/nexus-operator/pkg/controller/nexus"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, nexus.Add)
}
