package htmlword

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

type cachedResource struct {
	data []byte
	mime string
	err  error
	done chan struct{}
}
type resourceLoader struct {
	resolver ResourceResolver
	limits   Limits
	mu       sync.Mutex
	cache    map[string]*cachedResource
	total    int64
	limitErr error
	slots    chan struct{}
}

func newResourceLoader(r ResourceResolver, l Limits) *resourceLoader {
	return &resourceLoader{resolver: r, limits: l, cache: map[string]*cachedResource{}, slots: make(chan struct{}, 4)}
}
func (l *resourceLoader) fatal() error { l.mu.Lock(); defer l.mu.Unlock(); return l.limitErr }
func (l *resourceLoader) load(ctx context.Context, ref string) ([]byte, string, error) {
	if err := ctx.Err(); err != nil {
		return nil, "", err
	}
	l.mu.Lock()
	if entry, ok := l.cache[ref]; ok {
		l.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, "", ctx.Err()
		case <-entry.done:
			return entry.data, entry.mime, entry.err
		}
	}
	// Bound failed URLs too: hostile CSS cannot create an unbounded cache.
	if len(l.cache) >= l.limits.DOMNodes {
		l.limitErr = ErrLimit
		l.mu.Unlock()
		return nil, "", ErrLimit
	}
	entry := &cachedResource{done: make(chan struct{})}
	l.cache[ref] = entry
	l.mu.Unlock()
	var data []byte
	var contentType string
	var err error
	if l.resolver == nil {
		err = fmt.Errorf("%w: external resources are disabled", ErrResource)
	} else {
		select {
		case l.slots <- struct{}{}:
			defer func() { <-l.slots }()
		case <-ctx.Done():
			err = ctx.Err()
		}
		var resource Resource
		if err == nil {
			resource, err = l.resolver.Resolve(ctx, ref)
		}
		if resource.Body != nil {
			defer resource.Body.Close()
		}
		if err == nil && resource.Body == nil {
			err = fmt.Errorf("%w: resolver returned a nil body", ErrResource)
		}
		if err == nil {
			data, err = io.ReadAll(io.LimitReader(resource.Body, l.limits.ResourceBytes+1))
			contentType = resource.ContentType
			if int64(len(data)) > l.limits.ResourceBytes {
				err = fmt.Errorf("%w: resource bytes", ErrLimit)
				data = nil
			}
		}
		if err == nil && ctx.Err() != nil {
			err = ctx.Err()
		}
		if contentType == "" {
			contentType = http.DetectContentType(data)
		}
		if strings.ContainsAny(contentType, "\r\n") {
			err = fmt.Errorf("%w: invalid content type", ErrResource)
		}
	}
	l.mu.Lock()
	if err == nil {
		if int64(len(data)) > l.limits.TotalResourceBytes-l.total {
			err = fmt.Errorf("%w: total resource bytes", ErrLimit)
			data = nil
		} else {
			l.total += int64(len(data))
		}
	}
	if errors.Is(err, ErrLimit) {
		l.limitErr = err
	}
	entry.data, entry.mime, entry.err = data, contentType, err
	close(entry.done)
	l.mu.Unlock()
	return data, contentType, err
}

// FSResolver maps URL paths to an explicitly provided filesystem. Paths are
// checked with fs.ValidPath; it never chooses a host filesystem on its own.
// Query parameters are ignored. The caller controls filesystem permissions:
// os.DirFS, for example, follows symlinks and is not an OS-level sandbox.
func FSResolver(fsys fs.FS) ResourceResolver {
	return ResourceResolverFunc(func(ctx context.Context, ref string) (Resource, error) {
		if err := ctx.Err(); err != nil {
			return Resource{}, err
		}
		if fsys == nil {
			return Resource{}, fmt.Errorf("%w: nil filesystem", ErrResource)
		}
		u, err := url.Parse(ref)
		if err != nil || u.User != nil || (u.Scheme != "http" && u.Scheme != "https") {
			return Resource{}, fmt.Errorf("%w: invalid asset URL", ErrResource)
		}
		name := strings.TrimPrefix(u.Path, "/")
		if !fs.ValidPath(name) || strings.Contains(name, "\\") {
			return Resource{}, fmt.Errorf("%w: invalid asset path", ErrResource)
		}
		body, err := fsys.Open(name)
		if err != nil {
			return Resource{}, err
		}
		return Resource{Body: body, ContentType: mime.TypeByExtension(path.Ext(name))}, nil
	})
}

