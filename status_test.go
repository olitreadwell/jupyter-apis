package main

import (
	"testing"
	"time"

	kubeflowv1 "github.com/StatCan/kubeflow-apis/apis/kubeflow/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetStoppedStatus(t *testing.T) {
	stopped := &kubeflowv1.Notebook{
		ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{STOP_ANNOTATION: "x"}},
	}

	nb := stopped.DeepCopy()
	nb.Status.ReadyReplicas = 0
	phase, msg, key := getStoppedStatus(nb)
	if phase != NotebookPhaseStopped || msg != statusMessages[noPodsRunning] || key != noPodsRunning {
		t.Errorf("stopped -> (%q, %q, %q), want (%q, %q, %q)", phase, msg, key, NotebookPhaseStopped, statusMessages[noPodsRunning], noPodsRunning)
	}

	nb = stopped.DeepCopy()
	nb.Status.ReadyReplicas = 1
	phase, msg, key = getStoppedStatus(nb)
	if phase != NotebookPhaseWaiting || msg != statusMessages[notebookStopping] || key != notebookStopping {
		t.Errorf("stopping -> (%q, %q, %q), want (%q, %q, %q)", phase, msg, key, NotebookPhaseWaiting, statusMessages[notebookStopping], notebookStopping)
	}

	running := &kubeflowv1.Notebook{}
	phase, msg, key = getStoppedStatus(running)
	if phase != "" || msg != "" || key != "" {
		t.Errorf("running notebook -> (%q, %q, %q), want empty", phase, msg, key)
	}
}

func TestGetDeletedStatus(t *testing.T) {
	now := metav1.Now()
	del := &kubeflowv1.Notebook{ObjectMeta: metav1.ObjectMeta{DeletionTimestamp: &now}}
	phase, msg, key := getDeletedStatus(del)
	if phase != NotebookPhaseTerminating || msg != statusMessages[notebookDeleting] || key != notebookDeleting {
		t.Errorf("deleting -> (%q, %q, %q), want (%q, %q, %q)", phase, msg, key, NotebookPhaseTerminating, statusMessages[notebookDeleting], notebookDeleting)
	}

	phase, msg, key = getDeletedStatus(&kubeflowv1.Notebook{})
	if phase != "" || msg != "" || key != "" {
		t.Errorf("live notebook -> (%q, %q, %q), want empty", phase, msg, key)
	}
}

func TestCheckReadyNotebook(t *testing.T) {
	ready := &kubeflowv1.Notebook{}
	ready.Status.ReadyReplicas = 1
	phase, msg, key := checkReadyNotebook(ready)
	if phase != NotebookPhaseReady || msg != statusMessages[running] || key != running {
		t.Errorf("ready -> (%q, %q, %q), want (%q, %q, %q)", phase, msg, key, NotebookPhaseReady, statusMessages[running], running)
	}

	notReady := &kubeflowv1.Notebook{}
	notReady.Status.ReadyReplicas = 0
	if phase, msg, key := checkReadyNotebook(notReady); phase != "" || msg != "" || key != "" {
		t.Errorf("not ready -> (%q, %q, %q), want empty", phase, msg, key)
	}
}

func TestGetStatusFromContainerState(t *testing.T) {
	initializing := &kubeflowv1.Notebook{}
	initializing.Status.ContainerState = corev1.ContainerState{
		Waiting: &corev1.ContainerStateWaiting{Reason: "PodInitializing"},
	}
	phase, msg, key := getStatusFromContainerState(initializing)
	if phase != NotebookPhaseWaiting || msg != statusMessages[schedulingPod] || key != schedulingPod {
		t.Errorf("initializing -> (%q, %q, %q), want (%q, %q, %q)", phase, msg, key, NotebookPhaseWaiting, statusMessages[schedulingPod], schedulingPod)
	}

	warning := &kubeflowv1.Notebook{}
	warning.Status.ContainerState = corev1.ContainerState{
		Waiting: &corev1.ContainerStateWaiting{Reason: "ImagePullBackOff", Message: "pull failed"},
	}
	phase, msg, key = getStatusFromContainerState(warning)
	if phase != NotebookPhaseWarning || msg != "ImagePullBackOff : pull failed" || key != "" {
		t.Errorf("waiting-reason -> (%q, %q, %q), want (warning, %q, \"\")", phase, msg, key, "ImagePullBackOff : pull failed")
	}

	noWaiting := &kubeflowv1.Notebook{}
	if phase, msg, key := getStatusFromContainerState(noWaiting); phase != "" || msg != "" || key != "" {
		t.Errorf("no container state -> (%q, %q, %q), want empty", phase, msg, key)
	}
}

func TestGetStatusFromConditions(t *testing.T) {
	withReason := &kubeflowv1.Notebook{}
	withReason.Status.Conditions = []kubeflowv1.NotebookCondition{{Type: "Running", Reason: "Provisioning"}}
	phase, msg, key := getStatusFromConditions(withReason)
	if phase != NotebookPhaseWarning || msg != statusMessages[errorCondition] || key != errorCondition {
		t.Errorf("condition reason -> (%q, %q, %q), want (%q, %q, %q)", phase, msg, key, NotebookPhaseWarning, statusMessages[errorCondition], errorCondition)
	}

	empty := &kubeflowv1.Notebook{}
	if phase, msg, key := getStatusFromConditions(empty); phase != "" || msg != "" || key != "" {
		t.Errorf("no conditions -> (%q, %q, %q), want empty", phase, msg, key)
	}
}

func TestGetStatusFromEvents(t *testing.T) {
	warning := []*corev1.Event{{Type: corev1.EventTypeWarning}}
	phase, msg, key := getStatusFromEvents(warning)
	if phase != NotebookPhaseWarning || msg != statusMessages[errorEvent] || key != errorEvent {
		t.Errorf("warning event -> (%q, %q, %q), want (%q, %q, %q)", phase, msg, key, NotebookPhaseWarning, statusMessages[errorEvent], errorEvent)
	}

	normal := []*corev1.Event{{Type: corev1.EventTypeNormal}}
	if phase, msg, key := getStatusFromEvents(normal); phase != "" || msg != "" || key != "" {
		t.Errorf("normal event -> (%q, %q, %q), want empty", phase, msg, key)
	}

	if phase, msg, key := getStatusFromEvents(nil); phase != "" || msg != "" || key != "" {
		t.Errorf("no events -> (%q, %q, %q), want empty", phase, msg, key)
	}
}

func TestGetEmptyStatus(t *testing.T) {
	// period is not treated as newly-created.
	old := &kubeflowv1.Notebook{ObjectMeta: metav1.ObjectMeta{
		CreationTimestamp: metav1.NewTime(time.Now().Add(-time.Hour)),
	}}
	if phase, msg, key := getEmptyStatus(old); phase != "" || msg != "" || key != "" {
		t.Errorf("old notebook -> (%q, %q, %q), want empty", phase, msg, key)
	}

	reporting := &kubeflowv1.Notebook{ObjectMeta: metav1.ObjectMeta{
		CreationTimestamp: metav1.NewTime(time.Now()),
	}}
	reporting.Status.ContainerState = corev1.ContainerState{
		Waiting: &corev1.ContainerStateWaiting{Reason: "PodInitializing"},
	}
	if phase, msg, key := getEmptyStatus(reporting); phase != "" || msg != "" || key != "" {
		t.Errorf("reporting notebook -> (%q, %q, %q), want empty", phase, msg, key)
	}
}
