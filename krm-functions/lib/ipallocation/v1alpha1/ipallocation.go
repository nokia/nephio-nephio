package v1alpha1

import (
	"fmt"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"github.com/example.com/foo/pkg/parser"
	ipamv1alpha1 "github.com/nokia/k8s-ipam/apis/alloc/ipam/v1alpha1"
	"github.com/nokia/k8s-ipam/pkg/iputil"
)

var (
	prefixKind          = []string{"spec", "kind"}
	networkInstanceName = []string{"spec", "networkInstance", "name"}
	addressFamily       = []string{"spec", "addressFamily"}
	prefix              = []string{"spec", "prefix"}
	prefixLength        = []string{"spec", "prefixLength"}
	index               = []string{"spec", "index"}
	selectorLabels      = []string{"spec", "selector", "matchLabels"}
	labels              = []string{"spec", "labels"}
	createPrefix        = []string{"spec", "createPrefix"}
	allocatedPrefix     = []string{"status", "prefix"}
	allocatedGateway    = []string{"status", "gateway"}
)

type IPAllocation interface {
	parser.Parser[*ipamv1alpha1.IPAllocation]

	SetSpec(*ipamv1alpha1.IPAllocationSpec) error
	// GetPrefixKind returns the prefixKind from the spec
	// if an error occurs or the attribute is not present an empty string is returned
	GetPrefixKind() ipamv1alpha1.PrefixKind
	// GetNetworkInstanceName returns the name of the networkInstance from the spec
	// if an error occurs or the attribute is not present an empty string is returned
	GetNetworkInstanceName() string
	// GetNetworkInstanceName returns the name of the networkInstance from the spec
	// if an error occurs or the attribute is not present an empty string is returned
	GetAddressFamily() iputil.AddressFamily
	// GetPrefix returns the prefix from the spec
	// if an error occurs or the attribute is not present an empty string is returned
	GetPrefix() string
	// GetPrefixLength returns the prefixlength from the spec
	// if an error occurs or the attribute is not present 0 is returned
	GetPrefixLength() uint8
	// GetIndex returns the index from the spec
	// if an error occurs or the attribute is not present 0 is returned
	GetIndex() uint32
	// GetSelectorLabels returns the selector Labels from the spec
	// if an error occurs or the attribute is not present an empty map[string]string is returned
	GetSelectorLabels() map[string]string
	// GetSpecLabels returns the labelsfrom the spec
	// if an error occurs or the attribute is not present an empty map[string]string is returned
	GetSpecLabels() map[string]string
	// GetCreatePrefix returns the create prefix from the spec
	// if an error occurs or the attribute is not present false is returned
	GetCreatePrefix() bool
	// GetAllocatedPrefix returns the prefix from the status
	// if an error occurs or the attribute is not present an empty string is returned
	GetAllocatedPrefix() string
	// GetAllocatedGateway returns the gateway from the status
	// if an error occurs or the attribute is not present an empty string is returned
	GetAllocatedGateway() string
	// SetPrefixKind sets the prefixKind in the spec
	SetPrefixKind(ipamv1alpha1.PrefixKind) error
	// SetNetworkInstanceName sets the name of the networkInstance in the spec
	SetNetworkInstanceName(s string) error
	// SetAddressFamily sets the address family in the spec
	SetAddressFamily(iputil.AddressFamily) error
	// SetPrefix sets the prefix in the spec
	SetPrefix(string) error
	// SetPrefixLength sets the prefix length in the spec
	SetPrefixLength(uint8) error
	// SetIndex sets the index in the spec
	SetIndex(uint32) error
	// SetSelectorLabels sets the selector matchLabels in the spec
	SetSelectorLabels(map[string]string) error
	// SetSpecLabels sets the labels in the spec
	SetSpecLabels(map[string]string) error
	// SetCreatePrefix sets the create prefix in the spec
	SetCreatePrefix(bool) error
	// SetAllocatedPrefix sets the allocated prefix in the status
	SetAllocatedPrefix(string) error
	// SetAllocatedGateway sets the allocated gateway in the status
	SetAllocatedGateway(string) error
	// DeleteAddressFamily deletes the address family from the spec
	DeleteAddressFamily() error
	// DeletePrefix deletes the prefix from the spec
	DeletePrefix() error
	// DeletePrefixLength deletes the prefix length from the spec
	DeletePrefixLength() error
	// DeleteIndex deletes the index from the spec
	DeleteIndex() error
	// DeleteSpecLabels deletes the labels from the spec
	DeleteSpecLabels() error
	// DeleteCreatePrefix deletes the create prefix from the spec
	DeleteCreatePrefix() error
	// DeleteAllocatedPrefix deletes the allocated prefix from the status
	DeleteAllocatedPrefix() error
	// DeleteAllocatedGateway deletes the allocated gateway from the status
	DeleteAllocatedGateway() error
}

