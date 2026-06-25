package avr

import "testing"

func TestParseIntelHex(t *testing.T) {
	in := []byte(":100000000C9462000C948A000C948A000C948A0070\n:00000001FF\n")
	out, err := parseIntelHex(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 16 || out[0] != 0x0C || out[1] != 0x94 || out[2] != 0x62 {
		t.Fatalf("bad decode: %X", out)
	}

	// embedded firmware must decode and start with a reset vector jump (0x0C 0x94)
	fw, err := parseIntelHex(firmwareHex)
	if err != nil {
		t.Fatal(err)
	}
	if len(fw) == 0 || fw[0] != 0x0C || fw[1] != 0x94 {
		t.Fatalf("firmware decode looks wrong: len=%d head=%X", len(fw), fw[:2])
	}

	// flipped checksum byte must fail
	if _, err := parseIntelHex([]byte(":100000000C9462000C948A000C948A000C948A0071\n")); err == nil {
		t.Fatal("expected checksum error")
	}
}
