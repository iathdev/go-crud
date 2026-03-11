package common

import (
	"net"
	"net/http"
	"strings"
)

// ipHeaders is the ordered list of headers to check for the real client IP,
// matching the Laravel Helper::getIpAddress() behavior.
var ipHeaders = []string{
	"X-Client-IP",
	"X-Forwarded-For",
	"X-Forwarded",
	"X-Cluster-Client-IP",
	"Forwarded-For",
	"Forwarded",
}

// ResolveClientIP extracts the real client IP from request headers.
// It checks proxy headers in priority order, skips private/reserved IPs,
// and falls back to RemoteAddr.
func ResolveClientIP(r *http.Request) string {
	for _, header := range ipHeaders {
		value := r.Header.Get(header)
		if value == "" {
			continue
		}
		for _, raw := range strings.Split(value, ",") {
			ip := strings.TrimSpace(raw)
			if ip == "" {
				continue
			}
			parsed := net.ParseIP(ip)
			if parsed != nil && !isPrivateIP(parsed) {
				return ip
			}
		}
	}

	// Fallback: RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func isPrivateIP(ip net.IP) bool {
	privateRanges := []net.IPNet{
		{IP: net.IPv4(10, 0, 0, 0), Mask: net.CIDRMask(8, 32)},
		{IP: net.IPv4(172, 16, 0, 0), Mask: net.CIDRMask(12, 32)},
		{IP: net.IPv4(192, 168, 0, 0), Mask: net.CIDRMask(16, 32)},
		{IP: net.IPv4(127, 0, 0, 0), Mask: net.CIDRMask(8, 32)},
	}
	for _, r := range privateRanges {
		if r.Contains(ip) {
			return true
		}
	}
	return ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsLoopback()
}
