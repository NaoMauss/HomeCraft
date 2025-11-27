package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto copies all properties of this object into another object of the
// same type that is provided as a pointer.
func (in *MinecraftServer) DeepCopyInto(out *MinecraftServer) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy copies the receiver, creating a new MinecraftServer.
func (in *MinecraftServer) DeepCopy() *MinecraftServer {
	if in == nil {
		return nil
	}
	out := new(MinecraftServer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject copies the receiver, creating a new runtime.Object.
func (in *MinecraftServer) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto copies all properties of this object into another object of the
// same type that is provided as a pointer.
func (in *MinecraftServerList) DeepCopyInto(out *MinecraftServerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MinecraftServer, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy copies the receiver, creating a new MinecraftServerList.
func (in *MinecraftServerList) DeepCopy() *MinecraftServerList {
	if in == nil {
		return nil
	}
	out := new(MinecraftServerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject copies the receiver, creating a new runtime.Object.
func (in *MinecraftServerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto copies all properties of this object into another object of the
// same type that is provided as a pointer.
func (in *MinecraftServerSpec) DeepCopyInto(out *MinecraftServerSpec) {
	*out = *in
}

// DeepCopy copies the receiver, creating a new MinecraftServerSpec.
func (in *MinecraftServerSpec) DeepCopy() *MinecraftServerSpec {
	if in == nil {
		return nil
	}
	out := new(MinecraftServerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies all properties of this object into another object of the
// same type that is provided as a pointer.
func (in *MinecraftServerStatus) DeepCopyInto(out *MinecraftServerStatus) {
	*out = *in
	in.LastUpdated.DeepCopyInto(&out.LastUpdated)
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy copies the receiver, creating a new MinecraftServerStatus.
func (in *MinecraftServerStatus) DeepCopy() *MinecraftServerStatus {
	if in == nil {
		return nil
	}
	out := new(MinecraftServerStatus)
	in.DeepCopyInto(out)
	return out
}
