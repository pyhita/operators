package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AppService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AppServiceSpec `json:"spec"`
}

type AppServiceSpec struct {
	Image  string            `json:"image"`
	Labels map[string]string `json:"labels"`
	Ports  []Port            `json:"ports"`
}

type Port struct {
	Port       int `json:"port"`
	TargetPort int `json:"targetPort"`
	NodePort   int `json:"nodePort"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AppServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []AppService `json:"items"`
}
