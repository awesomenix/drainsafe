module github.com/awesomenix/drainsafe

go 1.13

require (
	github.com/awesomenix/repairman v0.0.0-20190704044539-e6ebab8c2993
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.3.0
	k8s.io/api v0.0.0
	k8s.io/apiextensions-apiserver v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/cli-runtime v0.0.0
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/component-base v0.0.0 // indirect
	k8s.io/kubectl v0.0.0
	sigs.k8s.io/controller-runtime v0.3.1-0.20191105233659-81842d0e78f7
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20191107030003-665c8a257c1a
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191107032734-f60a3abe8be9
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191107025710-670e6d490571
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20191107031416-60260b106f90
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20191107033206-d1f4c7562f79
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191107030346-a537b3b5272f
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20191107034400-7ce8bc796221
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20191107034217-39c51490b693
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20191107025440-35a828233ddd
	k8s.io/component-base => k8s.io/component-base v0.0.0-20191107031042-aea44d161014
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20191107035106-03d130a7dc28
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20191107034543-c8105f3abf3a
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20191107031811-3a3c0cc237b7
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20191107034037-3f4f28c46e7d
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20191107033534-0b34e978f7ad
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20191107033854-4ffc45aacb3f
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20191107035253-116445b61d51
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20191107033714-42ebcc85ab18
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20191107034732-430e6b919261
	k8s.io/metrics => k8s.io/metrics v0.0.0-20191107033012-d3ccb962eb82
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20191107032040-8f197c8b54f4
)