// NewFromKubeObject creates a new parser interface
// It expects a *fn.KubeObject as input representing the serialized yaml file
func NewFromKubeObject(o *fn.KubeObject) IPAllocation {
	return &obj{
		p: parser.NewFromKubeObject[*ipamv1alpha1.IPAllocation](o),
	}
}

// NewFromYaml creates a new parser interface
// It expects raw byte slice as input representing the serialized yaml file
func NewFromYAML(b []byte) (IPAllocation, error) {
	p, err := parser.NewFromYaml[*ipamv1alpha1.IPAllocation](b)
	if err != nil {
		return nil, err
	}
	return &obj{
		p: p,
	}, nil
}

// NewFromGoStruct creates a new parser interface
// It expects a go struct representing the interface krm resource
func NewFromGoStruct(x *ipamv1alpha1.IPAllocation) (IPAllocation, error) {
	p, err := parser.NewFromGoStruct[*ipamv1alpha1.IPAllocation](x)
	if err != nil {
		return nil, err
	}
	return &obj{
		p: p,
	}, nil
}

type obj struct {
	p parser.Parser[*ipamv1alpha1.IPAllocation]
}

// GetKubeObject returns the present kubeObject
func (r *obj) GetKubeObject() *fn.KubeObject {
	return r.p.GetKubeObject()
}

func (r *obj) GetGoStruct() (*ipamv1alpha1.IPAllocation, error) {
	return r.p.GetGoStruct()
}

func (r *obj) GetStringValue(fields ...string) string {
	return r.p.GetStringValue()
}

func (r *obj) GetBoolValue(fields ...string) bool {
	return r.p.GetBoolValue()
}

func (r *obj) GetIntValue(fields ...string) int {
	return r.p.GetIntValue()
}

func (r *obj) GetStringMap(fields ...string) map[string]string {
	return r.p.GetStringMap()
}

func (r *obj) SetNestedString(s string, fields ...string) error {
	return r.p.SetNestedString(s, fields...)
}

func (r *obj) SetNestedInt(s int, fields ...string) error {
	return r.p.SetNestedInt(s, fields...)
}

func (r *obj) SetNestedBool(s bool, fields ...string) error {
	return r.p.SetNestedBool(s, fields...)
}

func (r *obj) SetNestedMap(s map[string]string, fields ...string) error {
	return r.p.SetNestedMap(s, fields...)
}

func (r *obj) DeleteNestedField(fields ...string) error {
	return r.p.DeleteNestedField(fields...)
}

