package indexer

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type Renderer struct {
	BaseUrl string
	ctx     context.Context
}

func NewRenderer(ctx context.Context, baseUrl string) *Renderer {
	return &Renderer{
		BaseUrl: baseUrl,
		ctx:     ctx,
	}
}

func (r *Renderer) RenderSvg(source string) (string, error) {
	umlId, err := r.getUmlId(source)
	if err != nil {
		return "", err
	}

	svgBytes, err := r.doRequest("/svg/" + umlId)
	if err != nil {
		return "", err
	}
	return string(svgBytes), err
}

func (r *Renderer) RenderPng(source string) ([]byte, error) {
	umlId, err := r.getUmlId(source)
	if err != nil {
		return nil, err
	}

	return r.doRequest("/png/" + umlId)
}

func (r *Renderer) RenderAscii(source string) (string, error) {
	umlId, err := r.getUmlId(source)
	if err != nil {
		return "", err
	}

	asciiBytes, err := r.doRequest("/txt/" + umlId)
	if err != nil {
		return "", err
	}
	return string(asciiBytes), err
}

func (r *Renderer) getUmlId(source string) (string, error) {
	values := url.Values{}
	values.Add("text", source)
	req, _ := http.NewRequest("POST", r.BaseUrl+"/form", strings.NewReader(values.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := urlfetch.Client(r.ctx)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		// no follow redirect
		return http.ErrUseLastResponse
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Criticalf(r.ctx, "Failed to request to renderer /form: %s", err)
		return "", err
	}
	defer resp.Body.Close()

	locationUrl, err := resp.Location()
	if err != nil {
		log.Criticalf(r.ctx, "Failed to get location header: %s", err)
		return "", err
	}
	umlId := strings.TrimPrefix(locationUrl.Path, "/uml/")

	return umlId, nil
}

func (r *Renderer) doRequest(path string) ([]byte, error) {
	req, _ := http.NewRequest("GET", r.BaseUrl+path, nil)

	client := urlfetch.Client(r.ctx)
	resp, err := client.Do(req)
	if err != nil {
		log.Criticalf(r.ctx, "Failed to request to %s: err=%s", path, err)
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