// HTTPResolver explicitly opts into public HTTP(S) resource access. Empty
// AllowedHosts permits any public host; a nonempty list uses exact hostnames.
// Redirects are rechecked. Proxy environment variables are ignored. DNS results
// are checked and the connection is pinned to the checked address. Private,
// loopback, link-local, carrier-grade NAT and documentation networks are denied.
// For authenticated/private resources, supply your own ResourceResolver instead.
type HTTPResolver struct {
	AllowedHosts []string
	Timeout      time.Duration
}

func (r HTTPResolver) Resolve(ctx context.Context, ref string) (Resource, error) {
	validate := func(raw string) error {
		u, e := url.Parse(raw)
		if e != nil || u.Hostname() == "" || u.User != nil || (u.Scheme != "https" && u.Scheme != "http") {
			return fmt.Errorf("%w: invalid HTTP URL", ErrResource)
		}
		if port := u.Port(); port != "" && !(u.Scheme == "http" && port == "80") && !(u.Scheme == "https" && port == "443") {
			return fmt.Errorf("%w: only standard HTTP(S) ports are allowed", ErrResource)
		}
		if len(r.AllowedHosts) > 0 {
			allowed := false
			for _, host := range r.AllowedHosts {
				allowed = allowed || strings.EqualFold(strings.TrimSuffix(u.Hostname(), "."), strings.TrimSuffix(host, "."))
			}
			if !allowed {
				return fmt.Errorf("%w: host is not allowed", ErrResource)
			}
		}
		return nil
	}
	if err := validate(ref); err != nil {
		return Resource{}, err
	}
	transport := &http.Transport{Proxy: nil, DialContext: func(job context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		ips, err := net.DefaultResolver.LookupNetIP(job, "ip", host)
		if err != nil || len(ips) == 0 {
			return nil, fmt.Errorf("%w: hostname resolution failed", ErrResource)
		}
		for _, ip := range ips {
			if !publicIP(ip) {
				return nil, fmt.Errorf("%w: non-public destination", ErrResource)
			}
		}
		return (&net.Dialer{Timeout: 10 * time.Second}).DialContext(job, network, net.JoinHostPort(ips[0].String(), port))
	}}
	timeout := r.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	client := &http.Client{Transport: transport, Timeout: timeout, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) > 4 {
			return fmt.Errorf("%w: too many redirects", ErrResource)
		}
		return validate(req.URL.String())
	}}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, ref, nil)
	if err != nil {
		transport.CloseIdleConnections()
		return Resource{}, err
	}
	response, err := client.Do(request)
	if err != nil {
		transport.CloseIdleConnections()
		return Resource{}, err
	}
	if response.StatusCode != http.StatusOK {
		response.Body.Close()
		transport.CloseIdleConnections()
		return Resource{}, fmt.Errorf("%w: HTTP status %d", ErrResource, response.StatusCode)
	}
	return Resource{Body: &httpResourceBody{ReadCloser: response.Body, transport: transport}, ContentType: response.Header.Get("Content-Type")}, nil
}

type httpResourceBody struct {
	io.ReadCloser
	transport *http.Transport
}

func (b *httpResourceBody) Close() error {
	err := b.ReadCloser.Close()
	b.transport.CloseIdleConnections()
	return err
}
func publicIP(address netip.Addr) bool {
	address = address.Unmap()
	if !address.IsValid() || !address.IsGlobalUnicast() || address.IsPrivate() || address.IsLoopback() || address.IsLinkLocalUnicast() {
		return false
	}
	for _, prefix := range []string{"0.0.0.0/8", "100.64.0.0/10", "192.0.0.0/24", "192.0.2.0/24", "198.18.0.0/15", "198.51.100.0/24", "203.0.113.0/24", "240.0.0.0/4", "2001:db8::/32"} {
		if netip.MustParsePrefix(prefix).Contains(address) {
			return false
		}
	}
	return true
}
