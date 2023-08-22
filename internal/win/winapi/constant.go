package winapi

const (
	Th32csSnapprocess = 0x2
	MaxPath           = 260

	NimAdd    = 0x00000000
	NimDelete = 0x00000002

	NifMessage = 0x00000001
	NifIcon    = 0x00000002
	NifInfo    = 0x00000010

	CwUsedefault = ^0x7fffffff

	WsCaption          = 0x00c00000
	WsMaximizebox      = 0x00010000
	WsMinimizebox      = 0x00020000
	WsOverlapped       = 0x00000000
	WsSysmenu          = 0x00080000
	WsThickframe       = 0x00040000
	WsOverlappedwindow = WsOverlapped | WsCaption | WsSysmenu | WsThickframe | WsMinimizebox | WsMaximizebox

	WmDestroy       = 0x0002
	WmClose         = 0x10
	WmLbuttondblclk = 0x0203
	WmApp           = 0x8000
	WmTrayicon      = WmApp + 1

	SwpNomove = 0x0002
)
