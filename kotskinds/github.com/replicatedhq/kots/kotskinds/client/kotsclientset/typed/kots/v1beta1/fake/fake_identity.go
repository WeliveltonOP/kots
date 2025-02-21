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

// FakeIdentities implements IdentityInterface
type FakeIdentities struct {
	Fake *FakeKotsV1beta1
	ns   string
}

var identitiesResource = schema.GroupVersionResource{Group: "kots.io", Version: "v1beta1", Resource: "identities"}

var identitiesKind = schema.GroupVersionKind{Group: "kots.io", Version: "v1beta1", Kind: "Identity"}

// Get takes name of the identity, and returns the corresponding identity object, and an error if there is any.
func (c *FakeIdentities) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.Identity, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(identitiesResource, c.ns, name), &v1beta1.Identity{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Identity), err
}

// List takes label and field selectors, and returns the list of Identities that match those selectors.
func (c *FakeIdentities) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.IdentityList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(identitiesResource, identitiesKind, c.ns, opts), &v1beta1.IdentityList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.IdentityList{ListMeta: obj.(*v1beta1.IdentityList).ListMeta}
	for _, item := range obj.(*v1beta1.IdentityList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested identities.
func (c *FakeIdentities) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(identitiesResource, c.ns, opts))

}

// Create takes the representation of a identity and creates it.  Returns the server's representation of the identity, and an error, if there is any.
func (c *FakeIdentities) Create(ctx context.Context, identity *v1beta1.Identity, opts v1.CreateOptions) (result *v1beta1.Identity, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(identitiesResource, c.ns, identity), &v1beta1.Identity{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Identity), err
}

// Update takes the representation of a identity and updates it. Returns the server's representation of the identity, and an error, if there is any.
func (c *FakeIdentities) Update(ctx context.Context, identity *v1beta1.Identity, opts v1.UpdateOptions) (result *v1beta1.Identity, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(identitiesResource, c.ns, identity), &v1beta1.Identity{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Identity), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeIdentities) UpdateStatus(ctx context.Context, identity *v1beta1.Identity, opts v1.UpdateOptions) (*v1beta1.Identity, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(identitiesResource, "status", c.ns, identity), &v1beta1.Identity{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Identity), err
}

// Delete takes name of the identity and deletes it. Returns an error if one occurs.
func (c *FakeIdentities) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(identitiesResource, c.ns, name, opts), &v1beta1.Identity{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeIdentities) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(identitiesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1beta1.IdentityList{})
	return err
}

// Patch applies the patch and returns the patched identity.
func (c *FakeIdentities) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.Identity, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(identitiesResource, c.ns, name, pt, data, subresources...), &v1beta1.Identity{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Identity), err
}
