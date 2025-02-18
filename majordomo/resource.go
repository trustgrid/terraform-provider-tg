package majordomo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type HCL interface {
	UpdateFromTG(any)
}

type Resource[H *HCL, T any] struct {
	createURL func(H) string
	updateURL func(H) string
	deleteURL func(H) string
	getURL    func(H) string
	indexURL  func() string
	id        func(H) string
	remoteID  func(T) string
}

type ResourceArgs[H HCL] struct {
	CreateURL func(H) string
	DeleteURL func(H) string
	GetURL    func(H) string
	UpdateURL func(H) string
	ID        func(H) string
}

func NewResource[H *HCL, T any](args ResourceArgs[H]) *Resource[H, T] {
	return &Resource[H, T]{
		createURL: args.CreateURL,
		updateURL: args.UpdateURL,
		deleteURL: args.DeleteURL,
		id:        args.ID,
	}
}

func (r *Resource[H, T]) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[H](d)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, err := tgc.Post(ctx, r.createURL(tf), &tf); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.id(tf)) //tf.FQDN)

	return nil
}

func (r *Resource[H, T]) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[H](d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Put(ctx, r.updateURL(tf), &tf); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (r *Resource[H, T]) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[H](d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Delete(ctx, r.deleteURL(tf), &tf); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (r *Resource[H, T]) index(ctx context.Context, tf H, meta any) (*T, error) {
	tgc := tg.GetClient(meta)

	upstream := make([]T, 0)

	if err := tgc.Get(ctx, r.indexURL(), &upstream); err != nil {
		return new(T), err
	}

	for _, i := range upstream {
		if r.remoteID(i) == r.id(tf) {
			return &i, nil
		}
	}

	return nil, nil
}

func (r *Resource[H, T]) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tf, err := hcl.DecodeResourceData[H](d)
	if err != nil {
		return diag.FromErr(err)
	}

	/*switch {
	case r.getURL != nil:
		return r.read(ctx, d, meta)
	case r.indexURL != nil:*/
	t, err := r.index(ctx, tf, meta)
	switch {
	case err != nil:
		return diag.FromErr(err)
	case t == nil:
		d.SetId("")
		return nil
	}

	tf.UpdateFromTG(*t)

	hcl.EncodeResourceData(tf, d)

	return nil
	//}

	/*
		tgc := tg.GetClient(meta)

		certs := make([]tg.Cert, 0)

		if err := tgc.Get(ctx, "/v2/certificates", &certs); err != nil {
			return diag.FromErr(err)
		}

		cert := tg.Cert{}

		for _, c := range certs {
			if c.FQDN == d.Get("fqdn").(string) {
				cert = c
				break
			}
		}

		if cert.FQDN == "" {
			d.SetId("")
		}

		return nil
	*/
}
