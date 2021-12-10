package setecs

import (
	"regexp"
)

const (
	// IP4
	ip4Str = `((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)`

	// IP6，Refer to the following web pages.
	// http://blog.csdn.net/jiangfeng08/article/details/7642018
	ip6Str = `(([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|` +
		`(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|` +
		`(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|` +
		`(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|` +
		`(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|` +
		`(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|` +
		`(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|` +
		`(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))`

	// Matches both IP4 and IP6
	ipStr = "(" + ip4Str + ")|(" + ip6Str + ")"

	// match a domain name
	domainStr = `[a-zA-Z0-9][a-zA-Z0-9_-]{0,62}(\.[a-zA-Z0-9][a-zA-Z0-9_-]{0,62})*(\.[a-zA-Z][a-zA-Z0-9]{0,10}){1}`

	// Mach URL
	urlStr = `((https|http|ftp|rtsp|mms)?://)` + // protocols
		`(([0-9a-zA-Z]+:)?[0-9a-zA-Z_-]+@)?` + // pwd:user@
		"(" + ipStr + "|(" + domainStr + "))" + // IP or domain name
		`(:\d{1,4})?` + // port
		`(/+[a-zA-Z0-9][a-zA-Z0-9_.-]*/*)*` + // path
		`(\?([a-zA-Z0-9_-]+(=[a-zA-Z0-9_-]*)*)*)*` // query
)

func regexpCompile(str string) *regexp.Regexp {
	return regexp.MustCompile("^" + str + "$")
}

var (
	ip4      = regexpCompile(ip4Str)
	ip6      = regexpCompile(ip6Str)
	ip       = regexpCompile(ipStr)
	url      = regexpCompile(urlStr)
)

// Determine if val can match the regular expression in exp correctly.
// val can be of type []byte, []rune, string.
func isMatch(exp *regexp.Regexp, val interface{}) bool {
	switch v := val.(type) {
	case []rune:
		return exp.MatchString(string(v))
	case []byte:
		return exp.Match(v)
	case string:
		return exp.MatchString(v)
	default:
		return false
	}
}

// Verify that a value is in a standard URL format. Supports formats such as IP and domain name.
func IsURL(val interface{}) bool {
	return isMatch(url, val)
}

// Verify that a value is an IP, verifies IP4 and IP6
func IsIP(val interface{}) bool {
	return isMatch(ip, val)
}

// Verify that a value is IP6
func IsIP6(val interface{}) bool {
	return isMatch(ip6, val)
}

// Verify that a value is IP4.
func IsIP4(val interface{}) bool {
	return isMatch(ip4, val)
}

