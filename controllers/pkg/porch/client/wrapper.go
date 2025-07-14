package client

import (
	"context"
	"reflect"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WrapClient(c client.Client) client.Client {
	return &clientWrapper{c}
}

type clientWrapper struct {
	internalClient client.Client
}

var _ client.Client = &clientWrapper{}

func getKind(obj runtime.Object) string {
	if kind := obj.GetObjectKind().GroupVersionKind().Kind; kind != "" {
		return kind
	}

	return reflect.TypeOf(obj).Elem().Name()
}

func (c *clientWrapper) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	klog.Infof("Client GET %s::%s", getKind(obj), key)
	return c.internalClient.Get(ctx, key, obj, opts...)
}

func (c *clientWrapper) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	klog.Infof("Client LIST %s", getKind(list.(runtime.Object)))
	return c.internalClient.List(ctx, list, opts...)
}

func (c *clientWrapper) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	name := obj.GetName()
	if name == "" {
		name = "<unnamed>"
	}
	klog.Infof("Client CREATE %s::%s", getKind(obj), name)
	return c.internalClient.Create(ctx, obj, opts...)
}

func (c *clientWrapper) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	klog.Infof("Client DELETE %s::%s", getKind(obj), obj.GetName())
	return c.internalClient.Delete(ctx, obj, opts...)
}

func (c *clientWrapper) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	klog.Infof("Client UPDATE %s::%s", getKind(obj), obj.GetName())
	return c.internalClient.Update(ctx, obj, opts...)
}

func (c *clientWrapper) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	klog.Infof("Client PATCH %s::%s", getKind(obj), obj.GetName())
	return c.internalClient.Patch(ctx, obj, patch, opts...)
}

func (c *clientWrapper) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	klog.Infof("Client DELETE ALL %s", getKind(obj))
	return c.internalClient.DeleteAllOf(ctx, obj, opts...)
}

func (c *clientWrapper) Status() client.SubResourceWriter {
	return c.internalClient.Status()
}

func (c *clientWrapper) SubResource(subResource string) client.SubResourceClient {
	return c.internalClient.SubResource(subResource)
}

func (c *clientWrapper) Scheme() *runtime.Scheme {
	return c.internalClient.Scheme()
}

func (c *clientWrapper) RESTMapper() meta.RESTMapper {
	return c.internalClient.RESTMapper()
}

func (c *clientWrapper) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return c.internalClient.GroupVersionKindFor(obj)
}

func (c *clientWrapper) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return c.internalClient.IsObjectNamespaced(obj)
}
