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

package fake

import (
	"context"

	v1beta1 "github.com/replicatedhq/kots/kotskinds/apis/kots/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeIdentityConfigs implements IdentityConfigInterface
type FakeIdentityConfigs struct {
	Fake *FakeKotsV1beta1
	ns   string
}

var identityconfigsResource = schema.GroupVersionResource{Group: "kots.io", Version: "v1beta1", Resource: "identityconfigs"}

var identityconfigsKind = schema.GroupVersionKind{Group: "kots.io", Version: "v1beta1", Kind: "IdentityConfig"}

// Get takes name of the identityConfig, and returns the corresponding identityConfig object, and an error if there is any.
func (c *FakeIdentityConfigs) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.IdentityConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(identityconfigsResource, c.ns, name), &v1beta1.IdentityConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.IdentityConfig), err
}

// List takes label and field selectors, and returns the list of IdentityConfigs that match those selectors.
func (c *FakeIdentityConfigs) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.IdentityConfigList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(identityconfigsResource, identityconfigsKind, c.ns, opts), &v1beta1.IdentityConfigList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.IdentityConfigList{ListMeta: obj.(*v1beta1.IdentityConfigList).ListMeta}
	for _, item := range obj.(*v1beta1.IdentityConfigList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested identityConfigs.
func (c *FakeIdentityConfigs) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(identityconfigsResource, c.ns, opts))

}

// Create takes the representation of a identityConfig and creates it.  Returns the server's representation of the identityConfig, and an error, if there is any.
func (c *FakeIdentityConfigs) Create(ctx context.Context, identityConfig *v1beta1.IdentityConfig, opts v1.CreateOptions) (result *v1beta1.IdentityConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(identityconfigsResource, c.ns, identityConfig), &v1beta1.IdentityConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.IdentityConfig), err
}

// Update takes the representation of a identityConfig and updates it. Returns the server's representation of the identityConfig, and an error, if there is any.
func (c *FakeIdentityConfigs) Update(ctx context.Context, identityConfig *v1beta1.IdentityConfig, opts v1.UpdateOptions) (result *v1beta1.IdentityConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(identityconfigsResource, c.ns, identityConfig), &v1beta1.IdentityConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.IdentityConfig), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeIdentityConfigs) UpdateStatus(ctx context.Context, identityConfig *v1beta1.IdentityConfig, opts v1.UpdateOptions) (*v1beta1.IdentityConfig, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(identityconfigsResource, "status", c.ns, identityConfig), &v1beta1.IdentityConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.IdentityConfig), err
}

// Delete takes name of the identityConfig and deletes it. Returns an error if one occurs.
func (c *FakeIdentityConfigs) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(identityconfigsResource, c.ns, name, opts), &v1beta1.IdentityConfig{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeIdentityConfigs) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(identityconfigsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1beta1.IdentityConfigList{})
	return err
}

// Patch applies the patch and returns the patched identityConfig.
func (c *FakeIdentityConfigs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.IdentityConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(identityconfigsResource, c.ns, name, pt, data, subresources...), &v1beta1.IdentityConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.IdentityConfig), err
}
