//go:build windows

package main

import (
	"context"
	"github.com/energye/systray"
	"kakaotalkadblock/internal"
	"kakaotalkadblock/winres"
	"os/exec"
)
import _ "kakaotalkadblock/winres"

const VERSION = "2.2.0"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	destroy := func() {
		cancel()
		systray.Quit()
	}
	systray.SetOnRClick(func(menu systray.IMenu) {
		menu.ShowMenu()
	})
	systray.Run(func() {
		systray.SetIcon(winres.IconData)
		systray.SetTooltip("KakaoTalkAdBlock")
		systray.AddMenuItem(VERSION, VERSION).Disable()
		systray.AddMenuItem("Check latest releases", "Check latest releases").Click(func() {
			exec.Command(
				"rundll32",
				"url.dll,FileProtocolHandler",
				"https://github.com/blurfx/KakaoTalkAdBlock/releases",
			).Start()
		})
		systray.AddSeparator()
		systray.AddMenuItem("E&xit", "Exit").Click(destroy)

		internal.Run(ctx)
	}, func() {
		cancel()
	})
}
