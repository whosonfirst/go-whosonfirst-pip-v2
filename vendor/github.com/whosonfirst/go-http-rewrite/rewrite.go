package rewrite

import (
	"fmt"
	_ "log"
	"net/http"
	"regexp"
)

type RewriteRule interface {
	Match(string) bool
	Transform(string) string
	Last() bool
}

type RewriteRuleOptions map[string]interface{}

func DefaultRewriteRuleOptions() RewriteRuleOptions {

	opts := RewriteRuleOptions{
		"Last": false,
	}

	return opts
}

type RegexpRewriteRule struct {
	RewriteRule
	regexp  *regexp.Regexp
	replace string
	is_last bool
}

func (rule *RegexpRewriteRule) Match(path string) bool {
	return rule.regexp.MatchString(path)
}

func (rule *RegexpRewriteRule) Transform(path string) string {
	return rule.regexp.ReplaceAllString(path, rule.replace)
}

func (rule *RegexpRewriteRule) Last() bool {
	return rule.is_last
}

func RemovePrefixRewriteRule(path string, opts RewriteRuleOptions) RewriteRule {

	pat := fmt.Sprintf("^%s(.*)", path)
	re := regexp.MustCompile(pat)

	is_last := false

	_, ok := opts["Last"]

	if ok {
		is_last = opts["Last"].(bool)
	}

	rule := RegexpRewriteRule{
		regexp:  re,
		replace: "$1",
		is_last: is_last,
	}

	return &rule
}

func RewriteHandler(rules []RewriteRule, next http.Handler) (http.Handler, error) {

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		path := req.URL.Path

		for _, rw := range rules {

			if !rw.Match(path) {
				continue
			}

			path = rw.Transform(path)

			if rw.Last() {
				break
			}
		}

		req.URL.Path = path

		next.ServeHTTP(rsp, req)
		return
	}

	h := http.HandlerFunc(fn)
	return h, nil
}
