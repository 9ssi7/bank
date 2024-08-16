package state

import "context"

const DeviceIDKey ContextKey = "device_id"

// SetDeviceId sets the device id in the context
func SetDeviceId(ctx context.Context, deviceId string) context.Context {
	return context.WithValue(ctx, DeviceIDKey, deviceId)
}

// GetDeviceId gets the device id from the context
func GetDeviceId(ctx context.Context) string {
	if deviceId, ok := ctx.Value(DeviceIDKey).(string); ok {
		return deviceId
	}
	return ""
}
