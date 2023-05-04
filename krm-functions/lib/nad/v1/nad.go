/*
Copyright 2023 Nephio.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"encoding/json"
	"fmt"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	nadv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	"github.com/nephio-project/nephio/krm-functions/lib/kubeobject"
)

const (
	errKubeObjectNotInitialized = "KubeObject not initialized"
	CniVersion                  = "0.3.1"
	ModeBridge                  = "bridge"
	ModeL2                      = "l2"
	NadType                     = "static"
	TuningType                  = "tuning"
)

var (
	ConfigType = []string{"spec", "config"}
)

type NadConfig struct {
	CniVersion string          `json:"cniVersion,omitempty"`
	Vlan       int             `json:"vlan,omitempty"`
	Plugins    []PluginCniType `json:"plugins,omitempty"`
}

type PluginCniType struct {
	Type         string       `json:"type,omitempty"`
	Capabilities Capabilities `json:"capabilities,omitempty"`
	Master       string       `json:"master,omitempty"`
	Mode         string       `json:"mode,omitempty"`
	Ipam         Ipam         `json:"ipam,omitempty"`
}

type Capabilities struct {
	Ips bool `json:"ips,omitempty"`
	Mac bool `json:"mac,omitempty"`
}

type Ipam struct {
	Type      string      `json:"type,omitempty"`
	Addresses []Addresses `json:"addresses,omitempty"`
}

type Addresses struct {
	Address string `json:"address,omitempty"`
	Gateway string `json:"gateway,omitempty"`
}

type CniSpecType int64

const (
	BothIpamVlan CniSpecType = iota // 0
	VlanAllocOnly
	IpVlanType
	SriovType
	MacVlanType
)

type NadStruct struct {
	K           kubeobject.KubeObjectExt[nadv1.NetworkAttachmentDefinition]
	CniSpecType CniSpecType
}

// NewFromKubeObject creates a new parser interface
// It expects a *fn.KubeObject as input representing the serialized yaml file
func NewFromKubeObject(b *fn.KubeObject) (*NadStruct, error) {
	p, err := kubeobject.NewFromKubeObject[nadv1.NetworkAttachmentDefinition](b)
	if err != nil {
		return nil, err
	}
	return &NadStruct{K: *p}, nil
}

// NewFromYAML creates a new parser interface
// It expects a raw byte slice as input representing the serialized yaml file
func NewFromYAML(b []byte) (*NadStruct, error) {
	p, err := kubeobject.NewFromYaml[nadv1.NetworkAttachmentDefinition](b)
	if err != nil {
		return nil, err
	}
	return &NadStruct{K: *p}, nil
}

// NewFromGoStruct creates a new parser interface
// It expects a go struct representing the interface krm resource
func NewFromGoStruct(b *nadv1.NetworkAttachmentDefinition) (*NadStruct, error) {
	p, err := kubeobject.NewFromGoStruct[nadv1.NetworkAttachmentDefinition](*b)
	if err != nil {
		return nil, err
	}
	return &NadStruct{K: *p}, nil
}

func (r *NadStruct) GetStringValue(fields ...string) string {
	if r == nil {
		return ""
	}
	s, ok, err := r.K.NestedString(fields...)
	if err != nil {
		return ""
	}
	if !ok {
		return ""
	}
	return s
}

func (r *NadStruct) GetConfigSpec() string {
	if r == nil {
		return ""
	}
	return r.GetStringValue(ConfigType...)
}

func (r *NadStruct) GetExistingNadConfig() NadConfig {
	nadConfigStruct := NadConfig{}
	if err := json.Unmarshal([]byte(r.GetStringValue(ConfigType...)), &nadConfigStruct); err != nil {
		return nadConfigStruct
	}
	return nadConfigStruct
}

func (r *NadStruct) GetCNIType() string {
	for i, plugin := range r.GetExistingNadConfig().Plugins {
		if plugin.Type == TuningType {
			continue
		} else {
			return r.GetExistingNadConfig().Plugins[i].Type
		}
	}
	return ""
}

func (r *NadStruct) GetVlan() int {
	return r.GetExistingNadConfig().Vlan
}

func (r *NadStruct) GetNadMaster() string {
	for i, plugin := range r.GetExistingNadConfig().Plugins {
		if plugin.Type == TuningType {
			continue
		} else {
			return r.GetExistingNadConfig().Plugins[i].Master
		}
	}
	return ""
}

func (r *NadStruct) GetIpamAddress() []Addresses {
	for i, plugin := range r.GetExistingNadConfig().Plugins {
		if plugin.Type == TuningType {
			continue
		} else {
			return r.GetExistingNadConfig().Plugins[i].Ipam.Addresses
		}
	}
	return []Addresses{}
}

func (r *NadStruct) GetNewNadConfig() (NadConfig, error) {
	nadConfigStruct := NadConfig{}
	configSpec := r.GetConfigSpec()
	if configSpec == "" {
		configSpec = "{}"
	}
	if err := json.Unmarshal([]byte(configSpec), &nadConfigStruct); err != nil {
		return nadConfigStruct, fmt.Errorf("invalid NAD Config, %s", err)
	}
	nadConfigStruct.CniVersion = CniVersion
	if r.CniSpecType == VlanAllocOnly {
		return nadConfigStruct, nil
	} else if r.CniSpecType == IpVlanType {
		if nadConfigStruct.Plugins == nil || len(nadConfigStruct.Plugins) == 0 {
			nadConfigStruct.Plugins = []PluginCniType{
				{
					Capabilities: Capabilities{
						Ips: true,
					},
					Mode: ModeL2,
					Ipam: Ipam{
						Type: NadType,
						Addresses: []Addresses{
							{},
						},
					},
				},
			}
		}
		return nadConfigStruct, nil
	} else if r.CniSpecType == MacVlanType {
		if nadConfigStruct.Plugins == nil || len(nadConfigStruct.Plugins) == 0 {
			nadConfigStruct.Plugins = []PluginCniType{
				{
					Capabilities: Capabilities{Ips: true},
					Mode:         ModeBridge,
					Ipam: Ipam{
						Type: NadType,
						Addresses: []Addresses{
							{},
						},
					},
				},
				{
					Capabilities: Capabilities{Mac: true},
					Type:         TuningType,
				},
			}
		}
		return nadConfigStruct, nil
	} else if r.CniSpecType == SriovType {
		if nadConfigStruct.Plugins == nil || len(nadConfigStruct.Plugins) == 0 {
			nadConfigStruct.Plugins = []PluginCniType{
				{
					Capabilities: Capabilities{Ips: true},
					Mode:         ModeBridge,
					Ipam: Ipam{
						Type: NadType,
						Addresses: []Addresses{
							{},
						},
					},
				},
			}
		}
		return nadConfigStruct, nil
	} else if r.CniSpecType == BothIpamVlan {
		if nadConfigStruct.Plugins == nil || len(nadConfigStruct.Plugins) == 0 {
			nadConfigStruct.Plugins = []PluginCniType{
				{
					Capabilities: Capabilities{
						Ips: true,
					},
					Mode: ModeBridge,
					Ipam: Ipam{
						Type: NadType,
						Addresses: []Addresses{
							{},
						},
					},
				},
			}
		}
		return nadConfigStruct, nil
	}
	return nadConfigStruct, nil
}

// SetConfigSpec sets the spec attributes in the kubeobject according the go struct
func (r *NadStruct) SetConfigSpec(spec *nadv1.NetworkAttachmentDefinitionSpec) error {
	return r.K.SetNestedString(spec.Config, ConfigType...)
}

func (r *NadStruct) SetNadConfig(config NadConfig) error {
	b, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return r.K.SetNestedString(string(b), ConfigType...)
}

func (r *NadStruct) SetCNIType(cniType string) error {
	if cniType != "" {
		if cniType == "ipvlan" {
			r.CniSpecType = IpVlanType
		} else if cniType == "macvlan" {
			r.CniSpecType = MacVlanType
		} else if cniType == "sriov" {
			r.CniSpecType = SriovType
		}
		nadConfigStruct, err := r.GetNewNadConfig()
		if err != nil {
			return err
		}
		for i, plugin := range nadConfigStruct.Plugins {
			if plugin.Type == TuningType {
				continue
			} else {
				nadConfigStruct.Plugins[i].Type = cniType
			}
		}
		return r.SetNadConfig(nadConfigStruct)
	} else {
		return fmt.Errorf("unknown cniType")
	}
}

func (r *NadStruct) SetVlan(vlanType int) error {
	if vlanType != 0 {
		nadConfigStruct, err := r.GetNewNadConfig()
		if err != nil {
			return err
		}
		nadConfigStruct.Vlan = vlanType
		return r.SetNadConfig(nadConfigStruct)
	} else {
		return fmt.Errorf("unknown vlanType")
	}
}

func (r *NadStruct) SetNadMaster(nadMaster string) error {
	if nadMaster != "" {
		nadConfigStruct, err := r.GetNewNadConfig()
		if err != nil {
			return err
		}
		for i, plugin := range nadConfigStruct.Plugins {
			if plugin.Type == TuningType {
				continue
			} else {
				nadConfigStruct.Plugins[i].Master = nadMaster
			}
		}
		return r.SetNadConfig(nadConfigStruct)
	} else {
		return fmt.Errorf("unknown nad master interface")
	}
}

func (r *NadStruct) SetIpamAddress(ipam []Addresses) error {
	if ipam != nil {
		nadConfigStruct, err := r.GetNewNadConfig()
		if err != nil {
			return err
		}
		for i, plugin := range nadConfigStruct.Plugins {
			if plugin.Type == TuningType {
				continue
			} else {
				nadConfigStruct.Plugins[i].Ipam.Addresses = ipam
			}
		}
		return r.SetNadConfig(nadConfigStruct)
	} else {
		return fmt.Errorf("unknown IPAM address")
	}
}

func (n *NadConfig) String() string {
	b, err := json.Marshal(n)
	if err != nil {
		panic(err)
	}
	return string(b)
}
