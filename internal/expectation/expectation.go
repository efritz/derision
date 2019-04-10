package expectation

import (
	"regexp"

	"github.com/efritz/derision/internal/request"
)

type (
	Expectation interface {
		Matches(r *request.Request) *Match
	}

	Match struct {
		MethodGroups []string
		PathGroups   []string
		HeaderGroups map[string][]string
		BodyGroups   []string
	}

	expectation struct {
		method  *regexp.Regexp
		path    *regexp.Regexp
		headers map[string]*regexp.Regexp
		body    *regexp.Regexp
	}

	matcher func(*request.Request, *Match) *Match
)

func (e *expectation) Matches(r *request.Request) *Match {
	match := &Match{}
	for _, m := range []matcher{e.matchMethod, e.matchPath, e.matchHeaders, e.matchBody} {
		match = m(r, match)

		if match == nil {
			break
		}
	}

	return match
}

func (e *expectation) matchMethod(r *request.Request, m *Match) *Match {
	if match, groups := matchRegex(e.method, r.Method); match {
		m.MethodGroups = groups
		return m
	}

	return nil
}

func (e *expectation) matchPath(r *request.Request, m *Match) *Match {
	if match, groups := matchRegex(e.path, r.Path); match {
		m.PathGroups = groups
		return m
	}

	return nil
}

func (e *expectation) matchHeaders(r *request.Request, m *Match) *Match {
	headerGroups := map[string][]string{}

	for k, re := range e.headers {
		if match, groups := matchRegex(re, getFirst(r.Headers, k)); match {
			headerGroups[k] = groups
			continue
		}

		return nil
	}

	m.HeaderGroups = headerGroups
	return m
}

func (e *expectation) matchBody(r *request.Request, m *Match) *Match {
	if match, groups := matchRegex(e.body, r.Body); match {
		m.BodyGroups = groups
		return m
	}

	return nil
}

func matchRegex(re *regexp.Regexp, val string) (bool, []string) {
	if re == nil {
		return true, nil
	}

	if !re.MatchString(val) {
		return false, nil
	}

	return true, re.FindStringSubmatch(val)
}

func getFirst(headers map[string][]string, k string) string {
	if vals, ok := headers[k]; ok && len(vals) > 0 {
		return vals[0]
	}

	return ""
}
