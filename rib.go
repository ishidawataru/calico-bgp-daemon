// Copyright (C) 2017 Nippon Telegraph and Telephone Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"fmt"
	"log"
	"net"

	"github.com/armon/go-radix"
)

func ipToRadixKey(b []byte, max uint8) string {
	var buffer bytes.Buffer
	for i := 0; i < len(b) && i < int(max); i++ {
		buffer.WriteString(fmt.Sprintf("%08b", b[i]))
	}
	return buffer.String()[:max]
}

func cidrToRadixKey(cidr string) (bool, string, error) {
	ip, n, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, "", err
	}
	ones, _ := n.Mask.Size()
	return ip.To4() != nil, ipToRadixKey(n.IP, uint8(ones)), nil
}

type RIB struct {
	v4 *radix.Tree
	v6 *radix.Tree
}

func NewRIB() *RIB {
	return &RIB{
		v4: radix.New(),
		v6: radix.New(),
	}
}

func (r *RIB) Get(prefix string) (interface{}, error) {
	v4, k, err := cidrToRadixKey(prefix)
	if err != nil {
		return nil, err
	}
	var value interface{}
	if v4 {
		_, value, _ = r.v4.LongestPrefix(k)
	} else {
		_, value, _ = r.v6.LongestPrefix(k)
	}
	return value, nil
}

func (r *RIB) Add(prefix string, value interface{}) error {
	v4, k, err := cidrToRadixKey(prefix)
	if err != nil {
		return err
	}
	log.Printf("rib add: %s %s v4: %t", prefix, value, v4)
	if v4 {
		r.v4.Insert(k, value)
	} else {
		r.v6.Insert(k, value)
	}
	return nil
}

func (r *RIB) Delete(prefix string) error {
	v4, k, err := cidrToRadixKey(prefix)
	if err != nil {
		return err
	}
	log.Printf("rib del: %s", prefix)
	if v4 {
		r.v4.Delete(k)
	} else {
		r.v6.Delete(k)
	}
	return nil
}
