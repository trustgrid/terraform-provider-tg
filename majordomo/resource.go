package majordomo

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

// Resource manages CRUD operations and marshaling between `hcl` and `tg` types.
type Resource[T any, H hcl.HCL[T]] struct {
	createURL     func(H) string
	onCreateReply func(*schema.ResourceData, []byte) (string, error)
	onUpdateReply func(*schema.ResourceData, []byte) (string, error)
	getFromNode   func(tg.Node) (T, bool, error)
	updateURL     func(H) string
	deleteURL     func(H) string
	getURL        func(H) string
	indexURL      func() string
	id            func(H) string
	remoteID      func(T) string
}

type ResourceArgs[T any, H hcl.HCL[T]] struct {
	CreateURL     func(H) string                                     // CreateURL should return the URL for POST-ing the resource. If not set, calls to `Create` will attempt to call `Update`.
	OnCreateReply func(*schema.ResourceData, []byte) (string, error) // OnCreateReply is called after a successful POST request. The ID returned will be set as the resource ID.
	OnUpdateReply func(*schema.ResourceData, []byte) (string, error) // OnUpdateReply is called after a successful PUT request. The ID returned will be set as the resource ID.
	GetFromNode   func(tg.Node) (T, bool, error)                     // GetFromNode should return the `tg` resource from the `tg.Node` resource.
	DeleteURL     func(H) string                                     // DeleteURL should return the URL for DELETE-ing the resource.
	GetURL        func(H) string                                     // GetURL should return the URL for GET-ing the resource, provided the API supports individual lookups.
	UpdateURL     func(H) string                                     // UpdateURL should return the URL for PUT-ing the resource.
	IndexURL      func() string                                      // IndexURL should return the URL for GET-ing a list of resources. If this and RemoteID are provided and GetURL is not, `Read` will attempt to call `Index` and search for the resource.
	RemoteID      func(T) string                                     // RemoteID should return the ID of `tg` resource from the remote API.
	ID            func(H) string                                     // ID should return the ID of the `hcl` resource.
}

// NewResource returns a new `Resource`.
func NewResource[T any, H hcl.HCL[T]](args ResourceArgs[T, H]) *Resource[T, H] {
	return &Resource[T, H]{
		createURL:     args.CreateURL,
		updateURL:     args.UpdateURL,
		getFromNode:   args.GetFromNode,
		onCreateReply: args.OnCreateReply,
		onUpdateReply: args.OnUpdateReply,
		deleteURL:     args.DeleteURL,
		getURL:        args.GetURL,
		indexURL:      args.IndexURL,
		remoteID:      args.RemoteID,
		id:            args.ID,
	}
}

// Create calls the `CreateURL` function to get the URL for POST-ing the resource. If `CreateURL` is not set, calls `Update`.
// HCL information is decoded from the `ResourceData` and marshaled to `tg` type before POST-ing.
func (r *Resource[T, H]) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if r.createURL == nil {
		return r.Update(ctx, d, meta)
	}

	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[H](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tg := tf.ToTG()

	reply, err := tgc.Post(ctx, r.createURL(tf), tg)
	if err != nil {
		return diag.FromErr(err)
	}

	if r.onCreateReply != nil {
		id, err := r.onCreateReply(d, reply)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(id)
	} else {
		d.SetId(r.id(tf))
	}

	return nil
}

// Noop is a no-op function that returns no diagnostics.
func (r *Resource[T, H]) Noop(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return nil
}

// Update calls the `UpdateURL` function to get the URL for PUT-ing the resource.
// HCL information is decoded from the `ResourceData` and marshaled to `tg` type before PUT-ing.
func (r *Resource[T, H]) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[H](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tg := tf.ToTG()

	reply, err := tgc.Put(ctx, r.updateURL(tf), tg)
	if err != nil {
		return diag.FromErr(err)
	}

	if r.onUpdateReply != nil {
		id, err := r.onUpdateReply(d, reply)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(id)
	} else if d.Id() == "" {
		d.SetId(r.id(tf))
	}

	return nil
}

// Delete calls the `DeleteURL` function to get the URL for DELETE-ing the resource.
// HCL information is decoded from the `ResourceData` and marshaled to `tg` type before DELETE-ing.
func (r *Resource[T, H]) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[H](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tg := tf.ToTG()

	if err := tgc.Delete(ctx, r.deleteURL(tf), tg); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// Read calls the `GetURL` function to get the URL for GET-ing the resource.
// If `GetURL` is not set, calls `Index` and searches for the resource.
// HCL information is decoded from the `ResourceData` and marshaled to `tg` type before GET-ing.
// After retrieving the API record, the HCL resource will be updated with the new information.
func (r *Resource[T, H]) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var t T
	var err error
	ok := true

	tf, err := hcl.DecodeResourceData[H](d)
	if err != nil {
		return diag.FromErr(err)
	}

	switch {
	case r.getURL != nil:
		// TODO probably should have getURL just error if it needs something that isn't set
		if d.Id() == "" {
			return nil
		}
		t, ok, err = r.read(ctx, tf, meta)
	case r.indexURL != nil:
		t, ok, err = r.index(ctx, tf, meta)
	}

	switch {
	case err != nil:
		return diag.FromErr(err)
	case !ok:
		d.SetId("")
		return nil
	}

	updated := tf.UpdateFromTG(t)

	if err := hcl.EncodeResourceData(updated, d); err != nil {
		return diag.FromErr(err)
	}

	a, ok := updated.(H)
	if !ok {
		return diag.FromErr(errors.New("failed to cast"))
	}
	d.SetId(r.id(a))

	return nil
}

func (r *Resource[T, H]) index(ctx context.Context, tf H, meta any) (T, bool, error) {
	tgc := tg.GetClient(meta)

	upstream := make([]T, 0)

	var out T

	if err := tgc.Get(ctx, r.indexURL(), &upstream); err != nil {
		return out, false, err
	}

	for _, i := range upstream {
		if r.remoteID(i) == r.id(tf) {
			return i, true, nil
		}
	}

	return out, false, nil
}

func (r *Resource[T, H]) read(ctx context.Context, tf H, meta any) (T, bool, error) {
	tgc := tg.GetClient(meta)

	var out T
	if r.getFromNode != nil {
		var n tg.Node
		err := tgc.Get(ctx, r.getURL(tf), &n)
		var nferr *tg.NotFoundError
		switch {
		case errors.As(err, &nferr):
			return out, false, nil
		case err != nil:
			return out, false, err
		}

		return r.getFromNode(n)
	}

	err := tgc.Get(ctx, r.getURL(tf), &out)
	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		return out, false, nil
	case err != nil:
		return out, false, err
	}

	return out, true, nil
}
