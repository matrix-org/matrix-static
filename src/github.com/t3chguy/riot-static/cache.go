// Copyright 2017 Michael Telatynski <7t3chguy@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"sync"
	"time"
)

const CacheInvalidationNumHits = 500
const CacheInvalidationTime = 30 * time.Minute

// Simple expiring store counter, initializes to expired state.

type Cached struct {
	sync.RWMutex
	cacheExpiryTime time.Time
	numberOfHits    uint
}

func (c *Cached) Hit() {
	c.Lock()
	defer c.Unlock()
	c.numberOfHits++
}

func (c *Cached) CheckExpired() bool {
	c.RLock()
	defer c.RUnlock()
	if c.numberOfHits > CacheInvalidationNumHits {
		return true
	}
	if time.Now().After(c.cacheExpiryTime) {
		return true
	}
	return false
}

func (c *Cached) Reset() {
	c.Lock()
	defer c.Unlock()
	c.cacheExpiryTime = time.Now().Add(CacheInvalidationTime)
	c.numberOfHits = 0
}
