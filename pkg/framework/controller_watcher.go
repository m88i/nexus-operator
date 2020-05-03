//     Copyright 2020 Nexus Operator and/or its authors
//
//     This file is part of Nexus Operator.
//
//     Nexus Operator is free software: you can redistribute it and/or modify
//     it under the terms of the GNU General Public License as published by
//     the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.
//
//     Nexus Operator is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU General Public License for more details.
//
//     You should have received a copy of the GNU General Public License
//     along with Nexus Operator.  If not, see <https://www.gnu.org/licenses/>.

package framework

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// WatchedObjects objects that the controller supposed to watch for
type WatchedObjects struct {
	// GroupVersion for the watched objects
	GroupVersion schema.GroupVersion
	// AddToScheme function to register the Scheme to the Kubernetes Client. This will enable the controller to query for those objects during the reconcile loop.
	AddToScheme func(scheme *runtime.Scheme) error
	// Objects list of required objects that should be watched by the controller
	Objects []runtime.Object
	// Owner of the object if different from the actual controller
	Owner runtime.Object
}

// ControllerWatcher helps to add required objects to the controller watch list given the required runtime objects
type ControllerWatcher interface {
	// Watch add the given objects to the controller watch list
	Watch(objects ...WatchedObjects) (err error)
	// AreAllObjectsWatched verifies if this instance already has registered every required object in the watch list
	AreAllObjectsWatched() bool
	// IsGroupWatched verifies if the given group has it's objects watched or not
	IsGroupWatched(group string) bool
}

// NewControllerWatcher creates a new ControllerWatcher to control the objects that needed to be watched
func NewControllerWatcher(discoveryClient discovery.DiscoveryInterface, manager controllerruntime.Manager, controller controller.Controller, owner runtime.Object) ControllerWatcher {
	return &controllerWatcher{
		manager:          manager,
		controller:       controller,
		owner:            owner,
		groupsNotWatched: map[string]bool{},
		discoverClient:   discoveryClient,
	}
}

type controllerWatcher struct {
	discoverClient   discovery.DiscoveryInterface
	manager          controllerruntime.Manager
	controller       controller.Controller
	owner            runtime.Object
	groupsNotWatched map[string]bool
}

func (c *controllerWatcher) AreAllObjectsWatched() bool {
	return len(c.groupsNotWatched) == 0
}

func (c *controllerWatcher) IsGroupWatched(group string) bool {
	if len(c.groupsNotWatched) == 0 {
		return true
	}
	_, exists := c.groupsNotWatched[group]
	return !exists
}

func (c *controllerWatcher) Watch(watchedObjects ...WatchedObjects) (err error) {
	serverGroupMap, err := c.getServerGroupMap()
	if err != nil {
		return
	}

	var addToScheme runtime.SchemeBuilder
	var desiredObjects []WatchedObjects

	for _, object := range watchedObjects {
		// core resources
		if object.AddToScheme == nil {
			desiredObjects = append(desiredObjects, object)
		} else {
			if _, found := serverGroupMap[object.GroupVersion.Group][object.GroupVersion.Version]; found {
				addToScheme = append(addToScheme, object.AddToScheme)
				desiredObjects = append(desiredObjects, object)
				delete(c.groupsNotWatched, object.GroupVersion.Group)
			} else {
				c.groupsNotWatched[object.GroupVersion.Group] = true
				log.Warnf("Impossible to register GroupVersion %s. CRD not installed in the cluster, controller might not behave as expected", object.GroupVersion)
			}
		}
	}

	if len(addToScheme) > 0 {
		log.Debug("Registering additional controller schemes")
		if err = addToScheme.AddToScheme(c.manager.GetScheme()); err != nil {
			return
		}
	}

	ownerHandler := &handler.EnqueueRequestForOwner{IsController: true, OwnerType: c.owner}
	for _, desiredObject := range desiredObjects {
		for _, runtimeObj := range desiredObject.Objects {
			if desiredObject.Owner == nil {
				if err = c.controller.Watch(&source.Kind{Type: runtimeObj}, ownerHandler); err != nil {
					return
				}
			} else {
				if err = c.controller.Watch(
					&source.Kind{Type: runtimeObj},
					&handler.EnqueueRequestForOwner{IsController: true, OwnerType: desiredObject.Owner}); err != nil {
					return
				}
			}
		}
	}

	return
}

func (c *controllerWatcher) getServerGroupMap() (map[string]map[string]bool, error) {
	serverGroups, err := c.discoverClient.ServerGroups()
	if err != nil {
		return nil, fmt.Errorf("couldn't fetch server groups list: %v", err)
	}

	serverGroupMap := map[string]map[string]bool{}
	for _, serverGroup := range serverGroups.Groups {
		versions := make(map[string]bool)
		for _, version := range serverGroup.Versions {
			versions[version.Version] = true
		}
		serverGroupMap[serverGroup.Name] = versions
	}
	return serverGroupMap, nil
}
