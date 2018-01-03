package indexer

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type SyntaxCheckRequest struct {
	Source string `json:"source"`
}

type SyntaxCheckResult struct {
	Valid       bool   `json:"valid"`
	DiagramType string `json:"diagramType"`
	Description string `json:"description"`
}

func (r *SyntaxCheckResult) HasValidDiagram() bool {
	re := regexp.MustCompile(`^\(([0-9]+) .+\)$`)
	matched := re.FindStringSubmatch(r.Description)
	if len(matched) != 2 {
		// regard unexpected description as valid
		return true
	}
	if matched[1] == "0" {
		return false
	}
	return true
}

type SyntaxChecker struct {
	BaseUrl string
	ctx     context.Context
}

func NewSyntaxChecker(ctx context.Context, baseUrl string) *SyntaxChecker {
	return &SyntaxChecker{
		BaseUrl: baseUrl,
		ctx:     ctx,
	}
}

func (s *SyntaxChecker) CheckSyntax(source string) (*SyntaxCheckResult, error) {
	checkReq := &SyntaxCheckRequest{source}
	reqBody, err := json.Marshal(checkReq)
	if err != nil {
		return nil, err
	}

	log.Infof(s.ctx, "request body: %s", string(reqBody))

	req, err := http.NewRequest("POST", s.BaseUrl+"/check_syntax", strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	client := urlfetch.Client(s.ctx)
	resp, err := client.Do(req)
	if err != nil {
		log.Criticalf(s.ctx, "failed to request to syntax checker: err=%s", err)
		return nil, err
	}
	defer resp.Body.Close()

	var result SyntaxCheckResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
