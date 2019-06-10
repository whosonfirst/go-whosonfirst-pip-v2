package names

import (
	"regexp"
)

var RE_LANGUAGETAG *regexp.Regexp

func init() {

	RE_LANGUAGETAG = regexp.MustCompile(`^((?P<grandfathered>(en-GB-oed|i-ami|i-bnn|i-default|i-enochian|i-hak|i-klingon|i-lux|i-mingo|i-navajo|i-pwn|i-tao|i-tay|i-tsu|sgn-BE-FR|sgn-BE-NL|sgn-CH-DE)|(art-lojban|cel-gaulish|no-bok|no-nyn|zh-guoyu|zh-hakka|zh-min|zh-min-nan|zh-xiang))|((?P<language>([A-Za-z]{2,3}(_(?P<extlang>[A-Za-z]{3}(_[A-Za-z]{3}){0,2}))?)|[A-Za-z]{4}|[A-Za-z]{5,8})(_(?P<script>[A-Za-z]{4}))?(_(?P<region>[A-Za-z]{2}|[0-9]{3}))?(_(?P<variant>[A-Za-z0-9]{5,8}|[0-9][A-Za-z0-9]{3}))*(_(?P<extension>[0-9A-WY-Za-wy-z](-[A-Za-z0-9]{2,8})+))*(_(?P<privateuse>x(_[A-Za-z0-9]{1,})+))?)|(?P<privateuse_gf>x(_[A-Za-z0-9]{1,8})+))$`)

}
