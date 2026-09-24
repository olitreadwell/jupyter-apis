package main

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func strPtr(s string) *string { return &s }

func newPvc(name string, modes []corev1.PersistentVolumeAccessMode, storage string) NewPvc {
	return NewPvc{
		NewPvcMetadata: NewPvcMetadata{Name: strPtr(name)},
		NewPvcSpec: NewPvcSpec{
			AccessModes: modes,
			Resources:   Resources{Requests: Requests{Storage: resource.MustParse(storage)}},
		},
	}
}

func TestValidateNotebookResources(t *testing.T) {
	tests := []struct {
		name        string
		cpu         string
		cpuLimit    string
		memory      string
		memoryLimit string
		wantErrors  []string
	}{
		{name: "valid", cpu: "1", cpuLimit: "2", memory: "1Gi", memoryLimit: "2Gi"},
		{name: "zero cpu", cpu: "0", cpuLimit: "2", memory: "1Gi", memoryLimit: "2Gi", wantErrors: []string{"cpu must be positive"}},
		{name: "zero memory", cpu: "1", cpuLimit: "2", memory: "0", memoryLimit: "2Gi", wantErrors: []string{"memory must be positive"}},
		{name: "cpu limit below request", cpu: "2", cpuLimit: "1", memory: "1Gi", memoryLimit: "2Gi", wantErrors: []string{"cpu limit must be set and CPU limit must be greater than or equal to requested CPU"}},
		{name: "memory limit below request", cpu: "1", cpuLimit: "2", memory: "2Gi", memoryLimit: "1Gi", wantErrors: []string{"memory limit must be set and Memory limit must be greater than or equal to requested memory"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateNotebookResources(
				resource.MustParse(tt.cpu),
				resource.MustParse(tt.cpuLimit),
				resource.MustParse(tt.memory),
				resource.MustParse(tt.memoryLimit),
			)
			if len(got) != len(tt.wantErrors) {
				t.Fatalf("got %d errors %v, want %d errors %v", len(got), got, len(tt.wantErrors), tt.wantErrors)
			}
			for i, want := range tt.wantErrors {
				if got[i] != want {
					t.Errorf("error %d = %q, want %q", i, got[i], want)
				}
			}
		})
	}
}

