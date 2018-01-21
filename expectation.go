package main

import (
	"regexp"
)

type (
	expectation struct {
		method  *regexp.Regexp
		path    *regexp.Regexp
		headers map[string]*regexp.Regexp
		body    *regexp.Regexp
	}

	match struct {
		methodGroups []string
		pathGroups   []string
		headerGroups map[string][]string
		bodyGroups   []string
	}

	matcher func(*request, *match) *match
)

func (e *expectation) Matches(r *request) *match {
	match := &match{}
	for _, m := range []matcher{e.matchMethod, e.matchPath, e.matchHeaders, e.matchBody} {
		match = m(r, match)

		if match == nil {
			break
		}
	}

	return match
}

func (e *expectation) matchMethod(r *request, m *match) *match {
	if match, groups := matchRegex(e.method, r.Method); match {
		m.methodGroups = groups
		return m
	}

	return nil
}

func (e *expectation) matchPath(r *request, m *match) *match {
	if match, groups := matchRegex(e.path, r.Path); match {
		m.pathGroups = groups
		return m
	}

	return nil
}

func (e *expectation) matchHeaders(r *request, m *match) *match {
	headerGroups := map[string][]string{}

	for k, re := range e.headers {
		if match, groups := matchRegex(re, getFirst(r.Headers, k)); match {
			headerGroups[k] = groups
			continue
		}

		return nil
	}

	m.headerGroups = headerGroups
	return m
}

func (e *expectation) matchBody(r *request, m *match) *match {
	if match, groups := matchRegex(e.body, r.Body); match {
		m.bodyGroups = groups
		return m
	}

	return nil
}

//
// Helpers

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
