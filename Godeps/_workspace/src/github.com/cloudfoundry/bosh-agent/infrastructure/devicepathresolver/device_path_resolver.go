package devicepathresolver

type DevicePathResolver interface {
	GetRealDevicePath(devicePath string) (realPath string, timedOut bool, err error)
}