func (r *obj) SetSpec(spec *ipamv1alpha1.IPAllocationSpec) error {
	if spec == nil {
		return nil
	}
	if spec.PrefixKind != "" {
		if err := r.SetPrefixKind(spec.PrefixKind); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("prefixKind is required")
	}
	if spec.NetworkInstance != nil {
		if err := r.SetNetworkInstanceName(string(spec.NetworkInstance.Name)); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("networkInstance is required")
	}

	if spec.AddressFamily != "" {
		if err := r.SetAddressFamily(spec.AddressFamily); err != nil {
			return err
		}
	} else {
		if err := r.DeleteAddressFamily(); err != nil {
			return err
		}
	}
	if spec.Prefix != "" {
		if err := r.SetPrefix(spec.Prefix); err != nil {
			return err
		}
	} else {
		if err := r.DeletePrefix(); err != nil {
			return err
		}
	}
	if spec.PrefixLength != 0 {
		if err := r.SetPrefixLength(spec.PrefixLength); err != nil {
			return err
		}
	} else {
		if err := r.DeletePrefixLength(); err != nil {
			return err
		}
	}
	if spec.Index != 0 {
		if err := r.SetIndex(spec.Index); err != nil {
			return err
		}
	} else {
		if err := r.DeleteIndex(); err != nil {
			return err
		}
	}
	if spec.Selector != nil && spec.Selector.MatchLabels != nil {
		if err := r.SetSelectorLabels(spec.Selector.MatchLabels); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("selector matchlabels is required")
	}
	if spec.Labels != nil {
		if err := r.SetSpecLabels(spec.Labels); err != nil {
			return err
		}
	} else {
		if err := r.DeleteSpecLabels(); err != nil {
			return err
		}
	}
	if spec.CreatePrefix {
		if err := r.SetCreatePrefix(spec.CreatePrefix); err != nil {
			return err
		}
	} else {
		if err := r.DeleteCreatePrefix(); err != nil {
			return err
		}
	}

	return nil
}

// GetPrefixKind returns the prefixKind from the spec
// if an error occurs or the attribute is not present an empty string is returned
func (r *obj) GetPrefixKind() ipamv1alpha1.PrefixKind {
	return ipamv1alpha1.PrefixKind(r.p.GetStringValue(prefixKind...))
}

// GetNetworkInstanceName returns the name of the networkInstance from the spec
// if an error occurs or the attribute is not present an empty string is returned
func (r *obj) GetNetworkInstanceName() string {
	return r.p.GetStringValue(networkInstanceName...)
}

// GetNetworkInstanceName returns the name of the networkInstance from the spec
// if an error occurs or the attribute is not present an empty string is returned
func (r *obj) GetAddressFamily() iputil.AddressFamily {
	return iputil.AddressFamily(r.p.GetStringValue(addressFamily...))
}

// GetPrefix returns the prefix from the spec
// if an error occurs or the attribute is not present an empty string is returned
func (r *obj) GetPrefix() string {
	return r.p.GetStringValue(prefix...)
}

// GetPrefixLength returns the prefixlength from the spec
// if an error occurs or the attribute is not present 0 is returned
func (r *obj) GetPrefixLength() uint8 {
	return uint8(r.p.GetIntValue(prefixLength...))
}

// GetIndex returns the index from the spec
// if an error occurs or the attribute is not present 0 is returned
func (r *obj) GetIndex() uint32 {
	return uint32(r.p.GetIntValue(index...))
}

// GetSelectorLabels returns the selector Labels from the spec
// if an error occurs or the attribute is not present an empty map[string]string is returned
func (r *obj) GetSelectorLabels() map[string]string {
	return r.p.GetStringMap(selectorLabels...)
}

// GetSpecLabels returns the labelsfrom the spec
// if an error occurs or the attribute is not present an empty map[string]string is returned
func (r *obj) GetSpecLabels() map[string]string {
	return r.p.GetStringMap(labels...)
}

// GetCreatePrefix returns the create prefix from the spec
// if an error occurs or the attribute is not present false is returned
func (r *obj) GetCreatePrefix() bool {
	return r.p.GetBoolValue(createPrefix...)
}

// GetAllocatedPrefix returns the prefix from the status
// if an error occurs or the attribute is not present an empty string is returned
func (r *obj) GetAllocatedPrefix() string {
	return r.p.GetStringValue(allocatedPrefix...)
}

