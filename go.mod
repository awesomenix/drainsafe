module github.com/awesomenix/drainsafe

go 1.12

require (
	github.com/awesomenix/repairman v0.0.0-20190704044539-e6ebab8c2993
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.3 // indirect
	github.com/stretchr/testify v1.3.0
	k8s.io/api v0.0.0
	k8s.io/apiextensions-apiserver v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/cli-runtime v0.0.0
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/kubernetes v1.15.3
	sigs.k8s.io/controller-runtime v0.2.0
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190819141258-3544db3b9e44
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190819143637-0dbe462fe92d
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190819142446-92cc630367d0
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190819144027-541433d7ce35
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190819141724-e14f31a72a77
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190819145148-d91c85d212d5
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20190819145008-029dd04813af
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190612205613-18da4a14b22b
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190819141909-f0f7c184477d
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190817025403-3ae76f584e79
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20190819145328-4831a4ced492
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190819142756-13daafd3604f
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20190819144832-f53437941eef
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20190819144346-2e47de1df0f0
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20190819144657-d1a724e0828e
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20190819144524-827174bad5e8
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20190819145509-592c9a46fd00
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190819143841-305e1cef1ab1
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20190819143045-c84c31c165c4
)
