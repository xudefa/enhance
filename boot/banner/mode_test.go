package banner

import "testing"

func TestModeString_ExtendedCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mode BannerMode
		want string
	}{
		{BannerModeConsole, "console"},
		{BannerModeLog, "log"},
		{BannerModeOff, "off"},
		{BannerMode(99), "unknown"},
		{BannerMode(-1), "unknown"},
		{BannerMode(3), "unknown"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			if got := tt.mode.String(); got != tt.want {
				t.Errorf("BannerMode(%d).String() = %q, want %q", tt.mode, got, tt.want)
			}
		})
	}
}
