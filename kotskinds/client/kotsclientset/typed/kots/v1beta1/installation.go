/*
Copyright 2019 Replicated, Inc..

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
// Code generated by client-gen. DO NOT EDIT.

package v1beta1

import (
	"context"
	"time"

	v1beta1 "github.com/replicatedhq/kots/kotskinds/apis/kots/v1beta1"
	scheme "github.com/replicatedhq/kots/kotskinds/client/kotsclientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// InstallationsGetter has a method to return a InstallationInterface.
// A group's client should implement this interface.
type InstallationsGetter interface {
	Installations(namespace string) InstallationInterface
}

// InstallationInterface has methods to work with Installation resources.
type InstallationInterface interface {
	Create(ctx context.Context, installation *v1beta1.Installation, opts v1.CreateOptions) (*v1beta1.Installation, error)
	Update(ctx context.Context, installation *v1beta1.Installation, opts v1.UpdateOptions) (*v1beta1.Installation, error)
	UpdateStatus(ctx context.Context, installation *v1beta1.Installation, opts v1.UpdateOptions) (*v1beta1.Installation, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1beta1.Installation, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1beta1.InstallationList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.Installation, err error)
	InstallationExpansion
}

// installations implements InstallationInterface
type installations struct {
	client rest.Interface
	ns     string
}

// newInstallations returns a Installations
func newInstallations(c *KotsV1beta1Client, namespace string) *installations {
	return &installations{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the installation, and returns the corresponding installation object, and an error if there is any.
func (c *installations) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.Installation, err error) {
	result = &v1beta1.Installation{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("installations").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Installations that match those selectors.
func (c *installations) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.InstallationList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1beta1.InstallationList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("installations").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested installations.
func (c *installations) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("installations").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a installation and creates it.  Returns the server's representation of the installation, and an error, if there is any.
func (c *installations) Create(ctx context.Context, installation *v1beta1.Installation, opts v1.CreateOptions) (result *v1beta1.Installation, err error) {
	result = &v1beta1.Installation{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("installations").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(installation).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a installation and updates it. Returns the server's representation of the installation, and an error, if there is any.
func (c *installations) Update(ctx context.Context, installation *v1beta1.Installation, opts v1.UpdateOptions) (result *v1beta1.Installation, err error) {
	result = &v1beta1.Installation{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("installations").
		Name(installation.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(installation).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *installations) UpdateStatus(ctx context.Context, installation *v1beta1.Installation, opts v1.UpdateOptions) (result *v1beta1.Installation, err error) {
	result = &v1beta1.Installation{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("installations").
		Name(installation.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(installation).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the installation and deletes it. Returns an error if one occurs.
func (c *installations) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("installations").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *installations) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("installations").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched installation.
func (c *installations) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.Installation, err error) {
	result = &v1beta1.Installation{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("installations").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
