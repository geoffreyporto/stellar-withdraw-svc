// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

package getters

import (
	"github.com/tokend/stellar-withdraw-svc/internal/horizon/client"
	"github.com/tokend/stellar-withdraw-svc/internal/horizon/page"
	"github.com/tokend/stellar-withdraw-svc/internal/horizon/query"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	regources "gitlab.com/tokend/regources/generated"
)

type CreateWithdrawRequestPager interface {
	Next() (*regources.ReviewableRequestListResponse, error)
	Prev() (*regources.ReviewableRequestListResponse, error)
	Self() (*regources.ReviewableRequestListResponse, error)
	First() (*regources.ReviewableRequestListResponse, error)
}

type CreateWithdrawRequestGetter interface {
	SetFilters(filters query.CreateWithdrawRequestFilters)
	SetIncludes(includes query.CreateWithdrawRequestIncludes)
	SetPageParams(pageParams page.Params)
	SetParams(params query.CreateWithdrawRequestParams)

	Filter() query.CreateWithdrawRequestFilters
	Include() query.CreateWithdrawRequestIncludes
	Page() page.Params

	ByID(ID string) (*regources.ReviewableRequestResponse, error)
	List() (*regources.ReviewableRequestListResponse, error)
}

type CreateWithdrawRequestHandler interface {
	CreateWithdrawRequestGetter
	CreateWithdrawRequestPager
}

type defaultCreateWithdrawRequestHandler struct {
	base   Getter
	params query.CreateWithdrawRequestParams

	currentPageLinks *regources.Links
}

func NewDefaultCreateWithdrawRequestHandler(c *client.Client) *defaultCreateWithdrawRequestHandler {
	return &defaultCreateWithdrawRequestHandler{
		base: New(c),
	}
}

func (g *defaultCreateWithdrawRequestHandler) SetFilters(filters query.CreateWithdrawRequestFilters) {
	g.params.Filters = filters
}

func (g *defaultCreateWithdrawRequestHandler) SetIncludes(includes query.CreateWithdrawRequestIncludes) {
	g.params.Includes = includes
}

func (g *defaultCreateWithdrawRequestHandler) SetPageParams(pageParams page.Params) {
	g.params.PageParams = pageParams
}

func (g *defaultCreateWithdrawRequestHandler) SetParams(params query.CreateWithdrawRequestParams) {
	g.params = params
}

func (g *defaultCreateWithdrawRequestHandler) Params() query.CreateWithdrawRequestParams {
	return g.params
}

func (g *defaultCreateWithdrawRequestHandler) Filter() query.CreateWithdrawRequestFilters {
	return g.params.Filters
}

func (g *defaultCreateWithdrawRequestHandler) Include() query.CreateWithdrawRequestIncludes {
	return g.params.Includes
}

func (g *defaultCreateWithdrawRequestHandler) Page() page.Params {
	return g.params.PageParams
}

func (g *defaultCreateWithdrawRequestHandler) ByID(ID string) (*regources.ReviewableRequestResponse, error) {
	result := &regources.ReviewableRequestResponse{}
	err := g.base.GetPage(query.CreateWithdrawRequestByID(ID), g.params.Includes, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get record by id", logan.F{
			"id": ID,
		})
	}
	return result, nil
}

func (g *defaultCreateWithdrawRequestHandler) List() (*regources.ReviewableRequestListResponse, error) {
	result := &regources.ReviewableRequestListResponse{}
	err := g.base.GetPage(query.CreateWithdrawRequestList(), g.params, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get records list", logan.F{
			"query_params": g.params,
		})
	}
	g.currentPageLinks = result.Links
	return result, nil
}

func (g *defaultCreateWithdrawRequestHandler) Next() (*regources.ReviewableRequestListResponse, error) {
	if g.currentPageLinks == nil {
		return nil, errors.New("Empty links")
	}
	if g.currentPageLinks.Next == "" {
		return nil, errors.From(errors.New("No link to page"), logan.F{
			"page": "next",
		})
	}
	result := &regources.ReviewableRequestListResponse{}
	err := g.base.PageFromLink(g.currentPageLinks.Next, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get next page", logan.F{
			"link": g.currentPageLinks.Next,
		})
	}
	g.currentPageLinks = result.Links

	return result, nil
}

func (g *defaultCreateWithdrawRequestHandler) Prev() (*regources.ReviewableRequestListResponse, error) {
	if g.currentPageLinks == nil {
		return nil, errors.New("Empty links")
	}
	if g.currentPageLinks.Prev == "" {
		return nil, errors.From(errors.New("No link to page"), logan.F{
			"page": "prev",
		})
	}

	result := &regources.ReviewableRequestListResponse{}
	err := g.base.PageFromLink(g.currentPageLinks.Prev, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get next page", logan.F{
			"link": g.currentPageLinks.Prev,
		})
	}
	g.currentPageLinks = result.Links

	return result, nil
}

func (g *defaultCreateWithdrawRequestHandler) Self() (*regources.ReviewableRequestListResponse, error) {
	if g.currentPageLinks == nil {
		return nil, errors.New("Empty links")
	}
	if g.currentPageLinks.Self == "" {
		return nil, errors.From(errors.New("No link to page"), logan.F{
			"page": "self",
		})
	}
	result := &regources.ReviewableRequestListResponse{}
	err := g.base.PageFromLink(g.currentPageLinks.Self, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get same page", logan.F{
			"link": g.currentPageLinks.Self,
		})
	}
	g.currentPageLinks = result.Links

	return result, nil
}

func (g *defaultCreateWithdrawRequestHandler) First() (*regources.ReviewableRequestListResponse, error) {
	if g.currentPageLinks == nil {
		return nil, errors.New("Empty links")
	}
	if g.currentPageLinks.First == "" {
		return nil, errors.From(errors.New("No link to page"), logan.F{
			"page": "first",
		})
	}
	result := &regources.ReviewableRequestListResponse{}
	err := g.base.PageFromLink(g.currentPageLinks.First, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get first page", logan.F{
			"link": g.currentPageLinks.First,
		})
	}
	g.currentPageLinks = result.Links

	return result, nil
}
