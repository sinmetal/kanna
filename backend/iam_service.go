package backend

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/appengine/log"
)

// AddSpannerIAM is SpannerのIAMをAccountに追加する
func AddSpannerIAM(ctx context.Context, googleAccount string, serviceAccounts []string) error {
	const resource = "gcpug-public-spanner"

	client, err := google.DefaultClient(ctx, cloudresourcemanager.CloudPlatformScope)
	if err != nil {
		return errors.WithStack(err)
	}

	s, err := cloudresourcemanager.New(client)
	if err != nil {
		return errors.WithStack(err)
	}

	p, err := s.Projects.GetIamPolicy(resource, &cloudresourcemanager.GetIamPolicyRequest{}).Do()
	if err != nil {
		return errors.WithStack(err)
	}

	bs := []*cloudresourcemanager.Binding{}
	for _, b := range p.Bindings {
		log.Infof(ctx, "%+v", b)
		if b.Role == "roles/spanner.databaseAdmin" || b.Role == "roles/spanner.viewer" {
			b.Members = append(b.Members, fmt.Sprintf("user:%s", googleAccount))
			for _, sa := range serviceAccounts {
				b.Members = append(b.Members, fmt.Sprintf("serviceAccount:%s", sa))
			}
		}
		bs = append(bs, b)
	}

	_, err = s.Projects.SetIamPolicy(resource, &cloudresourcemanager.SetIamPolicyRequest{
		Policy: &cloudresourcemanager.Policy{
			Bindings: bs,
		},
	}).Do()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
