package helpers

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/net/publicsuffix"
)

// getParameterNames returns any the names of all parameters from
// a given URL
func GetParameterNames(domain string) ([]string, error) {
	u, err := url.Parse(domain)
	if err != nil {
		return nil, err
	}

	var parameters []string
	for k := range u.Query() {
		parameters = append(parameters, k)
	}
	if len(parameters) == 0 {
		return nil, fmt.Errorf("no parameters found")
	}
	return parameters, nil
}

// getFile gets the end of a path if path does not end in "/",
// a numeric value, or a guid.
func GetFile(domain string) (string, error) {
	u, err := url.Parse(domain)
	if err != nil {
		return "", err
	}
	if len(u.Path) == 0 {
		return "", fmt.Errorf("no path found")
	}

	// if it ends in '/' there's no file to find
	if u.Path[len(u.Path)-1] == '/' {
		return "", fmt.Errorf("no file here")
	}

	pathParts := strings.Split(u.Path, "/")
	out := pathParts[len(pathParts)-1]

	// check if it's a numeric or GUID. yeet it.
	if isNumericValue(out) || isGuidValue(out) {
		return "", fmt.Errorf("numeric or GUID-style value")
	}

	return out, nil
}

// isNumericValue checks if a string is just digits and returns
// a boolean value.
func isNumericValue(s string) bool {
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		if !unicode.IsDigit(runes[i]) {
			break
		}
		if i == len(runes)-1 {
			return true
		}
	}
	return false
}

// isGuidValue checks if a string matches a GUID format and returns
// a boolean value.
func isGuidValue(s string) bool {
	re := regexp.MustCompile(`^[{]?[0-9a-fA-F]{8}-([0-9a-fA-F]{4}-){3}[0-9a-fA-F]{12}[}]?$`)
	return re.MatchString(s)
}

// getPath returns the path of a given URL to a depth of two
// directories
func GetPath(domain string) (string, error) {
	u, err := url.Parse(domain)
	if err != nil {
		return "", err
	}

	// yes, this is odd. not sure I even understand why,
	// but it works.
	pathParts := strings.Split(u.Path, "/")
	var length int
	if len(pathParts) <= 2 && len(pathParts) > 0 {
		length = len(pathParts)
	} else if len(pathParts) > 2 {
		length = 3
	} else {
		return "", fmt.Errorf("no path found")
	}

	// get rid of the root path, because for some reason
	// len(pathParts) is the same for both / and /whatever
	out := strings.Join(pathParts[:length], "/")
	if len(out) == 1 {
		return "", fmt.Errorf("it's a root path")
	}
	return out, nil
}

// getDomain attempts to retrieve the subdomain from a given
// URL. Can be a little finnicky.
func GetSubdomain(domain string) (string, error) {
	u, err := url.Parse(domain)
	if err != nil {
		return "", err
	}

	tld, managed := publicsuffix.PublicSuffix(u.Hostname())

	// brace yourself for the fugly...
	if managed {
		// I over-complicated this with regular expressions.
		// It's a simple task...
		subdomain := strings.Replace(u.Hostname(), tld, "", 1)
		dotIndex := strings.Index(subdomain, ".")
		if dotIndex > 0 {
			return subdomain[:dotIndex], nil
		}
	} else if strings.IndexByte(tld, '.') >= 0 {
		subdomain := strings.Replace(u.Hostname(), tld, "", 1)
		return subdomain[:len(subdomain)-1], nil
	} // commented this out as I don't think we're ever in a condition to hit it
	/*else {
		re := fmt.Sprintf("([a-zA-Z-]+).+[a-zA-Z-]+.%v", tld)
		r := regexp.MustCompile(re)
		matches := r.FindStringSubmatch(u.Hostname())
		if len(matches) > 0 {
			return matches[1], nil
		}
	}*/
	return "", fmt.Errorf("unable to determine subdomain")
}
