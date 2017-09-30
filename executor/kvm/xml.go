package kvm

import (
	"bytes"
	"encoding/xml"
	"html/template"
	"io/ioutil"
)

var (
	baseXmlTmplate = `<domain type='kvm'>
        <name>{{.Name}}</name>
        <memory unit='MiB'>{{.Memory}}</memory>
        <vcpu>{{.Cpus}}</vcpu>
        <os>
		<type arch='x86_64' machine='pc'>hvm</type>
		<boot dev='fd'/>
		<boot dev='hd'/>
		<boot dev='cdrom'/>
		<boot dev='network'/>
		<bootmenu enable='yes' timeout='3000'/>
       </os>
       <features>
		<acpi/>
		<apic/>
		<pae/>
       </features>
       <clock offset='localtime'/>
       <on_poweroff>destroy</on_poweroff>
       <on_reboot>restart</on_reboot>
       <on_crash>destroy</on_crash>
       <devices>
		<emulator>/usr/libexec/qemu-kvm</emulator>

		<disk type='file' device='disk'>
			<driver name='qemu' type='qcow2'/>
			<source file='/var/lib/libvirt/images/{{.Name}}.qcow2'/>
			<target dev='hda' bus='ide'/>
		</disk>

		<disk type='file' device='cdrom'>
			<source file='/data/iso/{{.Iso}}'/>
			<target dev='hdb' bus='ide'/>
		</disk>

		<interface type='bridge'>
			<source bridge='virbr0'/>
		</interface>

		<input type='mouse' bus='ps2'/>

		<graphics type='vnc' port='{{.VncPort}}' passwd='{{.VncPassword}}' sharePolicy='allow-exclusive'>
    			<listen type='address' address='0.0.0.0'/>
  		</graphics>
       </devices>
</domain>
`
)

type KvmDomainOpts struct {
	Name        string
	Memory      uint // by MiB
	Cpus        int
	Iso         string
	VncPort     string
	VncPassword string
}

func (opts *KvmDomainOpts) Valid() error {
	return nil
}

