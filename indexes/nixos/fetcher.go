package nixos

import (
	"cmp"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/3timeslazy/nix-search-tv/indexer"
	"github.com/3timeslazy/nix-search-tv/indexes/readutil"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Fetcher struct{}

const prefix = "nixos/unstable/"

func (f *Fetcher) GetLatestRelease(ctx context.Context, md indexer.IndexMetadata) (string, error) {
	s3client := s3.NewFromConfig(aws.Config{
		Region: "eu-west-1",
	})

	// The `startAfter` is a marker for S3 to start iterating from. Just use the latest
	// at the moment of writing nixpkgs release to never iterate from the beginning
	startAfter := cmp.Or(md.CurrRelease, "nixos/unstable/nixos-25.05beta751650.64e75cd44acf")
	var latest types.Object
	input := &s3.ListObjectsV2Input{
		Bucket:     aws.String("nix-releases"),
		Prefix:     aws.String(prefix),
		Delimiter:  aws.String("/"),
		StartAfter: aws.String(startAfter),
	}
	p := s3.NewListObjectsV2Paginator(s3client, input)
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return "", fmt.Errorf("get next page: %w", err)
		}
		for _, obj := range page.Contents {
			latest = obj
		}
	}

	if latest.Key == nil {
		return md.CurrRelease, nil
	}
	return *latest.Key, nil
}

func (f *Fetcher) DownloadRelease(ctx context.Context, release string) (io.ReadCloser, error) {
	release = strings.TrimPrefix(release, prefix)
	url, _ := url.JoinPath("https://releases.nixos.org/nixos/unstable", release, "options.json.br")

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch packages: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected http 200, but %d", resp.StatusCode)
	}

	return readutil.PackagesWrapper(readutil.NewBrotli(resp.Body)), nil
}
