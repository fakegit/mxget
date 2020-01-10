package sreq

// SetHostDefault is a request interceptor that allows you set client level host,
// can be overwrite by request level option.
func SetHostDefault(host string) RequestInterceptor {
	return func(req *Request) error {
		if req.Host == "" {
			req.Host = host
		}
		return nil
	}
}

// SetHeadersDefault is a request interceptor that allows you to set client level headers,
// can be overwrite by request level option.
func SetHeadersDefault(headers Headers) RequestInterceptor {
	return func(req *Request) error {
		req.Headers.Merge(headers)
		return nil
	}
}

// SetContentTypeDefault is a request interceptor that allows you to set client level Content-Type header value,
// can be overwrite by request level option.
func SetContentTypeDefault(contentType string) RequestInterceptor {
	return func(req *Request) error {
		req.Headers.SetDefault("Content-Type", contentType)
		return nil
	}
}

// SetUserAgentDefault is a request interceptor that allows you to set client level User-Agent header value,
// can be overwrite by request level option.
func SetUserAgentDefault(userAgent string) RequestInterceptor {
	return func(req *Request) error {
		req.Headers.SetDefault("User-Agent", userAgent)
		return nil
	}
}

// SetRefererDefault is a request interceptor that allows you to set client level Referer header value,
// can be overwrite by request level option.
func SetRefererDefault(referer string) RequestInterceptor {
	return func(req *Request) error {
		req.Headers.SetDefault("Referer", referer)
		return nil
	}
}

// SetQueryDefault is a request interceptor that allows you to set client level query parameters,
// can be overwrite by request level option.
func SetQueryDefault(params Params) RequestInterceptor {
	return func(req *Request) error {
		req.Params.Merge(params)
		return nil
	}
}

// SetFormDefault is a request interceptor that allows you to set client level form payload,
// can be overwrite by request level option.
func SetFormDefault(form Form) RequestInterceptor {
	return func(req *Request) error {
		req.Form.Merge(form)
		return nil
	}
}

// SetCookiesDefault is a request interceptor that allows you to set client level cookies,
// can be overwrite by request level option.
func SetCookiesDefault(cookies Cookies) RequestInterceptor {
	return func(req *Request) error {
		req.Cookies.Merge(cookies)
		return nil
	}
}

// SetBasicAuthDefault is a request interceptor that allows you to set client level basic authentication,
// can be overwrite by request level option.
func SetBasicAuthDefault(username string, password string) RequestInterceptor {
	return func(req *Request) error {
		req.Headers.SetDefault("Authorization", "Basic "+basicAuth(username, password))
		return nil
	}
}

// SetBearerTokenDefault is a request interceptor that allows you to set client level bearer token,
// can be overwrite by request level option.
func SetBearerTokenDefault(token string) RequestInterceptor {
	return func(req *Request) error {
		req.Headers.SetDefault("Authorization", "Bearer "+token)
		return nil
	}
}
