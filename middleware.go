package traefik_gubernator_plugin

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// Config the plugin configuration.
type Config struct {
	Remote  string            `json:"remote" yaml:"remote"`
	Headers map[string]string `json:"headers" yaml:"headers"`

	Limits []RateLimitReq `json:"limits" yaml:"limits"`
}

// GubernatorPlugin
type GubernatorPlugin struct {
	noop bool
	next http.Handler

	client *gubernatorClient
	limits []*RequestWithTemplate
}

type RequestWithTemplate struct {
	config   RateLimitReq
	template *Template
}

// New creates a new GubernatorPlugin plugin.
func New(_ context.Context, next http.Handler, config *Config, _ string) (http.Handler, error) {
	if len(config.Remote) == 0 {
		return nil, fmt.Errorf("remote cannot be empty")
	}

	g := &GubernatorPlugin{next: next}
	if config.Remote == "noop" {
		g.noop = true
		return g, nil
	}

	c, err := newClient(config)
	if err != nil {
		return nil, err
	}
	g.client = c

	for _, v := range config.Limits {
		tmpl, err := NewTemplate(v.UniqueKey, "{", "}")
		if err != nil {
			return nil, err
		}
		if v.UniqueKey == "" {
			return nil, fmt.Errorf("unique key cannot be empty")
		}
		g.limits = append(g.limits, &RequestWithTemplate{
			template: tmpl,
			config:   v,
		})
	}
	return g, nil
}

func readUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}

func (a *GubernatorPlugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if a.noop {
		a.next.ServeHTTP(w, r)
		return
	}
	xs := make([]*RateLimitReq, 0)
	for _, rl := range a.limits {
		// make key for the individual rate limiter in this zone
		computedKey := rl.template.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
			splt := strings.SplitN(tag, ".", 2)
			switch splt[0] {
			case "header":
				if len(splt) == 1 {
					return w.Write([]byte(""))
				}
				return w.Write([]byte(r.Header.Get(splt[1])))
			case "client_ip":
				val := readUserIP(r)
				return w.Write([]byte(val))
			default:
				return w.Write([]byte("?unknown_tag?"))
			}
		})

		// clone the request
		clonedReq := &RateLimitReq{
			Name:      rl.config.Name,
			Algorithm: rl.config.Algorithm,
			UniqueKey: computedKey,
			Hits:      rl.config.Hits,
			Limit:     rl.config.Limit,
			Duration:  rl.config.Duration,
			Behavior:  rl.config.Behavior,
			Burst:     rl.config.Burst,
			Metadata:  rl.config.Metadata,
		}
		// assign the key
		xs = append(xs, clonedReq)
	}
	if len(xs) == 0 {
		a.next.ServeHTTP(w, r)
		return
	}
	ctx, cn := context.WithTimeout(context.Background(), 1*time.Second)
	ratelimitResponse, err := a.client.GetRateLimits(ctx, &GetRateLimitsReq{
		Requests: xs,
	})
	cn()
	if err != nil {
		os.Stdout.Write([]byte(fmt.Sprintf("rl request failed:%s \n", err)))
		http.Error(w, "failed to contact rate limit", http.StatusTooManyRequests)
		return
	}
	for idx, v := range ratelimitResponse.Responses {
		// TODO: allow this to be configured
		// early exit on any error
		if len(v.Error) > 0 {
			os.Stdout.Write([]byte(fmt.Sprintf("rl request errored: %s\n", v.Error)))
			http.Error(w, "try again later", http.StatusTooManyRequests)
			return
		}
		if idx < len(xs) {
			k := xs[idx].UniqueKey
			w.Header().Add("Ratelimit-Request-Key", k)
		} else {
			w.Header().Add("Ratelimit-Request-Key", "!ERROR!")
		}
		// Note the use of add. this should be within the standard.
		w.Header().Add("Ratelimit-Limit", v.Limit)
		w.Header().Add("Ratelimit-Remaining", v.Remaining)

		resetTimeInt, _ := strconv.Atoi(v.ResetTime)
		resetTime := resetTimeInt/1000 - int(time.Now().UnixMilli()/1000)
		if resetTime < 0 {
			resetTime = 0
		}
		w.Header().Add("Ratelimit-Reset", strconv.Itoa(int(resetTime)))
		remainingInt, _ := strconv.Atoi(v.Remaining)

		if len(v.Error) > 0 || remainingInt < 1 {
			// early exit on first error
			http.Error(w, "no remaining?", http.StatusTooManyRequests)
			return
		}
	}
	a.next.ServeHTTP(w, r)
}
