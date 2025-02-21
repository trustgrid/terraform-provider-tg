package majordomo

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type Resource[T any, H hcl.HCL[T]] struct {
	createURL     func(H) string
	onCreateReply func(*schema.ResourceData, []byte) (string, error)
	updateURL     func(H) string
	deleteURL     func(H) string
	getURL        func(H) string
	indexURL      func() string
	id            func(H) string
	remoteID      func(T) string
}

type ResourceArgs[T any, H hcl.HCL[T]] struct {
	CreateURL     func(H) string
	OnCreateReply func(*schema.ResourceData, []byte) (string, error)
	DeleteURL     func(H) string
	GetURL        func(H) string
	UpdateURL     func(H) string
	IndexURL      func() string
	RemoteID      func(T) string
	ID            func(H) string
}

func NewResource[T any, H hcl.HCL[T]](args ResourceArgs[T, H]) *Resource[T, H] {
	return &Resource[T, H]{
		createURL:     args.CreateURL,
		updateURL:     args.UpdateURL,
		onCreateReply: args.OnCreateReply,
		deleteURL:     args.DeleteURL,
		getURL:        args.GetURL,
		indexURL:      args.IndexURL,
		remoteID:      args.RemoteID,
		id:            args.ID,
	}
}

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

func (r *Resource[T, H]) Noop(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return nil
}

func (r *Resource[T, H]) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[H](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tg := tf.ToTG()

	if err := tgc.Put(ctx, r.updateURL(tf), tg); err != nil {
		return diag.FromErr(err)
	}

	if d.Id() == "" {
		d.SetId(r.id(tf))
	}

	return nil
}

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
