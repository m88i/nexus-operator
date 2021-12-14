//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Copyright 2020 Nexus Operator and/or its authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Nexus) DeepCopyInto(out *Nexus) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Nexus.
func (in *Nexus) DeepCopy() *Nexus {
	if in == nil {
		return nil
	}
	out := new(Nexus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Nexus) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NexusAutomaticUpdate) DeepCopyInto(out *NexusAutomaticUpdate) {
	*out = *in
	if in.MinorVersion != nil {
		in, out := &in.MinorVersion, &out.MinorVersion
		*out = new(int)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NexusAutomaticUpdate.
func (in *NexusAutomaticUpdate) DeepCopy() *NexusAutomaticUpdate {
	if in == nil {
		return nil
	}
	out := new(NexusAutomaticUpdate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NexusList) DeepCopyInto(out *NexusList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Nexus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NexusList.
func (in *NexusList) DeepCopy() *NexusList {
	if in == nil {
		return nil
	}
	out := new(NexusList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NexusList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NexusNetworking) DeepCopyInto(out *NexusNetworking) {
	*out = *in
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	out.TLS = in.TLS
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NexusNetworking.
func (in *NexusNetworking) DeepCopy() *NexusNetworking {
	if in == nil {
		return nil
	}
	out := new(NexusNetworking)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NexusNetworkingTLS) DeepCopyInto(out *NexusNetworkingTLS) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NexusNetworkingTLS.
func (in *NexusNetworkingTLS) DeepCopy() *NexusNetworkingTLS {
	if in == nil {
		return nil
	}
	out := new(NexusNetworkingTLS)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NexusPersistence) DeepCopyInto(out *NexusPersistence) {
	*out = *in
	if in.ExtraVolumes != nil {
		in, out := &in.ExtraVolumes, &out.ExtraVolumes
		*out = make([]NexusVolume, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NexusPersistence.
func (in *NexusPersistence) DeepCopy() *NexusPersistence {
	if in == nil {
		return nil
	}
	out := new(NexusPersistence)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NexusProbe) DeepCopyInto(out *NexusProbe) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NexusProbe.
func (in *NexusProbe) DeepCopy() *NexusProbe {
	if in == nil {
		return nil
	}
	out := new(NexusProbe)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NexusSpec) DeepCopyInto(out *NexusSpec) {
	*out = *in
	in.AutomaticUpdate.DeepCopyInto(&out.AutomaticUpdate)
	in.Resources.DeepCopyInto(&out.Resources)
	in.Persistence.DeepCopyInto(&out.Persistence)
	in.Networking.DeepCopyInto(&out.Networking)
	if in.LivenessProbe != nil {
		in, out := &in.LivenessProbe, &out.LivenessProbe
		*out = new(NexusProbe)
		**out = **in
	}
	if in.ReadinessProbe != nil {
		in, out := &in.ReadinessProbe, &out.ReadinessProbe
		*out = new(NexusProbe)
		**out = **in
	}
	out.ServerOperations = in.ServerOperations
	if in.Properties != nil {
		in, out := &in.Properties, &out.Properties
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NexusSpec.
func (in *NexusSpec) DeepCopy() *NexusSpec {
	if in == nil {
		return nil
	}
	out := new(NexusSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NexusStatus) DeepCopyInto(out *NexusStatus) {
	*out = *in
	in.DeploymentStatus.DeepCopyInto(&out.DeploymentStatus)
	if in.UpdateConditions != nil {
		in, out := &in.UpdateConditions, &out.UpdateConditions
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	out.ServerOperationsStatus = in.ServerOperationsStatus
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NexusStatus.
func (in *NexusStatus) DeepCopy() *NexusStatus {
	if in == nil {
		return nil
	}
	out := new(NexusStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NexusVolume) DeepCopyInto(out *NexusVolume) {
	*out = *in
	in.Volume.DeepCopyInto(&out.Volume)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NexusVolume.
func (in *NexusVolume) DeepCopy() *NexusVolume {
	if in == nil {
		return nil
	}
	out := new(NexusVolume)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OperationsStatus) DeepCopyInto(out *OperationsStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OperationsStatus.
func (in *OperationsStatus) DeepCopy() *OperationsStatus {
	if in == nil {
		return nil
	}
	out := new(OperationsStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServerOperationsOpts) DeepCopyInto(out *ServerOperationsOpts) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServerOperationsOpts.
func (in *ServerOperationsOpts) DeepCopy() *ServerOperationsOpts {
	if in == nil {
		return nil
	}
	out := new(ServerOperationsOpts)
	in.DeepCopyInto(out)
	return out
}
