package password

import (
	"strings"
	"testing"
)

func TestHashStringGeneratesVerifiableUniqueHashes(t *testing.T) {
	const password = "baby baby do baby do biruleibe leibe?"

	first, err := Hash(password)
	if err != nil {
		t.Fatalf("hash first password: %v", err)
	}
	second, err := Hash(password)
	if err != nil {
		t.Fatalf("hash second password: %v", err)
	}

	if first == second {
		t.Fatal("equal passwords produced identical hashes")
	}
	if !strings.HasPrefix(first, "$argon2id$v=19$m=65536,t=1,p=4$") {
		t.Fatalf("hash does not use expected PHC format: %q", first)
	}
	if !Verify(password, first) {
		t.Fatal("expected password to verify")
	}
	if Verify("wrong password", first) {
		t.Fatal("wrong password verified")
	}
}

func TestVerifyPasswordRejectsMalformedHash(t *testing.T) {
	if Verify("password", "$argon2id$v=19$m=1,t=1,p=1$bad$bad") {
		t.Fatal("malformed hash verified")
	}
}
