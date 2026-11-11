package winapi

const (
	WmClose = 0x10

	SwpNosize       = 0x0001
	SwpNomove       = 0x0002
	SwpNozorder     = 0x0004
	SwpNoactivate   = 0x0010
	SwpFramechanged = 0x0020

	DwmFrameWindowStyles = 0x00C40000
	WindowStateStyles    = 0x21000000

	WmUser     = 0x0400
	WmApp      = 0x8000
	PmNoremove = 0x0000

	WineventOutofcontext      = 0x0000
	EventObjectCreate         = 0x8000
	EventObjectHide           = 0x8003
	EventObjectLocationchange = 0x800B
	EventObjectNamechange     = 0x800C
	ObjidWindow               = 0
)
