package backend

import (
	"context"
	"net/http"

	"github.com/favclip/ucon"
	"github.com/favclip/ucon/swagger"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

func setupSpannerAccountAPI(swPlugin *swagger.Plugin) {
	api := &SpannerAccountAPI{}

	tag := swPlugin.AddTag(&swagger.Tag{Name: "SpannerAccount", Description: "SpannerAccount API List"})
	var hInfo *swagger.HandlerInfo

	hInfo = swagger.NewHandlerInfo(api.Post)
	ucon.Handle(http.MethodPost, "/api/1/account", hInfo)
	hInfo.Description, hInfo.Tags = "post to spanner-account", []string{tag.Name}
}

// SpannerAccountAPI is Organization Admin API Functions
type SpannerAccountAPI struct{}

// SpannerAccountAPIPostRequest is Organization Admin Post API Request
type SpannerAccountAPIPostRequest struct {
	GCPUGSlackID    string   `json:"gcpugSlackId"`
	ServiceAccounts []string `json:"serviceAccounts"`
}

// SpannerAccountAPIPostResponse is Organization Admin Post API Response
type SpannerAccountAPIPostResponse struct {
	*SpannerAccount
}

// Post is SpannerAccountを登録する
func (api *SpannerAccountAPI) Post(ctx context.Context, form *SpannerAccountAPIPostRequest) (*SpannerAccountAPIPostResponse, error) {
	u := user.Current(ctx)
	if u == nil {
		return nil, &HTTPError{Code: http.StatusForbidden, Message: "Login Required"}
	}

	if form.GCPUGSlackID == "" {
		return nil, &HTTPError{Code: http.StatusBadRequest, Message: "GCPUGSlackId Required"}
	}

	if err := AddSpannerIAM(ctx, u.Email, form.ServiceAccounts); err != nil {
		log.Errorf(ctx, "failed Add IAM : %s, error = %+v", u.Email, err)
		return nil, &HTTPError{Code: http.StatusInternalServerError, Message: "error"}
	}

	store, err := NewSpannerAccountStore(ctx)
	if err != nil {
		log.Errorf(ctx, "failed NewSpannerAccountStore: %+v", err)
		return nil, &HTTPError{Code: http.StatusInternalServerError, Message: "error"}
	}

	res, err := store.Upsert(ctx, store.NameKey(ctx, u.Email), &SpannerAccount{
		GCPUGSlackID:    form.GCPUGSlackID,
		ServiceAccounts: form.ServiceAccounts,
	})
	if err != nil {
		log.Errorf(ctx, "failed Put to Datastore : %s, error = %+v", u.Email, err)
		return nil, &HTTPError{Code: http.StatusInternalServerError, Message: "error"}
	}

	return &SpannerAccountAPIPostResponse{res}, nil
}
