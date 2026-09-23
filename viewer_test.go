package main

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestViewerStatus(t *testing.T) {
	// A zero value viewer has no identity yet.
	if phase := viewerStatus(pvcviewer{}); phase != NotebookPhaseUnitialized {
		t.Errorf("zero viewer -> %q, want %q", phase, NotebookPhaseUnitialized)
	}

	// A viewer with a deletion timestamp is terminating.
	deleting := pvcviewer{Metadata: map[string]string{"deletionTimestamp": "2026-09-24T00:00:00Z"}}
	if phase := viewerStatus(deleting); phase != NotebookPhaseTerminating {
		t.Errorf("deleting viewer -> %q, want %q", phase, NotebookPhaseTerminating)
	}

	// A viewer reporting ready is ready.
	ready := pvcviewer{Metadata: map[string]string{}, Status: map[string]string{"ready": "true"}}
	if phase := viewerStatus(ready); phase != NotebookPhaseReady {
		t.Errorf("ready viewer -> %q, want %q", phase, NotebookPhaseReady)
	}

	// Anything else is waiting.
	waiting := pvcviewer{Metadata: map[string]string{"name": "v"}, Status: map[string]string{"phase": "pending"}}
	if phase := viewerStatus(waiting); phase != NotebookPhaseWaiting {
		t.Errorf("waiting viewer -> %q, want %q", phase, NotebookPhaseWaiting)
	}
}

func TestGetOwningViewer(t *testing.T) {
	owned := corev1.Pod{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{POD_PARENT_VIEWER_LABEL_KEY: "viewer-a"}}}
	if got := getOwningViewer(owned); got != "viewer-a" {
		t.Errorf("owned pod -> %q, want %q", got, "viewer-a")
	}

	orphan := corev1.Pod{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}}}
	if got := getOwningViewer(orphan); got != "" {
		t.Errorf("orphan pod -> %q, want empty", got)
	}
}

func TestIsViewerPod(t *testing.T) {
	owned := corev1.Pod{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{POD_PARENT_VIEWER_LABEL_KEY: "viewer-a"}}}
	if !isViewerPod(owned) {
		t.Error("owned pod should be a viewer pod")
	}

	orphan := corev1.Pod{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "notebook"}}}
	if isViewerPod(orphan) {
		t.Error("orphan pod should not be a viewer pod")
	}
}
