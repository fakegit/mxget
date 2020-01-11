package sreq

// SetHostDefault is a before request hook that allows you set client level host,
// can be overwrite by request level option.
func SetHostDefault(host string) BeforeRequestHook {
	return func(req *Request) error {
		if req.Host == "" {
			req.Host = host
		}
		return nil
	}
}

// SetHeadersDefault is a before request hook that allows you to set client level headers,
// can be overwrite by request level option.
func SetHeadersDefault(headers Headers) BeforeRequestHook {
	return func(req *Request) error {
		req.Headers.Merge(headers)
		return nil
	}
}

// SetContentTypeDefault is a before request hook that allows you to set client level Content-Type header value,
// can be overwrite by request level option.
func SetContentTypeDefault(contentType string) BeforeRequestHook {
	return func(req *Request) error {
		req.Headers.SetDefault("Content-Type", contentType)
		return nil
	}
}

// SetUserAgentDefault is a before request hook that allows you to set client level User-Agent header value,
// can be overwrite by request level option.
func SetUserAgentDefault(userAgent string) BeforeRequestHook {
	return func(req *Request) error {
		req.Headers.SetDefault("User-Agent", userAgent)
		return nil
	}
}

// SetRefererDefault is a before request hook that allows you to set client level Referer header value,
// can be overwrite by request level option.
func SetRefererDefault(referer string) BeforeRequestHook {
	return func(req *Request) error {
		req.Headers.SetDefault("Referer", referer)
		return nil
	}
}

// SetQueryDefault is a before request hook that allows you to set client level query parameters,
// can be overwrite by request level option.
func SetQueryDefault(params Params) BeforeRequestHook {
	return func(req *Request) error {
		req.Params.Merge(params)
		return nil
	}
}

// SetFormDefault is a before request hook that allows you to set client level form payload,
// can be overwrite by request level option.
func SetFormDefault(form Form) BeforeRequestHook {
	return func(req *Request) error {
		req.Form.Merge(form)
		return nil
	}
}

// SetCookiesDefault is a before request hook that allows you to set client level cookies,
// can be overwrite by request level option.
func SetCookiesDefault(cookies Cookies) BeforeRequestHook {
	return func(req *Request) error {
		req.Cookies.Merge(cookies)
		return nil
	}
}

// SetBasicAuthDefault is a before request hook that allows you to set client level basic authentication,
// can be overwrite by request level option.
func SetBasicAuthDefault(username string, password string) BeforeRequestHook {
	return func(req *Request) error {
		req.Headers.SetDefault("Authorization", "Basic "+basicAuth(username, password))
		return nil
	}
}

// SetBearerTokenDefault is a before request hook that allows you to set client level bearer token,
// can be overwrite by request level option.
func SetBearerTokenDefault(token string) BeforeRequestHook {
	return func(req *Request) error {
		req.Headers.SetDefault("Authorization", "Bearer "+token)
		return nil
	}
}
