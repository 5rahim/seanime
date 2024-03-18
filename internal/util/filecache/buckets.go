package filecache

import "time"

type Bucket struct {
	name string
	ttl  time.Duration
}

func NewBucket(name string, ttl time.Duration) Bucket {
	return Bucket{
		name: name,
		ttl:  ttl,
	}
}
