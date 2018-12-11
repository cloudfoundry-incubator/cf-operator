/*

Don't alter this file, it was generated.

*/
// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "code.cloudfoundry.org/cf-operator/pkg/kube/apis/extendedstatefulset/v1alpha1"
	scheme "code.cloudfoundry.org/cf-operator/pkg/kube/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ExtendedStatefulSetsGetter has a method to return a ExtendedStatefulSetInterface.
// A group's client should implement this interface.
type ExtendedStatefulSetsGetter interface {
	ExtendedStatefulSets(namespace string) ExtendedStatefulSetInterface
}

// ExtendedStatefulSetInterface has methods to work with ExtendedStatefulSet resources.
type ExtendedStatefulSetInterface interface {
	Create(*v1alpha1.ExtendedStatefulSet) (*v1alpha1.ExtendedStatefulSet, error)
	Update(*v1alpha1.ExtendedStatefulSet) (*v1alpha1.ExtendedStatefulSet, error)
	UpdateStatus(*v1alpha1.ExtendedStatefulSet) (*v1alpha1.ExtendedStatefulSet, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.ExtendedStatefulSet, error)
	List(opts v1.ListOptions) (*v1alpha1.ExtendedStatefulSetList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ExtendedStatefulSet, err error)
	ExtendedStatefulSetExpansion
}

// extendedStatefulSets implements ExtendedStatefulSetInterface
type extendedStatefulSets struct {
	client rest.Interface
	ns     string
}

// newExtendedStatefulSets returns a ExtendedStatefulSets
func newExtendedStatefulSets(c *ExtendedstatefulsetV1alpha1Client, namespace string) *extendedStatefulSets {
	return &extendedStatefulSets{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the extendedStatefulSet, and returns the corresponding extendedStatefulSet object, and an error if there is any.
func (c *extendedStatefulSets) Get(name string, options v1.GetOptions) (result *v1alpha1.ExtendedStatefulSet, err error) {
	result = &v1alpha1.ExtendedStatefulSet{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("extendedstatefulsets").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ExtendedStatefulSets that match those selectors.
func (c *extendedStatefulSets) List(opts v1.ListOptions) (result *v1alpha1.ExtendedStatefulSetList, err error) {
	result = &v1alpha1.ExtendedStatefulSetList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("extendedstatefulsets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested extendedStatefulSets.
func (c *extendedStatefulSets) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("extendedstatefulsets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a extendedStatefulSet and creates it.  Returns the server's representation of the extendedStatefulSet, and an error, if there is any.
func (c *extendedStatefulSets) Create(extendedStatefulSet *v1alpha1.ExtendedStatefulSet) (result *v1alpha1.ExtendedStatefulSet, err error) {
	result = &v1alpha1.ExtendedStatefulSet{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("extendedstatefulsets").
		Body(extendedStatefulSet).
		Do().
		Into(result)
	return
}

// Update takes the representation of a extendedStatefulSet and updates it. Returns the server's representation of the extendedStatefulSet, and an error, if there is any.
func (c *extendedStatefulSets) Update(extendedStatefulSet *v1alpha1.ExtendedStatefulSet) (result *v1alpha1.ExtendedStatefulSet, err error) {
	result = &v1alpha1.ExtendedStatefulSet{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("extendedstatefulsets").
		Name(extendedStatefulSet.Name).
		Body(extendedStatefulSet).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *extendedStatefulSets) UpdateStatus(extendedStatefulSet *v1alpha1.ExtendedStatefulSet) (result *v1alpha1.ExtendedStatefulSet, err error) {
	result = &v1alpha1.ExtendedStatefulSet{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("extendedstatefulsets").
		Name(extendedStatefulSet.Name).
		SubResource("status").
		Body(extendedStatefulSet).
		Do().
		Into(result)
	return
}

// Delete takes name of the extendedStatefulSet and deletes it. Returns an error if one occurs.
func (c *extendedStatefulSets) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("extendedstatefulsets").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *extendedStatefulSets) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("extendedstatefulsets").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched extendedStatefulSet.
func (c *extendedStatefulSets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ExtendedStatefulSet, err error) {
	result = &v1alpha1.ExtendedStatefulSet{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("extendedstatefulsets").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
