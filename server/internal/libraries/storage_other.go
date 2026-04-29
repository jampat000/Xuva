//go:build !windows

package libraries

func windowsDriveStorageType(root string) StorageType {
	return StorageUnknown
}
