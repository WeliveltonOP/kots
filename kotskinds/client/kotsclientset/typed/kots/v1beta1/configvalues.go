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

// ConfigValuesesGetter has a method to return a ConfigValuesInterface.
// A group's client should implement this interface.
type ConfigValuesesGetter interface {
	ConfigValueses(namespace string) ConfigValuesInterface
}

// ConfigValuesInterface has methods to work with ConfigValues resources.
type ConfigValuesInterface interface {
	Create(ctx context.Context, configValues *v1beta1.ConfigValues, opts v1.CreateOptions) (*v1beta1.ConfigValues, error)
	Update(ctx context.Context, configValues *v1beta1.ConfigValues, opts v1.UpdateOptions) (*v1beta1.ConfigValues, error)
	UpdateStatus(ctx context.Context, configValues *v1beta1.ConfigValues, opts v1.UpdateOptions) (*v1beta1.ConfigValues, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1beta1.ConfigValues, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1beta1.ConfigValuesList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.ConfigValues, err error)
	ConfigValuesExpansion
}

// configValueses implements ConfigValuesInterface
type configValueses struct {
	client rest.Interface
	ns     string
}

// newConfigValueses returns a ConfigValueses
func newConfigValueses(c *KotsV1beta1Client, namespace string) *configValueses {
	return &configValueses{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the configValues, and returns the corresponding configValues object, and an error if there is any.
func (c *configValueses) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.ConfigValues, err error) {
	result = &v1beta1.ConfigValues{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("configvalueses").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ConfigValueses that match those selectors.
func (c *configValueses) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.ConfigValuesList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1beta1.ConfigValuesList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("configvalueses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested configValueses.
func (c *configValueses) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("configvalueses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a configValues and creates it.  Returns the server's representation of the configValues, and an error, if there is any.
func (c *configValueses) Create(ctx context.Context, configValues *v1beta1.ConfigValues, opts v1.CreateOptions) (result *v1beta1.ConfigValues, err error) {
	result = &v1beta1.ConfigValues{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("configvalueses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(configValues).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a configValues and updates it. Returns the server's representation of the configValues, and an error, if there is any.
func (c *configValueses) Update(ctx context.Context, configValues *v1beta1.ConfigValues, opts v1.UpdateOptions) (result *v1beta1.ConfigValues, err error) {
	result = &v1beta1.ConfigValues{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("configvalueses").
		Name(configValues.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(configValues).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *configValueses) UpdateStatus(ctx context.Context, configValues *v1beta1.ConfigValues, opts v1.UpdateOptions) (result *v1beta1.ConfigValues, err error) {
	result = &v1beta1.ConfigValues{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("configvalueses").
		Name(configValues.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(configValues).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the configValues and deletes it. Returns an error if one occurs.
func (c *configValueses) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("configvalueses").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *configValueses) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("configvalueses").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched configValues.
func (c *configValueses) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.ConfigValues, err error) {
	result = &v1beta1.ConfigValues{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("configvalueses").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
