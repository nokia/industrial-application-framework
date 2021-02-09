package integration

import (
	"fmt"
	"net/url"

	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	"gitlabe2.ext.net.nokia.com/Nokia_DAaaS/edge-microservices/componenttest-lib/pkg/cttestingfw/integration/internal"
)

// NewTinyCA creates a new a tiny CA utility for provisioning serving certs and client certs FOR TESTING ONLY.
// Don't use this for anything else!
var NewTinyCA = internal.NewTinyCA

// ControlPlane is a struct that knows how to start your test control plane.
//
// Right now, that means Etcd and your APIServer. This is likely to increase in
// future.
type ControlPlane struct {
	APIServer *APIServer
	Etcd      *Etcd
}

// Start will start your control plane processes. To stop them, call Stop().
func (f *ControlPlane) Start() error {
	if f.Etcd == nil {
		f.Etcd = &Etcd{}
	}
	if err := f.Etcd.Start(); err != nil {
		return err
	}

	if f.APIServer == nil {
		f.APIServer = &APIServer{}
	}
	f.APIServer.EtcdURL = f.Etcd.URL
	return f.APIServer.Start()
}

// Stop will stop your control plane processes, and clean up their data.
func (f *ControlPlane) Stop() error {
	if f.APIServer != nil {
		if err := f.APIServer.Stop(); err != nil {
			return err
		}
	}
	if f.Etcd != nil {
		if err := f.Etcd.Stop(); err != nil {
			return err
		}
	}
	return nil
}

// APIURL returns the URL you should connect to to talk to your API.
func (f *ControlPlane) APIURL() *url.URL {
	return f.APIServer.URL
}

// KubeCtl returns a pre-configured KubeCtl, ready to connect to this
// ControlPlane.
func (f *ControlPlane) KubeCtl() *KubeCtl {
	k := &KubeCtl{}
	k.Opts = append(k.Opts, fmt.Sprintf("--server=%s", f.APIURL()))
	return k
}

// RESTClientConfig returns a pre-configured restconfig, ready to connect to
// this ControlPlane.
func (f *ControlPlane) RESTClientConfig() (*rest.Config, error) {
	c := &rest.Config{
		Host: f.APIURL().String(),
		ContentConfig: rest.ContentConfig{
			NegotiatedSerializer: serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs},
		},
	}
	err := rest.SetKubernetesDefaults(c)
	return c, err
}