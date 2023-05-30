package msg

import (
	"testing"
	"time"
)

func TestParseMsgTags(t *testing.T) {
	tags := newMsgTags()
	parseMsgTags([]string{" ds = 2 ", "ttl=48h"}, tags)
	if tags.DS != 2 || tags.TTL != 48*time.Hour {
		t.Errorf("parse failed: tags=%d, ttl=%s\n", tags.DS, tags.TTL)
	}
	tags = newMsgTags()
	parseMsgTags([]string{" ds = 256 ", "ttl=73h"}, tags)
	if tags.DS != 0 || tags.TTL != MaxMsgTTL {
		t.Errorf("parse failed: tags=%d, ttl=%s\n", tags.DS, tags.TTL)
	}
	tags = newMsgTags()
	parseMsgTags(nil, tags)
	if tags.DS != 0 || tags.TTL != MaxMsgTTL {
		t.Errorf("parse failed: tags=%d, ttl=%s\n", tags.DS, tags.TTL)
	}
}
