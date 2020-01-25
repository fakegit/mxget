package sreq

import (
	neturl "net/url"
	"path"
)

// SetDefaultHost is a before request hook set client level Host header value,
// can be overwrite by request level option.
func SetDefaultHost(host string) BeforeRequestHook {
	return func(req *Request) error {
		req.headers.SetDefault("Host", host)
		return nil
	}
}

// SetDefaultHeaders is a before request hook to set client level headers,
// can be overwrite by request level option.
func SetDefaultHeaders(headers Headers) BeforeRequestHook {
	return func(req *Request) error {
		req.headers.Merge(headers)
		return nil
	}
}

// SetDefaultContentType is a before request hook to set client level Content-Type header value,
// can be overwrite by request level option.
func SetDefaultContentType(contentType string) BeforeRequestHook {
	return func(req *Request) error {
		req.headers.SetDefault("Content-Type", contentType)
		return nil
	}
}

// SetDefaultUserAgent is a before request hook to set client level User-Agent header value,
// can be overwrite by request level option.
func SetDefaultUserAgent(userAgent string) BeforeRequestHook {
	return func(req *Request) error {
		req.headers.SetDefault("User-Agent", userAgent)
		return nil
	}
}

// SetDefaultOrigin is a before request hook to set client level Origin header value,
// can be overwrite by request level option.
func SetDefaultOrigin(origin string) BeforeRequestHook {
	return func(req *Request) error {
		req.headers.SetDefault("Origin", origin)
		return nil
	}
}

// SetDefaultReferer is a before request hook to set client level Referer header value,
// can be overwrite by request level option.
func SetDefaultReferer(referer string) BeforeRequestHook {
	return func(req *Request) error {
		req.headers.SetDefault("Referer", referer)
		return nil
	}
}

// SetDefaultQuery is a before request hook to set client level query parameters,
// can be overwrite by request level option.
func SetDefaultQuery(query Params) BeforeRequestHook {
	return func(req *Request) error {
		req.query.Merge(query)
		return nil
	}
}

// SetDefaultForm is a before request hook to set client level form payload,
// can be overwrite by request level option.
func SetDefaultForm(form Form) BeforeRequestHook {
	return func(req *Request) error {
		req.form.Merge(form)
		return nil
	}
}

// SetDefaultCookies is a before request hook to set client level cookies,
// can be overwrite by request level option.
func SetDefaultCookies(cookies Cookies) BeforeRequestHook {
	return func(req *Request) error {
		req.cookies.Merge(cookies)
		return nil
	}
}

// SetDefaultBasicAuth is a before request hook to set client level basic authentication,
// can be overwrite by request level option.
func SetDefaultBasicAuth(username string, password string) BeforeRequestHook {
	return func(req *Request) error {
		req.headers.SetDefault("Authorization", "Basic "+basicAuth(username, password))
		return nil
	}
}

// SetDefaultBearerToken is a before request hook to set client level bearer token,
// can be overwrite by request level option.
func SetDefaultBearerToken(token string) BeforeRequestHook {
	return func(req *Request) error {
		req.headers.SetDefault("Authorization", "Bearer "+token)
		return nil
	}
}

// SetDefaultRetry is a before request hook to set client level retry policy,
// can be overwrite by request level option.
func SetDefaultRetry(maxAttempts int, backoff Backoff, triggers ...func(resp *Response) bool) BeforeRequestHook {
	return func(req *Request) error {
		if req.retry == nil {
			req.SetRetry(maxAttempts, backoff, triggers...)
		}
		return nil
	}
}

// SetReverseProxy is a before request hook to set reverse proxy for HTTP requests.
func SetReverseProxy(target string, publicPaths ...string) BeforeRequestHook {
	return func(req *Request) error {
		u, err := neturl.Parse(target)
		if err != nil {
			return err
		}

		req.URL.Scheme = u.Scheme
		req.URL.Host = u.Host
		req.Host = u.Host
		req.SetOrigin(target)

		if len(publicPaths) > 0 {
			publicPath := path.Join(publicPaths...)
			req.URL.Path = path.Join(publicPath, req.URL.Path)
		}
		return nil
	}
}
