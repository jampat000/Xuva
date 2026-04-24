package libraries

import "testing"

func TestDetectStorageType(t *testing.T) {
	tests := []struct {
		path string
		want StorageType
	}{
		{`C:\Media`, StorageLocal},
		{`X:\Movies`, StorageLocal},
		{`\\nas\media`, StorageNetwork},
		{`\\?\UNC\nas\media`, StorageNetwork},
		{`/mnt/media`, StorageMounted},
		{`/media/user/USB`, StorageMounted},
		{`/Users/me/Movies`, StorageLocal},
		{"", StorageUnknown},
	}

	for _, tt := range tests {
		if got := DetectStorageType(tt.path); got != tt.want {
			t.Fatalf("DetectStorageType(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestSetFillsStorageType(t *testing.T) {
	service := NewService()
	service.Set(Library{ID: "movies", Name: "Movies", Path: `\\nas\movies`, Kind: KindMovies})

	libraries := service.List()
	if len(libraries) != 1 {
		t.Fatalf("expected one library, got %d", len(libraries))
	}
	if libraries[0].StorageType != StorageNetwork {
		t.Fatalf("expected network storage, got %q", libraries[0].StorageType)
	}
}