func TestValidateNotebookVolume(t *testing.T) {
	tests := []struct {
		name string
		req  volrequest
		want string
	}{
		{name: "no volume", req: volrequest{}, want: ""},
		{name: "missing mount", req: volrequest{NewPvc: newPvc("vol", nil, "16Gi")}, want: "mount path is required"},
		{name: "invalid mount path", req: volrequest{Mount: "/tmp", NewPvc: newPvc("vol", nil, "16Gi")}, want: "mount path must be /home/jovyan or any of its subdirectories"},
		{name: "both existing and new", req: volrequest{Mount: "/home/jovyan", ExistingSource: ExistingSource{PersistentVolumeClaim: PersistentVolumeClaim{ClaimName: strPtr("pvc")}}, NewPvc: newPvc("vol", nil, "16Gi")}, want: "only one existing volume or new volume should be provided"},
		{name: "neither existing nor new", req: volrequest{Mount: "/home/jovyan"}, want: "either existing volume or new volume must be provided"},
		{name: "empty existing name", req: volrequest{Mount: "/home/jovyan", ExistingSource: ExistingSource{PersistentVolumeClaim: PersistentVolumeClaim{ClaimName: strPtr("")}}}, want: "existing volume name cannot be empty"},
		{name: "invalid existing name", req: volrequest{Mount: "/home/jovyan", ExistingSource: ExistingSource{PersistentVolumeClaim: PersistentVolumeClaim{ClaimName: strPtr("Bad_Name")}}}, want: "volume name must be one or more lowercase alphanumeric labels"},
		{name: "empty new name", req: volrequest{Mount: "/home/jovyan", NewPvc: newPvc("", nil, "16Gi")}, want: "new volume name cannot be empty"},
		{name: "no access modes", req: volrequest{Mount: "/home/jovyan", NewPvc: newPvc("vol", nil, "16Gi")}, want: "volume accessModes must have at least one value"},
		{name: "invalid access mode", req: volrequest{Mount: "/home/jovyan", NewPvc: newPvc("vol", []corev1.PersistentVolumeAccessMode{"ReadWriteMany"}, "16Gi")}, want: "volume accessModes must be ReadWriteOnce"},
		{name: "invalid storage size", req: volrequest{Mount: "/home/jovyan", NewPvc: newPvc("vol", []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"}, "3Gi")}, want: "storage request is invalid"},
		{name: "valid new volume", req: volrequest{Mount: "/home/jovyan", NewPvc: newPvc("vol", []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"}, "16Gi")}, want: ""},
		{name: "valid existing volume", req: volrequest{Mount: "/home/jovyan", ExistingSource: ExistingSource{PersistentVolumeClaim: PersistentVolumeClaim{ClaimName: strPtr("my-pvc")}}}, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNotebookVolume(tt.req)
			if tt.want == "" {
				if err != nil {
					t.Errorf("validateNotebookVolume() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("validateNotebookVolume() = nil, want error containing %q", tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("validateNotebookVolume() error = %q, want containing %q", err.Error(), tt.want)
			}
		})
	}
}

func TestValidateNotebook(t *testing.T) {
	valid := newnotebookrequest{
		Name:            "my-notebook",
		Namespace:       "default",
		Image:           "jupyter/minimal:latest",
		CPU:             resource.MustParse("1"),
		CPULimit:        resource.MustParse("2"),
		Memory:          resource.MustParse("1Gi"),
		MemoryLimit:     resource.MustParse("2Gi"),
		ImagePullPolicy: "Always",
	}
	with := func(mut func(*newnotebookrequest)) newnotebookrequest {
		r := valid
		mut(&r)
		return r
	}

	tests := []struct {
		name string
		req  newnotebookrequest
		want string
	}{
		{name: "valid", req: valid, want: ""},
		{name: "missing name", req: with(func(r *newnotebookrequest) { r.Name = "" }), want: "name is required"},
		{name: "invalid name", req: with(func(r *newnotebookrequest) { r.Name = "Bad_Name" }), want: "name must consist of lowercase alphanumeric characters"},
		{name: "missing namespace", req: with(func(r *newnotebookrequest) { r.Namespace = "" }), want: "namespace is required"},
		{name: "missing image", req: with(func(r *newnotebookrequest) { r.Image = "" }), want: "either Image must be provided or CustomImageCheck must be true"},
		{name: "custom image missing", req: with(func(r *newnotebookrequest) { r.Image = ""; r.CustomImageCheck = true }), want: "customImage is required when CustomImageCheck is true"},
		{name: "invalid image pull policy", req: with(func(r *newnotebookrequest) { r.ImagePullPolicy = "IfNotPresent" }), want: "invalid ImagePullPolicy"},
		{name: "invalid server type", req: with(func(r *newnotebookrequest) { r.ServerType = "spark" }), want: "invalid ServerType"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNotebook(tt.req)
			if tt.want == "" {
				if err != nil {
					t.Errorf("validateNotebook() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("validateNotebook() = nil, want error containing %q", tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("validateNotebook() error = %q, want containing %q", err.Error(), tt.want)
			}
		})
	}
}

func TestValidateUpdateNotebook(t *testing.T) {
	valid := updatenotebookrequest{
		CPU:         resource.MustParse("1"),
		CPULimit:    resource.MustParse("2"),
		Memory:      resource.MustParse("1Gi"),
		MemoryLimit: resource.MustParse("2Gi"),
	}
	withDataVol := func(r *updatenotebookrequest) {
		r.DataVolumes = []volrequest{{Mount: "/home/jovyan/data", NewPvc: newPvc("vol", []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"}, "16Gi")}}
	}

	tests := []struct {
		name string
		req  updatenotebookrequest
		want string
	}{
		{name: "valid", req: valid, want: ""},
		{name: "zero cpu", req: func() updatenotebookrequest { r := valid; r.CPU = resource.MustParse("0"); return r }(), want: "cpu must be positive"},
		{name: "zero cpu with data volume", req: func() updatenotebookrequest { r := valid; r.CPU = resource.MustParse("0"); withDataVol(&r); return r }(), want: "cpu must be positive"},
		{name: "invalid data volume", req: func() updatenotebookrequest { r := valid; withDataVol(&r); r.DataVolumes[0].Mount = "/tmp"; return r }(), want: "mount path must be /home/jovyan"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUpdateNotebook(tt.req)
			if tt.want == "" {
				if err != nil {
					t.Errorf("validateUpdateNotebook() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("validateUpdateNotebook() = nil, want error containing %q", tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("validateUpdateNotebook() error = %q, want containing %q", err.Error(), tt.want)
			}
		})
	}
}
