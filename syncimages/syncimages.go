package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client"
	"github.com/docker/distribution/registry/client/auth/challenge"
)

type Repository struct {
	Name            string
	ManifestService distribution.ManifestService
	TagService      distribution.TagService
	BlobStore       distribution.BlobStore
	ctx             context.Context
}

func NewRepository(ctx context.Context, name, tag, base string, tr http.RoundTripper) *Repository {
	ref, err := reference.WithName(name)
	if err != nil {
		log.Fatal(err)
	}

	if len(tag) > 0 {
		ref, err = reference.WithTag(ref, tag)
		if err != nil {
			log.Fatal(err)
		}
	}

	repo, err := client.NewRepository(ref, base, tr)
	if err != nil {
		log.Fatal(err)
	}

	mfs, err := repo.Manifests(ctx)
	if err != nil {
		log.Fatal(err)
	}

	return &Repository{
		Name:            name,
		ManifestService: mfs,
		TagService:      repo.Tags(ctx),
		BlobStore:       repo.Blobs(ctx),
		ctx:             ctx,
	}
}

func (r *Repository) Sync(dst *Repository, tags []string) {
	for _, t := range tags {
		types := distribution.WithManifestMediaTypes([]string{schema2.MediaTypeManifest, schema1.MediaTypeManifest})
		mfs, err := r.ManifestService.Get(r.ctx, "", distribution.WithTag(t), types)
		if err != nil {
			log.Fatal(err)
		}

		for _, ref := range mfs.References() {
			BlobSync(r.ctx, r.BlobStore, dst.BlobStore, ref)
		}
	}
}

func BlobSync(ctx context.Context, src, dst distribution.BlobStore, descriptor distribution.Descriptor) {
	log.Printf("in BlobSync, %s\n", descriptor.Digest)

	desc, err := src.Stat(ctx, descriptor.Digest)
	if err != nil {
		log.Fatal(err)
	}

	reader, err := src.Open(ctx, descriptor.Digest)
	if err != nil {
		log.Fatal(err)
	}

	writer, err := dst.Create(ctx)
	if err != nil {
		log.Fatal(err)
	}

	_, err = io.CopyN(writer, reader, desc.Size)
	if err != nil {
		log.Fatal(err)
	}

	_, err = writer.Commit(ctx, desc)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("out BlobSync, %s\n", descriptor.Digest)
}

const scheme = "bearer"

type Modifier interface {
	Modify(r *http.Request) error
}

type TokenModifier struct {
	rt   http.RoundTripper
	user string
	pass string
}

func NewTokenModifier(user, pass string, rt http.RoundTripper) Modifier {
	return &TokenModifier{
		rt:   rt,
		user: user,
		pass: pass,
	}
}

func (m *TokenModifier) Modify(req *http.Request) error {
	log.Printf("in Modify, %s%s\n", req.Host, req.URL.Path)
	cli := &http.Client{Transport: m.rt}

	realm, service, scopes, err := m.ping(cli, req)
	if err != nil {
		log.Fatalf("ping error, %v\n", err)
	}

	if len(realm) == 0 {
		log.Printf("empty realm skip")
		return nil
	}

	token, err := m.fetchToken(cli, realm, service, scopes)
	if err != nil {
		log.Fatalf("fetch token error %v\n", err)
	}

	req.Header.Add(http.CanonicalHeaderKey("Authorization"), fmt.Sprintf("Bearer %s", token))

	log.Printf("out Modify, %s%s\n", req.Host, req.URL.Path)
	return nil
}

func (m *TokenModifier) ping(cli *http.Client, req *http.Request) (string, string, string, error) {
	newreq, err := http.NewRequest(req.Method, req.URL.String(), nil)
	if err != nil {
		return "", "", "", err
	}

	resp, err := cli.Do(newreq)
	if err != nil {
		return "", "", "", err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		cs := challenge.ResponseChallenges(resp)
		for _, ch := range cs {
			if scheme == ch.Scheme {
				return ch.Parameters["realm"], ch.Parameters["service"], ch.Parameters["scope"], nil
			}
		}
		log.Printf("schemes %v are unsupported\n", cs)
	} else if resp.StatusCode == http.StatusOK {
		return "", "", "", nil
	}

	return "", "", "", fmt.Errorf("x scheme %s are unsupported\n", scheme)
}

func (m *TokenModifier) fetchToken(cli *http.Client, realm, service, scopes string) (string, error) {
	u, err := url.Parse(realm)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Add("service", service)
	if len(scopes) > 0 {
		q.Add("scope", scopes)
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}

	if len(m.user) > 0 {
		req.SetBasicAuth(m.user, m.pass)
	}

	resp, err := cli.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token read fail, status %s\n", resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	token := &Token{}
	if err := json.Unmarshal(data, token); err != nil {
		return "", err
	}

	return token.Token, nil
}

type Token struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
	IssuedAt  string `json:"issued_at"`
}

type Transport struct {
	rt        http.RoundTripper
	modifiers []Modifier
}

func NewTransport(rt http.RoundTripper, ms ...Modifier) *Transport {
	return &Transport{rt: rt, modifiers: ms}
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	log.Printf("in RoundTrip, %s%s\n", req.Host, req.URL.Path)

	for _, m := range t.modifiers {
		if m != nil {
			if err := m.Modify(req); err != nil {
				log.Fatalf("modify error, %v\n", err)
			}
		}
	}

	resp, err := t.rt.RoundTrip(req)
	if err != nil {
		log.Fatalf("round trip error, %v\n", err)
	}

	log.Printf("out RoundTrip, %s%s\n", req.Host, req.URL.Path)
	return resp, err
}

func main() {
	src := "https://index.docker.io"
	dst := "http://192.168.1.33:5000"

	srctr := &http.Transport{
		DisableKeepAlives:   true,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	modifer := NewTokenModifier("", "", srctr)
	srcTransport := NewTransport(srctr, modifer)
	srcRepo := NewRepository(context.TODO(), "library/alpine", "3.10.2", src, srcTransport)

	dsttr := &http.Transport{
		DisableKeepAlives:   true,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 100 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	dstTransport := NewTransport(dsttr, nil)
	dstRepo := NewRepository(context.TODO(), "library/alpine", "3.10.2", dst, dstTransport)

	srcRepo.Sync(dstRepo, []string{"3.10.2"})
}
