package indexer

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type Renderer struct {
	BaseUrl string
	UmlId   string
	ctx     context.Context
}

func NewRenderer(ctx context.Context, scheme, host string, port int, source string) (*Renderer, error) {
	baseUrl := fmt.Sprintf("%s://%s:%d", scheme, host, port)

	values := url.Values{}
	values.Add("text", source)
	req, _ := http.NewRequest("POST", baseUrl+"/form", strings.NewReader(values.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := urlfetch.Client(ctx)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		// no follow redirect
		return http.ErrUseLastResponse
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Criticalf(ctx, "Failed to request to renderer /form: %s", err)
		return nil, err
	}
	defer resp.Body.Close()

	locationUrl, err := resp.Location()
	if err != nil {
		log.Criticalf(ctx, "Failed to get location header: %s", err)
		return nil, err
	}
	umlId := strings.TrimPrefix(locationUrl.Path, "/uml/")

	return &Renderer{
		BaseUrl: baseUrl,
		UmlId:   umlId,
		ctx:     ctx,
	}, err
}

func (r *Renderer) RenderSvg() (string, error) {
	req, _ := http.NewRequest("GET", r.BaseUrl+"/svg/"+r.UmlId, nil)

	client := urlfetch.Client(r.ctx)
	resp, err := client.Do(req)
	if err != nil {
		log.Criticalf(r.ctx, "Failed to request to renderer /svg: err=%s", err)
		return "", err
	}
	defer resp.Body.Close()

	svgBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(svgBytes), nil
}

func (r *Renderer) RenderPng(source string) {
}

func (r *Renderer) RenderAscii(source string) {
}

func (r *Renderer) RenderCheck(source string) {
}