// GetAllocatedGateway returns the gateway from the status
// if an error occurs or the attribute is not present an empty string is returned
func (r *obj) GetAllocatedGateway() string {
	return r.p.GetStringValue(allocatedGateway...)
}

// SetPrefixKind sets the prefixKind in the spec
func (r *obj) SetPrefixKind(s ipamv1alpha1.PrefixKind) error {
	return r.p.SetNestedString(string(s), prefixKind...)
}

// SetNetworkInstanceName sets the name of the networkInstance in the spec
func (r *obj) SetNetworkInstanceName(s string) error {
	return r.p.SetNestedString(string(s), networkInstanceName...)
}

// SetAddressFamily sets the address family in the spec
func (r *obj) SetAddressFamily(s iputil.AddressFamily) error {
	return r.p.SetNestedString(string(s), addressFamily...)
}

// SetPrefix sets the prefix in the spec
func (r *obj) SetPrefix(s string) error {
	if _, err := iputil.New(s); err != nil {
		return err
	}
	return r.p.SetNestedString(string(s), addressFamily...)
}

// SetPrefixLength sets the prefix length in the spec
func (r *obj) SetPrefixLength(s uint8) error {
	return r.p.SetNestedInt(int(s), prefixLength...)
}

// SetIndex sets the index in the spec
func (r *obj) SetIndex(s uint32) error {
	return r.p.SetNestedInt(int(s), index...)
}

// SetSelectorLabels sets the selector matchLabels in the spec
func (r *obj) SetSelectorLabels(s map[string]string) error {
	return r.p.SetNestedMap(s, selectorLabels...)
}

// SetSpecLabels sets the labels in the spec
func (r *obj) SetSpecLabels(s map[string]string) error {
	return r.p.SetNestedMap(s, labels...)
}

// SetCreatePrefix sets the create prefix in the spec
func (r *obj) SetCreatePrefix(s bool) error {
	return r.p.SetNestedBool(s, createPrefix...)
}

// SetAllocatedPrefix sets the allocated prefix in the status
func (r *obj) SetAllocatedPrefix(s string) error {
	if _, err := iputil.New(s); err != nil {
		return err
	}
	return r.p.SetNestedString(s, createPrefix...)
}

// SetAllocatedGateway sets the allocated gateway in the status
func (r *obj) SetAllocatedGateway(s string) error {
	if _, err := iputil.New(s); err != nil {
		return err
	}
	return r.p.SetNestedString(s, createPrefix...)
}

// DeleteAddressFamily deletes the address family from the spec
func (r *obj) DeleteAddressFamily() error {
	return r.p.DeleteNestedField(addressFamily...)
}

// DeletePrefix deletes the prefix from the spec
func (r *obj) DeletePrefix() error {
	return r.p.DeleteNestedField(prefix...)
}

// DeletePrefixLength deletes the prefix length from the spec
func (r *obj) DeletePrefixLength() error {
	return r.p.DeleteNestedField(prefixLength...)
}

// DeleteIndex deletes the index from the spec
func (r *obj) DeleteIndex() error {
	return r.p.DeleteNestedField(index...)
}

// DeleteSpecLabels deletes the labels from the spec
func (r *obj) DeleteSpecLabels() error {
	return r.p.DeleteNestedField(labels...)
}

// DeleteCreatePrefix deletes the create prefix from the spec
func (r *obj) DeleteCreatePrefix() error {
	return r.p.DeleteNestedField(createPrefix...)
}

// DeleteAllocatedPrefix deletes the allocated prefix from the status
func (r *obj) DeleteAllocatedPrefix() error {
	return r.p.DeleteNestedField(allocatedPrefix...)
}

// DeleteAllocatedGateway deletes the allocated gateway from the status
func (r *obj) DeleteAllocatedGateway() error {
	return r.p.DeleteNestedField(allocatedGateway...)
}
