package traefik_gubernator_plugin

type RateLimitReq struct {
	// The name of the rate limit IE: 'requests_per_second', 'gets_per_minute`
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Uniquely identifies this rate limit IE: 'ip:10.2.10.7' or 'account:123445'
	UniqueKey string `json:"unique_key,omitempty" yaml:"uniqueKey,omitempty"`
	// Rate limit requests optionally specify the number of hits a request adds to the matched limit. If Hit
	// is zero, the request returns the current limit, but does not increment the hit count.
	Hits int64 `json:"hits,omitempty" yaml:"hits,omitempty"`
	// The number of requests that can occur for the duration of the rate limit
	Limit int64 `json:"limit,omitempty" yaml:"limit,omitempty"`
	// The duration of the rate limit in milliseconds
	// Second = 1000 Milliseconds
	// Minute = 60000 Milliseconds
	// Hour = 3600000 Milliseconds
	Duration int64 `json:"duration,omitempty" yaml:"duration,omitempty"`
	// The algorithm used to calculate the rate limit. The algorithm may change on
	// subsequent requests, when this occurs any previous rate limit hit counts are reset.
	Algorithm int32 `json:"algorithm,omitempty" yaml:"algorithm,omitempty"`
	// Behavior is a set of int32 flags that control the behavior of the rate limit in gubernator
	Behavior int32 `json:"behavior,omitempty" yaml:"behavior,omitempty"`
	// Maximum burst size that the limit can accept.
	Burst int64 `json:"burst,omitempty" yaml:"burst,omitempty"`
	// This is metadata that is associated with this rate limit. Peer to Peer communication will use
	// this to pass trace context to other peers. Might be useful for future clients to pass along
	// trace information to gubernator.
	Metadata map[string]string `json:"metadata,omitempty"  yaml:"metadata,omitempty"`
	// The exact time this request was created in Epoch milliseconds.  Due to
	// time drift between systems, it may be advantageous for a client to set the
	// exact time the request was created. It possible the system clock for the
	// client has drifted from the system clock where gubernator daemon is
	// running.
	//
	// The created time is used by gubernator to calculate the reset time for
	// both token and leaky algorithms. If it is not set by the client,
	// gubernator will set the created time when it receives the rate limit
	// request.
	CreatedAt *int64 `json:"created_at,omitempty" yaml:"createdAt,omitempty"`
}
type RateLimitResp struct {
	// The status of the rate limit.
	Status string `protobuf:"varint,1,opt,name=status,proto3,enum=pb.gubernator.Status" json:"status,omitempty"`
	// The currently configured request limit (Identical to [[RateLimitReq.limit]]).
	Limit string `protobuf:"varint,2,opt,name=limit,proto3" json:"limit,omitempty"`
	// This is the number of requests remaining before the rate limit is hit but after subtracting the hits from the current request
	Remaining string `protobuf:"varint,3,opt,name=remaining,proto3" json:"remaining,omitempty"`
	// This is the time when the rate limit span will be reset, provided as a unix timestamp in milliseconds.
	ResetTime string `protobuf:"varint,4,opt,name=reset_time,json=resetTime,proto3" json:"reset_time,omitempty"`
	// Contains the error; If set all other values should be ignored
	Error string `protobuf:"bytes,5,opt,name=error,proto3" json:"error,omitempty"`
	// This is additional metadata that a client might find useful. (IE: Additional headers, coordinator ownership, etc..)
	Metadata map[string]string `protobuf:"bytes,6,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

type GetRateLimitsReq struct {
	Requests []*RateLimitReq `protobuf:"bytes,1,rep,name=requests,proto3" json:"requests,omitempty"`
}
type GetRateLimitsResp struct {
	Responses []*RateLimitResp `protobuf:"bytes,1,rep,name=responses,proto3" json:"responses,omitempty"`
}
