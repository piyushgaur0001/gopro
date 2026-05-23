package utils

import "testing"

func TestParseQuality(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		fallback int
		want     int
		wantErr  bool
	}{
		{name: "default", fallback: 80, want: 80},
		{name: "clamps low fallback", fallback: 0, want: 1},
		{name: "valid", raw: "55", fallback: 80, want: 55},
		{name: "not number", raw: "fast", fallback: 80, wantErr: true},
		{name: "too low", raw: "0", fallback: 80, wantErr: true},
		{name: "too high", raw: "101", fallback: 80, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseQuality(tt.raw, tt.fallback)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCompressionRatio(t *testing.T) {
	got := CompressionRatio(100, 25)
	if got != 0.75 {
		t.Fatalf("got %v, want 0.75", got)
	}

	if got := CompressionRatio(0, 25); got != 0 {
		t.Fatalf("got %v, want 0 for zero original size", got)
	}
}

func TestParseTargetSizeKB(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    int64
		wantSet bool
		wantErr bool
	}{
		{name: "empty", wantSet: false},
		{name: "valid", raw: "150", want: 150 * 1024, wantSet: true},
		{name: "invalid", raw: "small", wantErr: true},
		{name: "zero", raw: "0", wantErr: true},
		{name: "negative", raw: "-10", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotSet, err := ParseTargetSizeKB(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want || gotSet != tt.wantSet {
				t.Fatalf("got (%d, %v), want (%d, %v)", got, gotSet, tt.want, tt.wantSet)
			}
		})
	}
}

func TestDetectImageTypeRejectsUnknown(t *testing.T) {
	_, _, ok := DetectImageType([]byte("plain text"))
	if ok {
		t.Fatal("expected non-image data to be rejected")
	}
}