func NewKvmDomain(opts *KvmDomainOpts) ([]byte, error) {
	if err := opts.Valid(); err != nil {
		return nil, err
	}

	tmpl, err := template.New("").Parse(baseXmlTmplate)
	if err != nil {
		return nil, err
	}

	var buf = bytes.NewBuffer(nil)
	if err = tmpl.Execute(buf, opts); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

//
// TODO
// use followings xml definations to replace above templates
//
//

// Domain is an libvirt xml domain xml config object
// See More: xml-example/example.xml
type Domain struct {
	XMLName    xml.Name           `xml:"domain"`
	Name       string             `xml:"name"`
	Type       string             `xml:"type,attr,omitempty"`
	Memory     *DomainMemory      `xml:"memory"`
	VCPU       *DomainVCPU        `xml:"vcpu"`
	OS         *DomainOS          `xml:"os"`
	Features   *DomainFeatureList `xml:"features"`
	Clock      *DomainClock       `xml:"clock,omitempty"`
	OnPoweroff string             `xml:"on_poweroff,omitempty"`
	OnReboot   string             `xml:"on_reboot,omitempty"`
	OnCrash    string             `xml:"on_crash,omitempty"`
	Devices    *DomainDeviceList  `xml:"devices"`
}

func (d *Domain) dumpAsFile(fileName string) error {
	bs, err := xml.MarshalIndent(d, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fileName, bs, 0644)
}

func NewKvmDomainNG(opts *KvmDomainOpts) (*Domain, error) {
	if err := opts.Valid(); err != nil {
		return nil, err
	}

	return &Domain{
		Name: opts.Name,
		Type: "kvm",
		Memory: &DomainMemory{
			Unit:  "MiB",
			Value: opts.Memory,
		},
		VCPU: &DomainVCPU{
			Value: opts.Cpus,
		},
		OS: &DomainOS{
			Type: &DomainOSType{
				Arch:    "x86_64",
				Machine: "pc",
				Type:    "hvm",
			},
			BootDevices: []DomainBootDevice{
				{Dev: "fd"},
				{Dev: "hd"},
				{Dev: "cdrom"},
				{Dev: "network"},
			},
			BootMenu: &DomainBootMenu{
				Enabled: "yes",
				Timeout: "3000",
			},
		},
		Features: &DomainFeatureList{
			PAE:  struct{}{},
			ACPI: struct{}{},
			APIC: struct{}{},
		},
		Clock: &DomainClock{
			Offset: "localtime",
		},
		OnPoweroff: "destroy",
		OnReboot:   "restart",
		OnCrash:    "destroy",
		Devices:    &DomainDeviceList{}, // TODO un-finished
	}, nil
}

//
// All Sub Field Definations
//
type DomainMemory struct {
	Value uint   `xml:",chardata"`
	Unit  string `xml:"unit,attr,omitempty"`
}

type DomainVCPU struct {
	Value int `xml:",chardata"`
}

type DomainOS struct {
	Type        *DomainOSType      `xml:"type"`
	BootDevices []DomainBootDevice `xml:"boot"`
	BootMenu    *DomainBootMenu    `xml:"bootmenu"`
}

type DomainOSType struct {
	Arch    string `xml:"arch,attr,omitempty"`
	Machine string `xml:"machine,attr,omitempty"`
	Type    string `xml:",chardata"`
}

type DomainBootDevice struct {
	Dev string `xml:"dev,attr"`
}

type DomainBootMenu struct {
	Enabled string `xml:"enabled,attr"`
	Timeout string `xml:"timeout,attr"`
}

type DomainFeatureList struct {
	PAE  struct{} `xml:"pae"`
	ACPI struct{} `xml:"acpi"`
	APIC struct{} `xml:"apic"`
}

type DomainClock struct {
	Offset string `xml:"offset,attr,omitempty"`
}

type DomainDeviceList struct {
	Emulator   string            `xml:"emulator,omitempty"`
	Disks      []DomainDisk      `xml:"disk"`
	Interfaces []DomainInterface `xml:"interface"`
	Inputs     []DomainInput     `xml:"input"`
	Graphics   []DomainGraphic   `xml:"graphics"`
}

type DomainDisk struct {
	XMLName xml.Name          `xml:"disk"`
	Type    string            `xml:"type,attr"`
	Device  string            `xml:"device,attr"`
	Driver  *DomainDiskDriver `xml:"driver"`
	Source  *DomainDiskSource `xml:"source"`
	Target  *DomainDiskTarget `xml:"target"`
}

type DomainDiskDriver struct {
	Name string `xml:"name,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
}

type DomainDiskSource struct {
	File string `xml:"file,attr,omitempty"`
}

type DomainDiskTarget struct {
	Dev string `xml:"dev,attr,omitempty"`
	Bus string `xml:"bus,attr,omitempty"`
}

type DomainInterface struct {
	XMLName xml.Name               `xml:"interface"`
	Type    string                 `xml:"type,attr"`
	Source  *DomainInterfaceSource `xml:"source"`
	MAC     *DomainInterfaceMAC    `xml:"mac"` // TODO
}

type DomainInterfaceSource struct {
	Bridge string `xml:"bridge,attr,omitempty"`
}

type DomainInterfaceMAC struct {
	Address string `xml:"address,attr"`
}

type DomainInput struct {
	XMLName xml.Name `xml:"input"`
	Type    string   `xml:"type,attr"`
	Bus     string   `xml:"bus,attr"`
}

type DomainGraphic struct {
	XMLName     xml.Name                `xml:"graphics"`
	Type        string                  `xml:"type,attr"`
	Port        int                     `xml:"port,attr,omitempty"`
	SharePolicy string                  `xml:"sharePolicy,attr,omitempty"`
	Listeners   []DomainGraphicListener `xml:"listen"`
}

type DomainGraphicListener struct {
	Type    string `xml:"type,attr"`
	Address string `xml:"address,attr,omitempty"`
}
